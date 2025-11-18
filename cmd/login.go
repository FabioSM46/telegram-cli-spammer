package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/FabioSM46/telegram-cli-spammer/internal/telegram"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to Telegram with phone number",
	Long:  `Authenticate with Telegram using your phone number and create a session file for future use.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		phone, _ := cmd.Flags().GetString("phone")
		
		reader := bufio.NewReader(os.Stdin)
		
		if phone == "" {
			fmt.Print("Enter your phone number (with country code, e.g., +1234567890): ")
			phoneInput, _ := reader.ReadString('\n')
			phone = strings.TrimSpace(phoneInput)
		}

		ctx := context.Background()
		client, err := telegram.NewClient()
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		return client.Login(ctx, phone)
	},
}

func init() {
	loginCmd.Flags().StringP("phone", "p", "", "Phone number with country code (e.g., +1234567890)")
}
