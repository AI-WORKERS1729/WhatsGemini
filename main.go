package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"

	"github.com/mdp/qrterminal/v3"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"

	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/protobuf/proto"
)

type SendRequest struct {
	JID     string `json:"jid"`
	Message string `json:"message"`
}

type ReceivedMessage struct {
	From    string `json:"from"`
	Message string `json:"message"`
}

var (
	client           *whatsmeow.Client
	receivedMessages []ReceivedMessage
	mu               sync.Mutex
	debugMode        bool
)

func main() {
	// Parse the --debug flag
	flag.BoolVar(&debugMode, "debug", false, "Enable debug output")
	flag.Parse()

	fmt.Println("✅ Code started...")

	// Use loggers only if debug mode is enabled
	var dbLog, clientLog waLog.Logger
	if debugMode {
		dbLog = waLog.Stdout("DB", "INFO", true)
		clientLog = waLog.Stdout("Client", "INFO", true)
	}

	container, err := sqlstore.New("sqlite3", "file:whatsapp.db?_foreign_keys=on", dbLog)
	if err != nil {
		panic(err)
	}

	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		panic(err)
	}

	client = whatsmeow.NewClient(deviceStore, clientLog)

	// Message event handler
	client.AddEventHandler(func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Message:
			// Ignore messages from ourselves
			if v.Info.MessageSource.IsFromMe {
				return
			}

			var text string

			switch {
			case v.Message.GetConversation() != "":
				text = v.Message.GetConversation()
			case v.Message.GetExtendedTextMessage() != nil:
				text = v.Message.GetExtendedTextMessage().GetText()
			case v.Message.GetImageMessage() != nil:
				text = "[Image message]"
			case v.Message.GetVideoMessage() != nil:
				text = "[Video message]"
			case v.Message.GetDocumentMessage() != nil:
				text = "[Document message]"
			case v.Message.GetButtonsMessage() != nil:
				text = "[Buttons message]"
			case v.Message.GetListMessage() != nil:
				text = "[List message]"
			default:
				text = "[Unsupported or empty message type]"
			}

			if debugMode {
				fmt.Printf("📩 Message from %s: %s\n", v.Info.Sender.User, text)
			}

			mu.Lock()
			receivedMessages = append(receivedMessages, ReceivedMessage{
				From:    v.Info.Sender.User,
				Message: text,
			})
			mu.Unlock()
		}
	})

	// QR code login if not connected
	if client.Store.ID == nil {
		qrChan, _ := client.GetQRChannel(context.Background())
		go func() {
			for evt := range qrChan {
				if evt.Event == "code" {
					qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
					fmt.Println("Scan the QR code above with WhatsApp.")
				} else if debugMode {
					fmt.Println("QR Event:", evt.Event)
				}
			}
		}()

		err = client.Connect()
		if err != nil {
			panic(err)
		}
	} else {
		err = client.Connect()
		if err != nil {
			panic(err)
		}
		fmt.Println("✅ Reconnected using saved session.")
	}

	// Set up HTTP server
	http.HandleFunc("/send", sendHandler)
	http.HandleFunc("/messages", messagesHandler)

	go func() {
		fmt.Println("🌐 Server started on http://localhost:8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			panic(err)
		}
	}()

	// Wait for Ctrl+C
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	client.Disconnect()
	fmt.Println("👋 Clean shutdown.")
}

func sendHandler(w http.ResponseWriter, r *http.Request) {
	var req SendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	jid := types.NewJID(req.JID, "s.whatsapp.net")
	_, err := client.SendMessage(context.Background(), jid, &waProto.Message{
		Conversation: proto.String(req.Message),
	})

	if err != nil {
		http.Error(w, "Failed to send message", http.StatusInternalServerError)
		return
	}

	if debugMode {
		fmt.Printf("✅ Sent message to %s: %s\n", req.JID, req.Message)
	}

	w.Write([]byte("✅ Message sent"))
}

func messagesHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(receivedMessages); err != nil {
		http.Error(w, "Failed to fetch messages", http.StatusInternalServerError)
		return
	}

	receivedMessages = []ReceivedMessage{}
}
