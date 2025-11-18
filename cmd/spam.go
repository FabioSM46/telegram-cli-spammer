package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/FabioSM46/telegram-cli-spammer/internal/telegram"
	"github.com/spf13/cobra"
)

var spamCmd = &cobra.Command{
	Use:   "spam [chat_id]",
	Short: "Send images from the images folder to a chat",
	Long:  `Send all images from the ./images folder to the specified chat. Use the chat ID from the 'list' command.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		chatID := args[0]
		
		// Try to parse as integer
		chatIDInt, err := strconv.ParseInt(chatID, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid chat ID: %w", err)
		}

		imagesDir, _ := cmd.Flags().GetString("images-dir")
		delay, _ := cmd.Flags().GetInt("delay")

		ctx := context.Background()
		client, err := telegram.NewClient()
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		return client.SpamImages(ctx, chatIDInt, imagesDir, delay)
	},
}

func init() {
	spamCmd.Flags().StringP("images-dir", "d", "./images", "Directory containing images to send")
	spamCmd.Flags().IntP("delay", "t", 1000, "Delay between messages in milliseconds")
}
