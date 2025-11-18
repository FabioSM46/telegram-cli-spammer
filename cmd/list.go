package cmd

import (
	"context"
	"fmt"

	"github.com/FabioSM46/telegram-cli-spammer/internal/telegram"
	"github.com/spf13/cobra"
)

var listChatsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available chats",
	Long:  `Display a list of all chats (private chats, groups, and channels) accessible to the authenticated user.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		client, err := telegram.NewClient()
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		return client.ListChats(ctx)
	},
}
