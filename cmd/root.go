package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "travel",
	Short: "Flight search tool powered by Google Flights",
	Long:  "A CLI tool that searches Google Flights for the best deals using their internal API.",
}

func Execute() error {
	return rootCmd.Execute()
}
