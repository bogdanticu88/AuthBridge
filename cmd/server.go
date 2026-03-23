package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/bogdanticu88/AuthBridge/internal/api"
	"github.com/bogdanticu88/AuthBridge/internal/store"
)

var (
	port   int
	host   string
	apiKey string
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the AuthBridge daemon",
	Run: func(cmd *cobra.Command, args []string) {
		home, _ := os.UserHomeDir()
		baseDir := filepath.Join(home, ".authbridge")
		os.MkdirAll(baseDir, 0700)

		dbPath := filepath.Join(baseDir, "credentials.db")
		log.Info().Str("db", dbPath).Msg("Initializing store...")

		// Initialize store
		s, err := store.NewSQLiteStore(dbPath)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to initialize store")
		}
		defer s.Close()

		// Initialize encryption
		e, err := store.NewEncryptionManager("") // Use default key source
		if err != nil {
			log.Fatal().Err(err).Msg("failed to initialize encryption")
		}

		// Initialize and start API server
		addr := fmt.Sprintf("%s:%d", host, port)
		server := api.NewServer(addr, s, e, apiKey)

		go func() {
			log.Info().Str("addr", addr).Msg("AuthBridge daemon listening")
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatal().Err(err).Msg("failed to start API server")
			}
		}()

		// Graceful shutdown
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop

		log.Info().Msg("Shutting down daemon...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Error().Err(err).Msg("server forced to shutdown")
		}

		log.Info().Msg("Daemon stopped")
	},
}

func init() {
	startCmd.Flags().IntVarP(&port, "port", "p", 9999, "port to listen on")
	startCmd.Flags().StringVarP(&host, "host", "H", "127.0.0.1", "host to listen on")
	startCmd.Flags().StringVar(&apiKey, "api-key", "", "API key for hardening the Web GUI and REST API (optional)")
	rootCmd.AddCommand(startCmd)
}
