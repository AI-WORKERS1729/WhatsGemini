# WhatsappAutomation

To set up a **GitHub Codespace environment** with **CGO** support (which allows Go to use C code for compiling), we'll walk through:

- **Setting up MinGW (for CGO)** in GitHub Codespaces.
- **Using Go and C libraries** in your GitHub Codespace with a working environment for CGO.

GitHub Codespaces are Linux-based environments, so **CGO support** is easier because most of the required tools (like GCC) are already available in the environment. However, we’ll need to make sure that we properly set up everything so that **MinGW** and **Go with CGO** work smoothly.

---

### ✅ Step 1: Create a New GitHub Codespace

1. Go to [GitHub](https://github.com/).
2. Navigate to your repository or create a new one.
3. Click the **Code** button and select **Open with Codespaces** → **New codespace**.

This will launch a Linux-based environment inside GitHub Codespaces, which is ideal for Go development with CGO.

---

### ✅ Step 2: Install MinGW in GitHub Codespaces

In the GitHub Codespace terminal, run the following commands to install **MinGW** (this installs GCC, G++, and other tools needed for CGO):

```bash
# Update and install necessary packages
sudo apt update
sudo apt install -y build-essential gcc g++ make mingw-w64

# Verify installation
gcc --version
g++ --version
```

This should set up **MinGW-w64**, which includes both the **32-bit and 64-bit Windows toolchains** for compiling with CGO. You'll also have GCC and G++ for C and C++ compilation.

---

### ✅ Step 3: Set Up Go in GitHub Codespace

Go should already be installed in GitHub Codespaces, but to verify:

```bash
go version
```

If Go is not installed, run:

```bash
sudo apt install -y golang
```

---

### ✅ Step 4: Set CGO_ENABLED to 1

To enable **CGO** (which allows Go to call C code), set the **CGO_ENABLED** environment variable to `1` and ensure the Go environment is using CGO for Windows or other compiled code.

In the terminal, you can run:

```bash
export CGO_ENABLED=1
```

For **permanent setting** (every time you open the terminal), add this to your `.bashrc` file:

```bash
echo "export CGO_ENABLED=1" >> ~/.bashrc
source ~/.bashrc
```

---

### ✅ Step 5: Create the `go.mod` File for Your Project

If you don’t have a Go project yet, initialize one in your Codespace:

```bash
go mod init your_project_name
```

Then, you can update the **go.mod** to include the dependencies, such as the WhatsMeow package:

```go
module your_project_name

go 1.21

require (
    go.mau.fi/whatsmeow v0.0.0-20250417131650-164ddf482526
    modernc.org/sqlite v1.27.0
)
```

---

### ✅ Step 6: Build and Run

Now, you can build and run the application using the following commands:

```bash
go mod tidy
go run main.go
```

- This should print a QR code to the terminal.
- Scan the QR code with WhatsApp to authenticate.
- Once authenticated, a message will be sent to the specified number.

---

### ✅ Step 7: Handling CGO Dependencies in Codespace

Since you’re using **MinGW** inside GitHub Codespaces (Linux environment), it should automatically support CGO. You can also verify that **CGO** is enabled by checking:

```bash
go env CGO_ENABLED
```

It should return `1`, indicating that CGO is enabled and ready to work with the C compiler (MinGW in this case).

---

### 🎉 Conclusion

By following these steps, you’ve:

1. Set up **MinGW** inside **GitHub Codespaces** for **CGO**.
2. Configured a **Go project** with **WhatsMeow**.
3. Ensured **SQLite** is working with **CGO**.
4. Successfully run the WhatsMeow client, sending messages via WhatsApp.
