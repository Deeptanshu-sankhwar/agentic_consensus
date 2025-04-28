package core

import (
	"bytes"
	"encoding/json"
	"sync"
)

type Database struct {
	buffer map[string]*bytes.Buffer
	mu     sync.RWMutex
}

// Creates a new in-memory database instance
func NewDatabase() *Database {
	return &Database{
		buffer: make(map[string]*bytes.Buffer),
	}
}

// Stores a value in the database with the given key
func (db *Database) Set(key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	if db.buffer[key] == nil {
		db.buffer[key] = &bytes.Buffer{}
	}
	db.buffer[key].Reset()
	db.buffer[key].Write(data)
	return nil
}

// Retrieves a value from the database using the given key
func (db *Database) Get(key string, value interface{}) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if buf := db.buffer[key]; buf != nil {
		return json.NewDecoder(bytes.NewReader(buf.Bytes())).Decode(value)
	}
	return nil
}

// Removes a key and its associated value from the database
func (db *Database) Delete(key string) {
	db.mu.Lock()
	defer db.mu.Unlock()
	delete(db.buffer, key)
}
