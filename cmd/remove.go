package cmd

import (
	"context"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/twister69/authbridge/internal/store"
)

var removeCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove a credential",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		home, _ := os.UserHomeDir()
		dbPath := filepath.Join(home, ".authbridge", "credentials.db")

		s, err := store.NewSQLiteStore(dbPath)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to open store")
		}
		defer s.Close()

		if err := s.DeleteCredential(context.Background(), name); err != nil {
			log.Fatal().Err(err).Msg("failed to delete credential")
		}

		log.Info().Str("name", name).Msg("Credential removed successfully")
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
