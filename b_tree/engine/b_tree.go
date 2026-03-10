package engine

import (
	"encoding/gob"
	"os"
)

type Page struct {
	Keys     []string
	Values   [][]byte
	Children []*Page
	IsLeaf   bool
}

type BTree struct {
	Root *Page
}

func (t *BTree) Get(key string) ([]byte, bool) {
	if t.Root == nil {
		return nil, false
	}
	return t.Root.search(key)
}

func (t *BTree) Put(key string, value []byte) {
	if t.Root == nil {
		t.Root = &Page{IsLeaf: true}
	}
	t.Root.insert(key, value)
	t.SaveToFile("btree.db")
}

func (t *BTree) SaveToFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)

	return encoder.Encode(t.Root)
}

func LoadFromFile(filename string) (*BTree, error) {
	EnsureDatabaseExists(filename)

	info, err := os.Stat(filename)
	if os.IsNotExist(err) || (err == nil && info.Size() == 0) {
		return &BTree{Root: &Page{IsLeaf: true}}, nil
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var root Page
	decoder := gob.NewDecoder(file)

	err = decoder.Decode(&root)
	if err != nil {
		return nil, err
	}

	return &BTree{Root: &root}, nil
}

func EnsureDatabaseExists(filename string) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		f, _ := os.Create(filename)
		f.Close()
	}
}

func (p *Page) insert(key string, value []byte) {
	i := 0
	for i < len(p.Keys) && key > p.Keys[i] {
		i++
	}

	if p.IsLeaf {
		p.Keys = append(p.Keys, "")
		copy(p.Keys[i+1:], p.Keys[i:])
		p.Keys[i] = key

		p.Values = append(p.Values, nil)
		copy(p.Values[i+1:], p.Values[i:])
		p.Values[i] = value
	} else {
		p.Children[i].insert(key, value)
	}
}

func (p *Page) search(key string) ([]byte, bool) {
	i := 0
	for i < len(p.Keys) && key > p.Keys[i] {
		i++
	}

	if i < len(p.Keys) && key == p.Keys[i] {
		return p.Values[i], true
	}

	if p.IsLeaf {
		return nil, false
	}

	return p.Children[i].search(key)
}
