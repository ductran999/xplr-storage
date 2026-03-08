package engine

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/btree"
)

type LSMTree struct {
	mu sync.RWMutex

	dataDir  string
	walPath  string
	wal      *WAL
	sstables []*SSTable

	threshold int // Threshold to flush to disk
	memtable  *btree.BTree
	memSize   int
}

func NewLSMTree(dataDir string, threshold int) (*LSMTree, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	// If not found any wal file, create new one
	walPath := filepath.Join(dataDir, "wal_current.log")
	wal, err := NewWAL(walPath)
	if err != nil {
		return nil, fmt.Errorf("init wal error: %w", err)
	}

	// Read data from wal to reconstruct memtable
	memtable := btree.New(32)
	items, err := wal.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("recover wal error: %w", err)
	}
	for k, v := range items {
		memtable.ReplaceOrInsert(Item{Key: k, Value: v})
	}
	log.Printf("[WAL]: Recovered %d items\n", len(items))

	// Discover existing sstables on disk
	loadedSSTables, err := discoverSSTables(dataDir)
	if err != nil {
		return nil, err
	}

	return &LSMTree{
		wal:       wal,
		memtable:  memtable,
		threshold: threshold,
		walPath:   walPath,
		dataDir:   dataDir,
		sstables:  loadedSSTables,
	}, nil
}

func (lsm *LSMTree) Put(key string, value []byte) error {
	// Lock for writing
	lsm.mu.Lock()
	defer lsm.mu.Unlock()

	// Write to WAL first for durability
	if err := lsm.wal.Append([]byte(key), value); err != nil {
		return fmt.Errorf("failed to write to WAL: %v", err)
	}
	log.Println("[WAL]: append success for key:", key)

	// Insert to memtable
	lsm.memtable.ReplaceOrInsert(Item{Key: key, Value: value})
	lsm.memSize += len(key) + len(value)
	log.Println("[Memtable]: put key", key, "memtable size", lsm.memSize)

	if lsm.memSize >= lsm.threshold {
		log.Println("[Memtable]: full, flushing to SSTable...")
		if err := lsm.flush(); err != nil {
			return err
		}
	}

	return nil
}

func (lsm *LSMTree) Get(key string) ([]byte, bool) {
	lsm.mu.RLock()
	defer lsm.mu.RUnlock()

	if val := lsm.memtable.Get(Item{Key: key}); val != nil {
		fmt.Println(">>> key:", key, "found in Memtable")
		return val.(Item).Value, true
	}

	for _, ss := range lsm.sstables {
		val, found := searchSSTable(ss, key)
		if found {
			fmt.Println(">>> key:", key, "found in SSTable", ss.FilePath)
			return val, true
		}
	}

	return nil, false
}

func (lsm *LSMTree) flush() error {
	// Create new SStable file
	fileName := fmt.Sprintf("%s/sstable_%d.db", lsm.dataDir, time.Now().UnixNano())
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	var sparseIndex []SparseEntry
	var currentOffset int64 = 0
	count := 0

	// Iterate the memtable write them to file and build sparseIndex
	var writeErr error
	lsm.memtable.Ascend(func(i btree.Item) bool {
		item := i.(Item)

		// create sparse index, e.g every 5 items (granularity ~ 5)
		if count%5 == 0 {
			sparseIndex = append(sparseIndex, SparseEntry{Key: item.Key, Offset: currentOffset})
		}

		buf := make([]byte, 8) // 8 bytes Header [ Key Size (4bytes) ] + [ Value Size (4 bytes) ]
		binary.LittleEndian.PutUint32(buf[0:4], uint32(len(item.Key)))
		binary.LittleEndian.PutUint32(buf[4:8], uint32(len(item.Value)))

		// Write header
		if _, err := file.Write(buf); err != nil {
			writeErr = err
			return false
		}
		// Write key
		if _, err := file.Write([]byte(item.Key)); err != nil {
			writeErr = err
			return false
		}
		// Write value
		if _, err := file.Write(item.Value); err != nil {
			writeErr = err
			return false
		}

		currentOffset += int64(8 + len(item.Key) + len(item.Value))
		count++
		return true
	})

	if writeErr != nil {
		return writeErr
	}

	// Write sparse index by format [keySize][key][offset]
	indexStartOffset := currentOffset
	for _, entry := range sparseIndex {
		binary.Write(file, binary.LittleEndian, int64(len(entry.Key)))
		file.WriteString(entry.Key)
		binary.Write(file, binary.LittleEndian, entry.Offset)
	}

	// Write footer (start index offset)
	binary.Write(file, binary.LittleEndian, indexStartOffset)
	if err := file.Sync(); err != nil {
		return fmt.Errorf("fsync to disk failed: %w", err)
	}

	// Rotate WAL
	lsm.wal.Close()
	os.Rename(lsm.walPath, fmt.Sprintf("%s/wal_%d.log", lsm.dataDir, time.Now().UnixNano()))

	newWalPath := fmt.Sprintf("%s/wal_current.log", lsm.dataDir)
	lsm.wal, _ = NewWAL(newWalPath)
	lsm.walPath = newWalPath

	// Move SSTable to latest and reset memtable
	lsm.sstables = append([]*SSTable{{FilePath: fileName, SparseIndex: sparseIndex}}, lsm.sstables...)
	lsm.memtable.Clear(false)
	lsm.memSize = 0

	return nil
}
