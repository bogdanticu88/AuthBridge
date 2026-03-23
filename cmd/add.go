package cmd

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/twister69/authbridge/internal/store"
)

var (
	credName     string
	credType     string
	credToken    string
	credMetadata string
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new credential",
	Run: func(cmd *cobra.Command, args []string) {
		home, _ := os.UserHomeDir()
		baseDir := filepath.Join(home, ".authbridge")
		os.MkdirAll(baseDir, 0700)
		dbPath := filepath.Join(baseDir, "credentials.db")

		s, err := store.NewSQLiteStore(dbPath)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to open store")
		}
		defer s.Close()

		e, err := store.NewEncryptionManager("")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to initialize encryption")
		}

		encryptedToken, err := e.Encrypt(credToken)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to encrypt token")
		}

		cred := &store.Credential{
			ID:         uuid.New().String(),
			Name:       credName,
			Type:       credType,
			Token:      encryptedToken,
			Metadata:   credMetadata,
			CreatedAt:  time.Now(),
			UsageCount: 0,
		}

		if err := s.AddCredential(context.Background(), cred); err != nil {
			log.Fatal().Err(err).Msg("failed to add credential")
		}

		log.Info().Str("name", credName).Msg("Credential added successfully")
	},
}

func init() {
	addCmd.Flags().StringVarP(&credName, "name", "n", "", "name of the credential (required)")
	addCmd.Flags().StringVarP(&credType, "type", "t", "jwt", "type of the credential (jwt, oauth2, basic, cookie, etc.)")
	addCmd.Flags().StringVarP(&credToken, "token", "T", "", "the actual token/credential string (required)")
	addCmd.Flags().StringVarP(&credMetadata, "metadata", "m", "", "JSON metadata for complex auth types (e.g. refresh_token for oauth2)")
	addCmd.MarkFlagRequired("name")
	addCmd.MarkFlagRequired("token")
	rootCmd.AddCommand(addCmd)
}
