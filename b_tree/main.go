package main

import (
	"fmt"

	"github.com/google/btree"
)

type Item struct {
	Key   string
	Value string
}

func (i Item) Less(than btree.Item) bool {
	return i.Key < than.(Item).Key
}

func main() {
	tree := btree.New(32)

	fmt.Println("--- Inserting data ---")
	tree.ReplaceOrInsert(Item{Key: "user:1", Value: "Alice"})
	tree.ReplaceOrInsert(Item{Key: "user:2", Value: "Bob"})
	tree.ReplaceOrInsert(Item{Key: "user:3", Value: "Charlie"})

	fmt.Println("--- Searching data ---")
	item := tree.Get(Item{Key: "user:2"})
	if item != nil {
		fmt.Printf("Found: %s\n", item.(Item).Value)
	}

	fmt.Println("--- Range Query (All users) ---")
	tree.Ascend(func(i btree.Item) bool {
		user := i.(Item)
		fmt.Printf("%s: %s\n", user.Key, user.Value)
		return true
	})

	fmt.Println("--- Deleting user:1 ---")
	tree.Delete(Item{Key: "user:1"})

	fmt.Printf("Tree size: %d\n", tree.Len())
}
