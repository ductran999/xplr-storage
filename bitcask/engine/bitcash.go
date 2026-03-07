package engine

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"
	"os"
	"sync"
	"time"
)

// IndexEntry store metadata about a key's location in the file
type IndexEntry struct {
	FileID    int
	Offset    int64
	Size      int
	Timestamp int64
}

// Bitcask engine
type Bitcask struct {
	mu         sync.RWMutex
	index      map[string]IndexEntry
	writeFile  *os.File
	currOffset int64
}

type BitcaskIterator struct {
	db      *Bitcask
	keys    []string
	currIdx int
	currKey string
	currVal []byte
	err     error
}

func Open(path string) (*Bitcask, error) {
	f, _ := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)

	bc := &Bitcask{
		index:     make(map[string]IndexEntry),
		writeFile: f,
	}

	// Scan file to build in-memory index
	var offset int64
	for {
		// Read header to get key/value size
		header := make([]byte, 20)
		_, err := f.ReadAt(header, offset)
		if err == io.EOF {
			break
		}

		ks := binary.LittleEndian.Uint32(header[12:16])
		vs := binary.LittleEndian.Uint32(header[16:20])
		recordSize := int64(20 + ks + vs)

		// Read key and add to map
		keyBuf := make([]byte, ks)
		f.ReadAt(keyBuf, offset+20)

		bc.index[string(keyBuf)] = IndexEntry{
			Offset: offset,
			Size:   int(recordSize),
		}

		offset += recordSize
	}
	bc.currOffset = offset

	return bc, nil
}

func (b *Bitcask) Put(key string, value []byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Encode entry: [Header] + Key + Value
	// [Header: CRC, TS, KS, VS] + Key + Value
	entry := b.encodeEntry(key, value)
	entrySize := len(entry)

	// Append to file (Write-ahead log style)
	n, err := b.writeFile.Write(entry)
	if err != nil {
		return err
	}

	// Update in-memory index
	b.index[key] = IndexEntry{
		Offset:    b.currOffset,
		Size:      entrySize,
		Timestamp: now(),
	}

	// Move offset for next write
	b.currOffset += int64(n)

	return nil
}

func (b *Bitcask) ListKeys() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	keys := make([]string, 0, len(b.index))
	for k := range b.index {
		keys = append(keys, k)
	}

	return keys
}

func (b *Bitcask) Get(key string) ([]byte, error) {
	b.mu.RLock()
	entry, ok := b.index[key]
	b.mu.RUnlock()

	if !ok {
		return nil, errors.New("key not found")
	}

	buf := make([]byte, entry.Size)
	_, err := b.writeFile.ReadAt(buf, entry.Offset)
	if err != nil {
		return nil, err
	}

	_, value, err := b.decodeEntry(buf)
	return value, err
}

func (db *Bitcask) Iterator() *BitcaskIterator {
	db.mu.RLock()
	defer db.mu.RUnlock()

	keys := make([]string, 0, len(db.index))
	for k := range db.index {
		keys = append(keys, k)
	}
	return &BitcaskIterator{keys: keys, db: db, currIdx: -1}
}

func (it *BitcaskIterator) Next() bool {
	it.currIdx++
	if it.currIdx >= len(it.keys) {
		return false
	}

	it.currKey = it.keys[it.currIdx]
	it.currVal = nil
	return true
}

func (it *BitcaskIterator) Key() string {
	return it.currKey
}

func (it *BitcaskIterator) Value() []byte {
	if it.currVal == nil {
		val, err := it.db.Get(it.currKey)
		if err != nil {
			it.err = err
			return nil
		}
		it.currVal = val
	}
	return it.currVal
}

func (it *BitcaskIterator) Error() error {
	return it.err
}

func (b *Bitcask) Close() error {
	return b.writeFile.Close()
}

func (b *Bitcask) encodeEntry(key string, value []byte) []byte {
	ks := uint32(len(key))
	vs := uint32(len(value))
	timestamp := uint64(time.Now().Unix())

	// Calculate total size: 4(CRC) + 8(TS) + 4(KS) + 4(VS) + len(K) + len(V)
	size := 4 + 8 + 4 + 4 + ks + vs
	buf := make([]byte, size)

	// Byte 4-12: Timestamp
	binary.LittleEndian.PutUint64(buf[4:12], timestamp)
	// Byte 12-16: KeySize
	binary.LittleEndian.PutUint32(buf[12:16], ks)
	// Byte 16-20: ValueSize
	binary.LittleEndian.PutUint32(buf[16:20], vs)

	// Byte 20 -> 20+ks: Key
	copy(buf[20:20+ks], key)
	// Byte 20+ks -> end of Value
	copy(buf[20+ks:], value)

	// Calculate CRC32 from byte 4 to end of entry
	// (CRC does not include itself)
	checkSum := crc32.ChecksumIEEE(buf[4:])
	binary.LittleEndian.PutUint32(buf[0:4], checkSum)

	return buf
}

func (b *Bitcask) decodeEntry(data []byte) (string, []byte, error) {
	// Check CRC32
	savedCRC := binary.LittleEndian.Uint32(data[0:4])
	actualCRC := crc32.ChecksumIEEE(data[4:])
	if savedCRC != actualCRC {
		return "", nil, errors.New("data corruption: CRC mismatch")
	}

	ks := binary.LittleEndian.Uint32(data[12:16])

	key := string(data[20 : 20+ks])
	value := data[20+ks:]

	return key, value, nil
}

func now() int64 {
	return time.Now().Unix()
}
