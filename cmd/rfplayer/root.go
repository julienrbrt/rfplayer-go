package main

import "github.com/spf13/cobra"

// RFPlayerCmd returns the root command for the rfplayer command line tool.
func RFPlayerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rfplayer",
		Short: "RFPlayer-Go is a command line tool to interact with RFPlayer devices",
	}

	return cmd
}
