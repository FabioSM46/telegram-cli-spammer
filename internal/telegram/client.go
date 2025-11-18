package telegram

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	sessionFile = "session.json"
	configFile  = ".telegram-config"
)

// terminalAuth implements auth.UserAuthenticator for interactive authentication
type terminalAuth struct {
	phone  string
	reader *bufio.Reader
}

func (t *terminalAuth) Phone(ctx context.Context) (string, error) {
	return t.phone, nil
}

func (t *terminalAuth) Password(ctx context.Context) (string, error) {
	fmt.Print("Enter your 2FA password: ")
	password, err := t.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(password), nil
}

func (t *terminalAuth) Code(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
	fmt.Print("Enter the code you received: ")
	code, err := t.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(code), nil
}

func (t *terminalAuth) AcceptTermsOfService(ctx context.Context, tos tg.HelpTermsOfService) error {
	return nil
}

func (t *terminalAuth) SignUp(ctx context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{}, fmt.Errorf("sign up not supported, please register with an official Telegram client first")
}

type Client struct {
	client *telegram.Client
	api    *tg.Client
	appID  int
	appHash string
}

// NewClient creates a new Telegram client instance
func NewClient() (*Client, error) {
	appID, appHash, err := loadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w\nPlease create a .telegram-config file with your API ID and Hash", err)
	}

	return &Client{
		appID:   appID,
		appHash: appHash,
	}, nil
}

// loadConfig loads API credentials from config file
func loadConfig() (int, string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return 0, "", err
	}

	configPath := filepath.Join(homeDir, configFile)
	
	// Also check current directory
	if _, err := os.Stat(configFile); err == nil {
		configPath = configFile
	} else if _, err := os.Stat(configPath); err != nil {
		return 0, "", fmt.Errorf("config file not found. Create %s or ~/%s with API_ID and API_HASH", configFile, configFile)
	}

	file, err := os.Open(configPath)
	if err != nil {
		return 0, "", err
	}
	defer file.Close()

	var appID int
	var appHash string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "API_ID":
			fmt.Sscanf(value, "%d", &appID)
		case "API_HASH":
			appHash = value
		}
	}

	if appID == 0 || appHash == "" {
		return 0, "", fmt.Errorf("API_ID and API_HASH must be set in config file")
	}

	return appID, appHash, nil
}

// createLogger creates a zap logger
func createLogger() (*zap.Logger, error) {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	return config.Build()
}

// Login authenticates with Telegram using phone number
func (c *Client) Login(ctx context.Context, phone string) error {
	logger, err := createLogger()
	if err != nil {
		return err
	}
	defer logger.Sync()

	sessionStorage := &telegram.FileSessionStorage{
		Path: sessionFile,
	}

	client := telegram.NewClient(c.appID, c.appHash, telegram.Options{
		SessionStorage: sessionStorage,
		Logger:         logger,
	})

	return client.Run(ctx, func(ctx context.Context) error {
		reader := bufio.NewReader(os.Stdin)

		// Custom authenticator that handles both code and 2FA password interactively
		termAuth := &terminalAuth{
			phone:  phone,
			reader: reader,
		}

		flow := auth.NewFlow(
			termAuth,
			auth.SendCodeOptions{},
		)

		if err := client.Auth().IfNecessary(ctx, flow); err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}

		status, err := client.Auth().Status(ctx)
		if err != nil {
			return err
		}

		if !status.Authorized {
			return fmt.Errorf("not authorized")
		}

		user, ok := status.User.AsNotEmpty()
		if !ok {
			return fmt.Errorf("user is empty")
		}

		fmt.Printf("\n✓ Successfully logged in as %s %s (ID: %d)\n", user.FirstName, user.LastName, user.ID)
		fmt.Printf("Session saved to %s\n", sessionFile)

		return nil
	})
}

// ListChats lists all available chats
func (c *Client) ListChats(ctx context.Context) error {
	logger, err := createLogger()
	if err != nil {
		return err
	}
	defer logger.Sync()

	sessionStorage := &telegram.FileSessionStorage{
		Path: sessionFile,
	}

	if _, err := os.Stat(sessionFile); os.IsNotExist(err) {
		return fmt.Errorf("not logged in. Please run 'login' command first")
	}

	client := telegram.NewClient(c.appID, c.appHash, telegram.Options{
		SessionStorage: sessionStorage,
		Logger:         logger,
	})

	return client.Run(ctx, func(ctx context.Context) error {
		api := client.API()

		// Get dialogs
		dialogs, err := api.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
			OffsetPeer: &tg.InputPeerEmpty{},
			Limit:      100,
		})
		if err != nil {
			return fmt.Errorf("failed to get dialogs: %w", err)
		}

		dialogsSlice, ok := dialogs.(*tg.MessagesDialogsSlice)
		if !ok {
			// Try full dialogs
			dialogsFull, ok := dialogs.(*tg.MessagesDialogs)
			if !ok {
				return fmt.Errorf("unexpected dialogs type")
			}
			dialogsSlice = &tg.MessagesDialogsSlice{
				Dialogs: dialogsFull.Dialogs,
				Messages: dialogsFull.Messages,
				Chats: dialogsFull.Chats,
				Users: dialogsFull.Users,
			}
		}

		fmt.Println("\n=== Available Chats ===")

		// Create maps for quick lookup
		userMap := make(map[int64]*tg.User)
		for _, u := range dialogsSlice.Users {
			if user, ok := u.(*tg.User); ok {
				userMap[user.ID] = user
			}
		}

		chatMap := make(map[int64]tg.ChatClass)
		for _, ch := range dialogsSlice.Chats {
			switch chat := ch.(type) {
			case *tg.Chat:
				chatMap[chat.ID] = chat
			case *tg.Channel:
				chatMap[chat.ID] = chat
			}
		}

		for _, d := range dialogsSlice.Dialogs {
			dialog, ok := d.(*tg.Dialog)
			if !ok {
				continue
			}

			switch peer := dialog.Peer.(type) {
			case *tg.PeerUser:
				if user, ok := userMap[peer.UserID]; ok {
					name := fmt.Sprintf("%s %s", user.FirstName, user.LastName)
					if user.Username != "" {
						name += fmt.Sprintf(" (@%s)", user.Username)
					}
					fmt.Printf("User   | ID: %d | %s\n", peer.UserID, name)
				}
			case *tg.PeerChat:
				if chat, ok := chatMap[peer.ChatID]; ok {
					switch c := chat.(type) {
					case *tg.Chat:
						fmt.Printf("Group  | ID: %d | %s\n", peer.ChatID, c.Title)
					}
				}
			case *tg.PeerChannel:
				if chat, ok := chatMap[peer.ChannelID]; ok {
					switch c := chat.(type) {
					case *tg.Channel:
						channelType := "Channel"
						if c.Megagroup {
							channelType = "SuperGroup"
						}
						username := ""
						if c.Username != "" {
							username = fmt.Sprintf(" (@%s)", c.Username)
						}
						fmt.Printf("%s | ID: %d%s | %s\n", channelType, peer.ChannelID, username, c.Title)
					}
				}
			}
		}

		fmt.Println("\nNote: Use the ID value when sending messages with the 'spam' command")

		return nil
	})
}

// SpamImages sends all images from a directory to a specific chat
func (c *Client) SpamImages(ctx context.Context, chatID int64, imagesDir string, delay int) error {
	logger, err := createLogger()
	if err != nil {
		return err
	}
	defer logger.Sync()

	sessionStorage := &telegram.FileSessionStorage{
		Path: sessionFile,
	}

	if _, err := os.Stat(sessionFile); os.IsNotExist(err) {
		return fmt.Errorf("not logged in. Please run 'login' command first")
	}

	if _, err := os.Stat(imagesDir); os.IsNotExist(err) {
		return fmt.Errorf("images directory '%s' does not exist", imagesDir)
	}

	client := telegram.NewClient(c.appID, c.appHash, telegram.Options{
		SessionStorage: sessionStorage,
		Logger:         logger,
	})

	return client.Run(ctx, func(ctx context.Context) error {
		api := client.API()

		// Read images from directory
		files, err := os.ReadDir(imagesDir)
		if err != nil {
			return fmt.Errorf("failed to read images directory: %w", err)
		}

		imageFiles := []string{}
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			ext := strings.ToLower(filepath.Ext(file.Name()))
			if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".webp" {
				imageFiles = append(imageFiles, filepath.Join(imagesDir, file.Name()))
			}
		}

		if len(imageFiles) == 0 {
			return fmt.Errorf("no image files found in %s", imagesDir)
		}

		fmt.Printf("Found %d images to send\n", len(imageFiles))

		// Determine the peer type
		var inputPeer tg.InputPeerClass

		// First, try to get the entity to determine the correct peer type
		dialogs, err := api.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
			OffsetPeer: &tg.InputPeerEmpty{},
			Limit:      100,
		})
		if err != nil {
			return fmt.Errorf("failed to get dialogs: %w", err)
		}

		found := false
		switch d := dialogs.(type) {
		case *tg.MessagesDialogsSlice:
			for _, dialog := range d.Dialogs {
				if dlg, ok := dialog.(*tg.Dialog); ok {
					switch peer := dlg.Peer.(type) {
					case *tg.PeerUser:
						if peer.UserID == chatID {
							inputPeer = &tg.InputPeerUser{UserID: chatID}
							// We need access hash, let's get it from users
							for _, u := range d.Users {
								if user, ok := u.(*tg.User); ok && user.ID == chatID {
									inputPeer = &tg.InputPeerUser{UserID: chatID, AccessHash: user.AccessHash}
									found = true
									break
								}
							}
						}
					case *tg.PeerChat:
						if peer.ChatID == chatID {
							inputPeer = &tg.InputPeerChat{ChatID: chatID}
							found = true
						}
					case *tg.PeerChannel:
						if peer.ChannelID == chatID {
							// Find access hash from chats
							for _, ch := range d.Chats {
								if channel, ok := ch.(*tg.Channel); ok && channel.ID == chatID {
									inputPeer = &tg.InputPeerChannel{ChannelID: chatID, AccessHash: channel.AccessHash}
									found = true
									break
								}
							}
						}
					}
					if found {
						break
					}
				}
			}
		case *tg.MessagesDialogs:
			for _, dialog := range d.Dialogs {
				if dlg, ok := dialog.(*tg.Dialog); ok {
					switch peer := dlg.Peer.(type) {
					case *tg.PeerUser:
						if peer.UserID == chatID {
							for _, u := range d.Users {
								if user, ok := u.(*tg.User); ok && user.ID == chatID {
									inputPeer = &tg.InputPeerUser{UserID: chatID, AccessHash: user.AccessHash}
									found = true
									break
								}
							}
						}
					case *tg.PeerChat:
						if peer.ChatID == chatID {
							inputPeer = &tg.InputPeerChat{ChatID: chatID}
							found = true
						}
					case *tg.PeerChannel:
						if peer.ChannelID == chatID {
							for _, ch := range d.Chats {
								if channel, ok := ch.(*tg.Channel); ok && channel.ID == chatID {
									inputPeer = &tg.InputPeerChannel{ChannelID: chatID, AccessHash: channel.AccessHash}
									found = true
									break
								}
							}
						}
					}
					if found {
						break
					}
				}
			}
		}

		if !found {
			return fmt.Errorf("chat with ID %d not found. Use 'list' command to see available chats", chatID)
		}

		// Send images using uploader
		u := uploader.NewUploader(api)
		
		for i, imagePath := range imageFiles {
			fmt.Printf("[%d/%d] Sending %s...\n", i+1, len(imageFiles), filepath.Base(imagePath))

			// Open and read file
			file, err := os.Open(imagePath)
			if err != nil {
				fmt.Printf("  ✗ Failed to open file: %v\n", err)
				continue
			}

			// Get file info
			fileInfo, err := file.Stat()
			if err != nil {
				file.Close()
				fmt.Printf("  ✗ Failed to get file info: %v\n", err)
				continue
			}

			// Upload file
			upload, err := u.Upload(ctx, uploader.NewUpload(filepath.Base(imagePath), file, fileInfo.Size()))
			file.Close()
			
			if err != nil {
				fmt.Printf("  ✗ Failed to upload: %v\n", err)
				continue
			}

			// Send as photo
			_, err = api.MessagesSendMedia(ctx, &tg.MessagesSendMediaRequest{
				Peer: inputPeer,
				Media: &tg.InputMediaUploadedPhoto{
					File: upload,
				},
				RandomID: time.Now().UnixNano(),
			})

			if err != nil {
				fmt.Printf("  ✗ Failed to send: %v\n", err)
				continue
			}

			fmt.Printf("  ✓ Sent successfully\n")

			// Delay before next message
			if i < len(imageFiles)-1 && delay > 0 {
				time.Sleep(time.Duration(delay) * time.Millisecond)
			}
		}

		fmt.Printf("\nCompleted sending %d images\n", len(imageFiles))

		return nil
	})
}
