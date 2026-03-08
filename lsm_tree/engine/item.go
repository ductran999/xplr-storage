package engine

import "github.com/google/btree"

type Item struct {
	Key   string
	Value []byte
}

func (a Item) Less(b btree.Item) bool {
	return a.Key < b.(Item).Key
}
