# Telegram CLI Spammer

A command-line tool for Telegram that allows you to authenticate, list chats, and send multiple images to chats using MTProto.

## Features

- 🔐 **Login**: Authenticate with Telegram using your phone number
- 📋 **List Chats**: View all your chats (private, groups, channels)
- 📸 **Spam Images**: Send multiple images from a folder to any chat
- 💾 **Session Persistence**: Stay logged in between uses

## Prerequisites

- Go 1.21 or higher
- Telegram API credentials (API ID and API Hash)

## Getting Your Telegram API Credentials

1. Visit [https://my.telegram.org/auth](https://my.telegram.org/auth)
2. Log in with your phone number
3. Go to "API development tools"
4. Create a new application (if you haven't already)
5. Copy your `api_id` and `api_hash`

## Installation

1. Clone the repository:

```bash
git clone https://github.com/FabioSM46/telegram-cli-spammer.git
cd telegram-cli-spammer
```

2. Install dependencies:

```bash
go mod download
```

3. Build the application:

```bash
go build -o telegram-spammer
```

## Configuration

Create a `.telegram-config` file in the project directory or your home directory with your API credentials:

```
API_ID=your_api_id
API_HASH=your_api_hash
```

Example:

```
API_ID=12345678
API_HASH=0123456789abcdef0123456789abcdef
```

## Usage

### 1. Login

First, authenticate with your Telegram account:

```bash
./telegram-spammer login
```

Or specify the phone number directly:

```bash
./telegram-spammer login --phone +1234567890
```

You'll be prompted to:

- Enter your phone number (if not provided)
- Enter the verification code sent to your Telegram app
- Enter your 2FA password (if enabled)

The session will be saved to `session.json` for future use.

### 2. List Chats

View all your available chats:

```bash
./telegram-spammer list
```

This will display:

- Private chats with users
- Group chats
- Channels and supergroups

Example output:

```
=== Available Chats ===

User   | ID: 123456789 | John Doe (@johndoe)
Group  | ID: 987654321 | My Friends
Channel | ID: 111222333 (@mychannel) | My Channel
SuperGroup | ID: 444555666 | Large Group
```

### 3. Spam Images

Send all images from the `./images` folder to a specific chat:

```bash
./telegram-spammer spam <chat_id>
```

Example:

```bash
./telegram-spammer spam 123456789
```

#### Options

- `-d, --images-dir`: Specify a different images directory (default: `./images`)
- `-t, --delay`: Delay between messages in milliseconds (default: `1000`)

Example with options:

```bash
./telegram-spammer spam 123456789 --images-dir ./my-photos --delay 2000
```

## Images Folder

Create an `images` folder in the project directory and place your images there:

```bash
mkdir images
# Copy your images to the images folder
cp /path/to/your/images/*.jpg ./images/
```

Supported image formats:

- JPG/JPEG
- PNG
- GIF
- WebP

## Project Structure

```
telegram-cli-spammer/
├── cmd/                    # CLI commands
│   ├── root.go            # Root command
│   ├── login.go           # Login command
│   ├── list.go            # List chats command
│   └── spam.go            # Spam images command
├── internal/
│   └── telegram/
│       └── client.go      # Telegram client implementation
├── images/                # Place your images here
├── main.go               # Application entry point
├── go.mod                # Go module file
├── .telegram-config      # Your API credentials (create this)
└── README.md            # This file
```

## Important Notes

⚠️ **Rate Limits**: Telegram has rate limits. Sending too many messages quickly may result in temporary restrictions on your account.

⚠️ **Privacy**: Keep your `.telegram-config` and `session.json` files private. Never commit them to version control.

⚠️ **Terms of Service**: Make sure your use of this tool complies with Telegram's Terms of Service.

## Troubleshooting

### "config file not found" error

Create a `.telegram-config` file with your API_ID and API_HASH.

### "not logged in" error

Run the `login` command first to authenticate.

### "chat with ID not found" error

Use the `list` command to see available chat IDs.

### Images not sending

Ensure:

- The images directory exists and contains supported image formats
- You have permission to send messages in the target chat
- You're not hitting Telegram's rate limits

## Development

To run without building:

```bash
go run main.go login
go run main.go list
go run main.go spam <chat_id>
```

## License

MIT License - feel free to use and modify as needed.

## Disclaimer

This tool is for educational purposes. Use responsibly and in accordance with Telegram's Terms of Service. The developers are not responsible for any misuse of this tool.
