# Quick Start Guide

## Initial Setup

1. **Get Telegram API Credentials**

   - Go to https://my.telegram.org/auth
   - Login and create an app under "API development tools"
   - Note your `api_id` and `api_hash`

2. **Configure the Application**

   ```bash
   cp .telegram-config.example .telegram-config
   ```

   Edit `.telegram-config` and add your credentials:

   ```
   API_ID=your_actual_api_id
   API_HASH=your_actual_api_hash
   ```

3. **Build the Application**
   ```bash
   go build -o telegram-spammer
   ```

## Usage Flow

### Step 1: Login

```bash
./telegram-spammer login
```

Follow the prompts to enter your phone number and verification code.

### Step 2: List Your Chats

```bash
./telegram-spammer list
```

Copy the ID of the chat where you want to send images.

### Step 3: Prepare Images

```bash
# Add images to the images folder
cp /path/to/your/photos/*.jpg ./images/
```

### Step 4: Send Images

```bash
./telegram-spammer spam <chat_id>
```

Example:

```bash
./telegram-spammer spam 123456789
```

## Tips

- **Custom delay**: Use `--delay` to set milliseconds between messages

  ```bash
  ./telegram-spammer spam 123456789 --delay 2000
  ```

- **Different folder**: Use `--images-dir` for a different image source

  ```bash
  ./telegram-spammer spam 123456789 --images-dir ./my-photos
  ```

- **Stay safe**: Don't spam too fast to avoid Telegram rate limits

## Troubleshooting

- **Not logged in**: Run `./telegram-spammer login` first
- **Chat not found**: Use `./telegram-spammer list` to get valid chat IDs
- **No images found**: Make sure `./images` folder contains .jpg, .png, .gif, or .webp files
