package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/bogdanticu88/AuthBridge/cmd"
)

func main() {
	// Configure logging
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	if err := cmd.Execute(); err != nil {
		log.Error().Err(err).Msg("failed to execute command")
		os.Exit(1)
	}
}
