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
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
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

type MediaSendRequest struct {
	JID       string `json:"jid"`
	FilePath  string `json:"file_path"`
	MediaType string `json:"media_type"`
	Caption   string `json:"caption,omitempty"`
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
	flag.BoolVar(&debugMode, "debug", false, "Enable debug output")
	flag.Parse()

	fmt.Println("✅ Code started...")

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

	client.AddEventHandler(func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Message:
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
			case v.Message.GetAudioMessage() != nil:
				text = "[Audio message]"
			case v.Message.GetStickerMessage() != nil:
				text = "[Sticker message]"
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

	http.HandleFunc("/send", sendHandler)
	http.HandleFunc("/messages", messagesHandler)
	http.HandleFunc("/sendMedia", sendMediaHandler)

	go func() {
		fmt.Println("🌐 Server started on http://localhost:8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			panic(err)
		}
	}()

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

func sendMediaHandler(w http.ResponseWriter, r *http.Request) {
	var req MediaSendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	fileData, err := os.ReadFile(req.FilePath)
	if err != nil {
		http.Error(w, "Failed to read media file", http.StatusInternalServerError)
		return
	}

	// Upload media to WhatsApp
	uploadedMedia, err := client.Upload(context.Background(), fileData, whatsmeow.MediaImage)
	if err != nil {
		http.Error(w, "Failed to upload media", http.StatusInternalServerError)
		if debugMode {
			fmt.Printf("❌ Media upload failed: %v\n", err)
		}
		return
	}

	jid := types.NewJID(req.JID, "s.whatsapp.net")
	var msg *waProto.Message

	// Prepare the message depending on the media type
	switch req.MediaType {
	case "photo":
		msg = &waProto.Message{
			ImageMessage: &waProto.ImageMessage{
				Caption:       proto.String(req.Caption),
				Mimetype:      proto.String("image/png"),
				URL:           proto.String(uploadedMedia.URL),
				DirectPath:    proto.String(uploadedMedia.DirectPath),
				MediaKey:      uploadedMedia.MediaKey,
				FileEncSHA256: uploadedMedia.FileEncSHA256,
				FileSHA256:    uploadedMedia.FileSHA256,
				FileLength:    proto.Uint64(uploadedMedia.FileLength),
			},
		}
	case "video":
		msg = &waProto.Message{
			VideoMessage: &waProto.VideoMessage{
				Caption:       proto.String(req.Caption),
				Mimetype:      proto.String("video/mp4"),
				URL:           proto.String(uploadedMedia.URL),
				DirectPath:    proto.String(uploadedMedia.DirectPath),
				MediaKey:      uploadedMedia.MediaKey,
				FileEncSHA256: uploadedMedia.FileEncSHA256,
				FileSHA256:    uploadedMedia.FileSHA256,
				FileLength:    proto.Uint64(uploadedMedia.FileLength),
			},
		}
	case "document":
		msg = &waProto.Message{
			DocumentMessage: &waProto.DocumentMessage{
				Title:         proto.String(req.Caption),
				Mimetype:      proto.String("application/pdf"),
				URL:           proto.String(uploadedMedia.URL),
				DirectPath:    proto.String(uploadedMedia.DirectPath),
				MediaKey:      uploadedMedia.MediaKey,
				FileEncSHA256: uploadedMedia.FileEncSHA256,
				FileSHA256:    uploadedMedia.FileSHA256,
				FileLength:    proto.Uint64(uploadedMedia.FileLength),
			},
		}
	case "audio":
		msg = &waProto.Message{
			AudioMessage: &waProto.AudioMessage{
				Mimetype:      proto.String("audio/opus"),
				URL:           proto.String(uploadedMedia.URL),
				DirectPath:    proto.String(uploadedMedia.DirectPath),
				MediaKey:      uploadedMedia.MediaKey,
				FileEncSHA256: uploadedMedia.FileEncSHA256,
				FileSHA256:    uploadedMedia.FileSHA256,
				FileLength:    proto.Uint64(uploadedMedia.FileLength),
			},
		}
	case "sticker":
		msg = &waProto.Message{
			StickerMessage: &waProto.StickerMessage{
				Mimetype: 	proto.String("image/webp"),
				URL:           proto.String(uploadedMedia.URL),
				DirectPath:    proto.String(uploadedMedia.DirectPath),
				MediaKey:      uploadedMedia.MediaKey,
				FileEncSHA256: uploadedMedia.FileEncSHA256,
				FileSHA256:    uploadedMedia.FileSHA256,
				FileLength:    proto.Uint64(uploadedMedia.FileLength),
			},
		}
	// Handle other media types (video, document, etc.) if needed
	default:
		http.Error(w, "Unsupported media type", http.StatusBadRequest)
		return
	}

	// Send the media message
	_, err = client.SendMessage(context.Background(), jid, msg)
	if err != nil {
		http.Error(w, "Failed to send media", http.StatusInternalServerError)
		return
	}

	if debugMode {
		fmt.Printf("✅ Sent %s to %s: %s\n", req.MediaType, req.JID, req.FilePath)
	}

	w.Write([]byte("✅ Media sent"))
}
