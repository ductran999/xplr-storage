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
	"storage-journey/b_tree/engine"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// connectCmd represents the connect command
var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "interactive mode",
	Long:  `interactive mode`,
	Run: func(cmd *cobra.Command, args []string) {
		runShell()
	},
}

func runShell() {
	log.Println("Connect DB use btree engine...")
	app, err := engine.LoadFromFile("btree.db")
	if err != nil {
		log.Fatalln("connect error", err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("b-tree$ ")

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

		handleCommand(app, line)
	}
}

func handleCommand(db *engine.BTree, line string) {
	parts := strings.Fields(line)
	cmd := parts[0]

	start := time.Now()
	slog.Info("command start", "cmd", cmd)

	switch cmd {

	case "get":
		handleGet(db, parts)
	case "put":
		handlePut(db, parts)
	default:
		fmt.Println("commands: get put list exit")
		slog.Warn("unknown command", "command", cmd)
	}

	slog.Info("command finished",
		"cmd", cmd,
		"duration", time.Since(start),
	)
}

func handleGet(db *engine.BTree, parts []string) {
	if len(parts) < 2 {
		fmt.Println("usage: get <key>")
		return
	}

	key := parts[1]
	val, found := db.Get(key)
	if !found {
		slog.Error("not found")
		return
	}
	slog.Info("key found", "key:", key, "---> value", string(val))
}

func handlePut(db *engine.BTree, parts []string) {
	if len(parts) < 3 {
		fmt.Println("usage: put <key> <value>")
		return
	}

	key := parts[1]
	value := parts[2]

	db.Put(key, []byte(value))
	slog.Info("put success", "key", key)
}

func init() {
	rootCmd.AddCommand(connectCmd)
}
