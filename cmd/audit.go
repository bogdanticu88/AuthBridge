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

var (
	auditName  string
	auditLimit int
)

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Show audit logs of credential usage",
	Run: func(cmd *cobra.Command, args []string) {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get user home directory")
		}
		dbPath := filepath.Join(home, ".authbridge", "credentials.db")

		s, err := store.NewSQLiteStore(dbPath)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to open store")
		}
		defer s.Close()

		logs, err := s.ListAuditLogs(context.Background(), auditName, auditLimit)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to list audit logs")
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "TIMESTAMP\tACTION\tCREDENTIAL\tIP\tTOOL\tSTATUS")
		for _, l := range logs {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				l.Timestamp.Format("2006-01-02 15:04:05"),
				l.Action,
				l.CredentialName,
				l.SourceIP,
				l.SourceTool,
				l.Status,
			)
		}
		if err := w.Flush(); err != nil {
			log.Error().Err(err).Msg("failed to flush output")
		}
	},
}

func init() {
	auditCmd.Flags().StringVarP(&auditName, "name", "n", "", "filter by credential name")
	auditCmd.Flags().IntVarP(&auditLimit, "limit", "l", 50, "limit the number of log entries")
	rootCmd.AddCommand(auditCmd)
}
