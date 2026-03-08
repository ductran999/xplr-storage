package engine

import (
	"encoding/binary"
	"hash/crc32"
	"io"
	"os"
	"sync"
)

type WAL struct {
	file *os.File
	mu   sync.Mutex
}

// Record structure: [CRC(4)][KeySize(4)][ValueSize(4)][Key][Value]
func NewWAL(path string) (*WAL, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	return &WAL{file: f}, nil
}

func (w *WAL) Append(key []byte, value []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 1. Calculate CRC32 Key + Value
	checksum := crc32.ChecksumIEEE(append(key, value...))

	// 2. Create buffer for writing (4+4+4 = 12 bytes header)
	buf := make([]byte, 12+len(key)+len(value))
	binary.LittleEndian.PutUint32(buf[0:4], checksum)
	binary.LittleEndian.PutUint32(buf[4:8], uint32(len(key)))
	binary.LittleEndian.PutUint32(buf[8:12], uint32(len(value)))
	copy(buf[12:], key)
	copy(buf[12+len(key):], value)

	// 3. Write to disk
	if _, err := w.file.Write(buf); err != nil {
		return err
	}

	// 4. Important: Ensure data is flushed to disk
	return w.file.Sync()
}

func (w *WAL) ReadAll() (map[string][]byte, error) {
	data := make(map[string][]byte)
	w.file.Seek(0, 0)

	for {
		header := make([]byte, 12)
		_, err := w.file.Read(header)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		checksum := binary.LittleEndian.Uint32(header[0:4])
		kSize := binary.LittleEndian.Uint32(header[4:8])
		vSize := binary.LittleEndian.Uint32(header[8:12])

		record := make([]byte, kSize+vSize)
		_, err = io.ReadFull(w.file, record)
		if err != nil || crc32.ChecksumIEEE(record) != checksum {
			break // File got corrupted
		}

		key := string(record[:kSize])
		data[key] = record[kSize:]
	}

	return data, nil
}

func (w *WAL) Size() (int64, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	info, err := w.file.Stat()
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func (w *WAL) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.file.Close()
}
