/*
Copyright © 2026 ductran999
*/
package cmd

import (
	"log"
	"os"
	"storage-journey/bitcask/engine"

	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show db status",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Analyzing...")
		db, _ := engine.Open("data.db")
		fileInfo, _ := os.Stat("data.db")
		fileSizeMB := float64(fileInfo.Size()) / 1024 / 1024

		log.Printf(">>> Total keys in RAM: %d\n", len(db.ListKeys()))
		log.Printf(">>> Size: %.2f MB\n", fileSizeMB)
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
