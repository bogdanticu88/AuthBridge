package cmd

import (
	"context"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/bogdanticu88/AuthBridge/internal/sync"
)

var (
	s3Bucket string
	s3Region string
	s3Key    string
)

var cloudCmd = &cobra.Command{
	Use:   "cloud",
	Short: "Sync credentials with cloud storage",
}

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push encrypted vault to cloud",
	Run: func(cmd *cobra.Command, args []string) {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get user home directory")
		}
		dbPath := filepath.Join(home, ".authbridge", "credentials.db")

		client, err := sync.NewS3SyncClient(context.Background(), s3Bucket, s3Region)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to initialize cloud client")
		}

		log.Info().Str("bucket", s3Bucket).Str("key", s3Key).Msg("Pushing vault to cloud...")
		if err := client.Push(context.Background(), s3Key, dbPath); err != nil {
			log.Fatal().Err(err).Msg("failed to push vault")
		}

		log.Info().Msg("Vault pushed successfully")
	},
}

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull encrypted vault from cloud",
	Run: func(cmd *cobra.Command, args []string) {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get user home directory")
		}
		dbPath := filepath.Join(home, ".authbridge", "credentials.db")

		client, err := sync.NewS3SyncClient(context.Background(), s3Bucket, s3Region)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to initialize cloud client")
		}

		log.Info().Str("bucket", s3Bucket).Str("key", s3Key).Msg("Pulling vault from cloud...")
		if err := client.Pull(context.Background(), s3Key, dbPath); err != nil {
			log.Fatal().Err(err).Msg("failed to pull vault")
		}

		log.Info().Msg("Vault pulled successfully. Restart the daemon to see changes.")
	},
}

func init() {
	cloudCmd.PersistentFlags().StringVar(&s3Bucket, "bucket", "", "S3 bucket name")
	cloudCmd.PersistentFlags().StringVar(&s3Region, "region", "us-east-1", "AWS region")
	cloudCmd.PersistentFlags().StringVar(&s3Key, "key", "authbridge/vault.db", "S3 key (path in bucket)")
	
	cloudCmd.MarkPersistentFlagRequired("bucket")

	cloudCmd.AddCommand(pushCmd)
	cloudCmd.AddCommand(pullCmd)
	rootCmd.AddCommand(cloudCmd)
}
