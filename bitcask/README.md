# Mini Bitcask

Mini Bitcask is a simple implementation of the Bitcask Key-Value Store. It is designed for educational purposes and is not intended for production use.

> **Bitcask philosophy: "Append-only on Disk, Hash-map in RAM"**

---

# Architecture

Mini Bitcask follows the core Bitcask design:

```sh
             +------------------+
PUT/GET ---> |  In-memory Index |
             |    (Hash Map)    |
             +--------+---------+
                      |
                      v
             +------------------+
             | Append-only Log  |
             | data.db          |
             +------------------+
```

### Write Path

1. New key-value entries are **appended to disk**
2. In-memory index is **updated with latest offset**

### Read Path

1. Lookup key in **in-memory index**
2. Use stored **file offset**
3. Read value directly from disk

---

# Features

- Append-only storage engine
- In-memory hash index
- Fast `GET` operations (O(1))
- CLI interface for interacting with the database
- Interactive shell mode
- Iterator support for scanning records
- Structured logging using `slog`

---

# Running the CLI

Change working directory to `bitcask`

```bash
cd bitcask
```

Build the CLI application

```makefile
make build
```

Feed mock data into database

```bash
./bin/bitcask-cli feed
```

Start interactive shell

```bash
./bin/bitcask-cli connect
```

In the shell, you can execute commands like:

```bash
# to list keys
list 10 20

# get value of key
get <KEY>

# set value with key
put <KEY> <VALUE>
```

# Limitations

This implementation is simplified and does **not include**:

- Compaction / merge process
- Crash recovery optimization
- Hint files
- File rotation
- Background merge

These features exist in the original Bitcask design.

---

# Learning Goals

This project helps explore:

- Log-structured storage engines
- Disk I/O patterns
- Key-value indexing
- Iterators
- Simple database CLI design

---

# License

MIT License.
