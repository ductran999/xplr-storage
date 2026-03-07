/*
Copyright © 2026 ductran999
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "bitcask",
	Short: "mini bitcask is a simple implementation of bitcask.",
	Long:  "Mini bitcask is a simple implementation of bitcask. This project for education purpose not for production",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
