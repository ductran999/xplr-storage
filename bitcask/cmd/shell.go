/*
Copyright © 2026 ductran999
*/
package cmd

import (
	"bufio"
	"fmt"
	"log"
	"log/slog"
	"os"
	"storage-journey/bitcask/engine"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var shellCmd = &cobra.Command{
	Use:   "connect",
	Short: "interact mode",
	Run: func(cmd *cobra.Command, args []string) {
		runShell()
	},
}

func runShell() {
	log.Println("Starting Bitcask Shell...")

	start := time.Now()
	db, err := engine.Open("data.db")
	if err != nil {
		slog.Error("open db failed", "error", err)
		return
	}
	defer db.Close()

	slog.Info("Bitcask Shell v1.0", "startup_time", time.Since(start))

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("bitcask$ ")

		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			continue
		}

		if line == "exit" {
			slog.Info("bye 👋")
			break
		}

		handleCommand(db, line)
	}
}

func handleCommand(db *engine.Bitcask, line string) {
	parts := strings.Fields(line)
	cmd := parts[0]

	start := time.Now()
	slog.Info("command start", "cmd", cmd)

	switch cmd {

	case "get":
		handleGet(db, parts)
	case "put":
		handlePut(db, parts)
	case "list":
		handleList(db, parts)
	default:
		fmt.Println("commands: get put list exit")
		slog.Warn("unknown command", "command", cmd)
	}

	slog.Info("command finished",
		"cmd", cmd,
		"duration", time.Since(start),
	)
}

func handleGet(db *engine.Bitcask, parts []string) {
	if len(parts) < 2 {
		fmt.Println("usage: get <key>")
		return
	}

	key := parts[1]

	val, err := db.Get(key)
	if err != nil {
		slog.Error("get failed", "key", key, "error", err)
		return
	}

	fmt.Println(string(val))
	slog.Info("get success", "key", key)
}

func handlePut(db *engine.Bitcask, parts []string) {
	if len(parts) < 3 {
		fmt.Println("usage: put <key> <value>")
		return
	}

	key := parts[1]
	value := parts[2]

	err := db.Put(key, []byte(value))
	if err != nil {
		slog.Error("put failed", "key", key, "error", err)
		return
	}

	slog.Info("put success", "key", key)
}

func handleList(db *engine.Bitcask, parts []string) {
	limit := 10
	offset := 0

	if len(parts) >= 2 {
		limit, _ = strconv.Atoi(parts[1])
	}

	if len(parts) >= 3 {
		offset, _ = strconv.Atoi(parts[2])
	}

	it := db.Iterator()

	count := 0
	skipped := 0

	for it.Next() {

		if skipped < offset {
			skipped++
			continue
		}

		fmt.Printf("%s -> %s\n", it.Key(), it.Value())

		count++
		if count >= limit {
			break
		}
	}

	if err := it.Error(); err != nil {
		slog.Error("iterator failed", "error", err)
	}

	slog.Info("list completed",
		"limit", limit,
		"offset", offset,
		"returned", count,
	)
}

func init() {
	rootCmd.AddCommand(shellCmd)
}
