package engine

import (
	"encoding/binary"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
)

type SparseEntry struct {
	Key    string
	Offset int64
}

type SSTable struct {
	FilePath    string
	SparseIndex []SparseEntry
}

func discoverSSTables(dataDir string) ([]*SSTable, error) {
	// 1. List all file in data foulder
	files, err := os.ReadDir(dataDir)
	if err != nil {
		return nil, err
	}

	// Find all .db files and load them as SStable
	var tables []*SSTable
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".db" {
			path := filepath.Join(dataDir, f.Name())
			if sst, err := loadSSTableFromFile(path); err == nil {
				tables = append(tables, sst)
			}
		}
	}

	// 2. Sort file by name (timestamp) by descending order
	sort.Slice(tables, func(i, j int) bool {
		return tables[i].FilePath > tables[j].FilePath
	})

	log.Printf("[SSTable]: Loaded %d files\n", len(tables))

	return tables, nil
}

func loadSSTableFromFile(path string) (*SSTable, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	fileSize := fi.Size()

	// 1. Read footer
	if _, err := f.Seek(fileSize-8, io.SeekStart); err != nil {
		return nil, err
	}

	var indexOffset int64
	if err := binary.Read(f, binary.LittleEndian, &indexOffset); err != nil {
		return nil, err
	}

	// 2. Jump to offset start of first index entry
	if _, err := f.Seek(indexOffset, io.SeekStart); err != nil {
		return nil, err
	}

	var entries []SparseEntry
	currentPos := indexOffset

	for currentPos < fileSize-8 {
		var keySize int64
		if err := binary.Read(f, binary.LittleEndian, &keySize); err != nil {
			return nil, err
		}

		keyBuf := make([]byte, keySize)
		if _, err := io.ReadFull(f, keyBuf); err != nil {
			return nil, err
		}

		var offset int64
		if err := binary.Read(f, binary.LittleEndian, &offset); err != nil {
			return nil, err
		}

		entries = append(entries, SparseEntry{
			Key:    string(keyBuf),
			Offset: offset,
		})

		currentPos += 8 + keySize + 8
	}

	return &SSTable{
		FilePath:    path,
		SparseIndex: entries,
	}, nil
}

func searchSSTable(ss *SSTable, key string) ([]byte, bool) {
	// Find closet SparseEntry <= key
	idx := sort.Search(len(ss.SparseIndex), func(i int) bool {
		return ss.SparseIndex[i].Key > key
	}) - 1

	// If first key of block is < key, it means key doesn't exist in this SSTable
	if idx < 0 {
		return nil, false
	}

	// Start lookup in SStable form offset
	file, _ := os.Open(ss.FilePath)
	defer file.Close()
	file.Seek(ss.SparseIndex[idx].Offset, io.SeekStart)

	// Start sequential scan until we find the key or reach the end of block
	for {
		// Read header 8 bytes [ keySize (4 bytes) ][ valueSize (4 bytes) ]
		header := make([]byte, 8)
		_, err := file.Read(header)
		if err == io.EOF {
			break
		}

		ks := binary.LittleEndian.Uint32(header[0:4]) // Parse header to get key size
		vs := binary.LittleEndian.Uint32(header[4:8]) // Parse header to get value size

		// Read key data
		kBuf := make([]byte, ks)
		file.Read(kBuf)

		currentKey := string(kBuf)
		if currentKey == key { // If key found we read data
			vBuf := make([]byte, vs)
			file.Read(vBuf)
			return vBuf, true
		}

		// current key > key so key doesn't exist (the sparse index already sort)
		if currentKey > key {
			break
		}
		file.Seek(int64(vs), io.SeekCurrent) // skip value if not match (move cursor to next header)
	}

	return nil, false
}
