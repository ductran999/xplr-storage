/*
Copyright © 2026 ductran999
*/
package cmd

import (
	"encoding/json"
	"log"
	"log/slog"
	"storage-journey/bitcask/engine"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

type User struct {
	Name     string `json:"name"`
	Age      int    `json:"age"`
	PhoneNum string `json:"phone_num"`
}

// feedCmd represents the feed command
var feedCmd = &cobra.Command{
	Use:   "feed",
	Short: "Feed mock data to file",
	Run: func(cmd *cobra.Command, args []string) {

		db, err := engine.Open("data.db")
		if err != nil {
			slog.Error("open db failed", "error", err)
			return
		}
		defer db.Close()

		gofakeit.Seed(0)

		total := 1_000_000
		for i := range total {
			key := uuid.New().String()
			user := User{
				Name:     gofakeit.Name(),
				Age:      gofakeit.Number(18, 80),
				PhoneNum: gofakeit.Phone(),
			}

			value, _ := json.Marshal(user)

			err := db.Put(key, value)
			if err != nil {
				slog.Error("put failed", "key", key, "error", err)
			}

			if i%10000 == 0 {
				percent := float64(i) / float64(total) * 100
				slog.Info("feeding progress",
					"records", i,
					"percent", percent,
				)
			}
		}

		log.Println("feed completed")
	},
}

func init() {
	rootCmd.AddCommand(feedCmd)
}
