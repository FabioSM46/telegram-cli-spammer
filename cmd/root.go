package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "telegram-spammer",
	Short: "A Telegram CLI tool for managing chats and sending messages",
	Long:  `A command-line tool to authenticate with Telegram, list chats, and send images to chats.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(listChatsCmd)
	rootCmd.AddCommand(spamCmd)
}
