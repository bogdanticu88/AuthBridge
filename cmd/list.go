package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/bogdanticu88/AuthBridge/internal/store"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all stored credentials",
	Run: func(cmd *cobra.Command, args []string) {
		home, _ := os.UserHomeDir()
		dbPath := filepath.Join(home, ".authbridge", "credentials.db")

		s, err := store.NewSQLiteStore(dbPath)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to open store")
		}
		defer s.Close()

		creds, err := s.ListCredentials(context.Background())
		if err != nil {
			log.Fatal().Err(err).Msg("failed to list credentials")
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "NAME\tTYPE\tCREATED AT\tUSAGE\tLAST USED")
		for _, c := range creds {
			lastUsed := "Never"
			if !c.LastUsed.IsZero() {
				lastUsed = c.LastUsed.Format("2006-01-02 15:04:05")
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\n",
				c.Name, c.Type, c.CreatedAt.Format("2006-01-02 15:04:05"), c.UsageCount, lastUsed)
		}
		w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
