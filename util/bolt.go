package util

import (
	"sync"
	"time"

	"github.com/boltdb/bolt"
)

var (
	DefaultBucket = []byte("chia")
	DefaultDB     = "/etc/chia-storage-server.db"
)

var (
	dbpath     string
	boltClient sync.Map
	lock       sync.Mutex
)

func InitBoltClient(path string) {
	if path == "" {
		path = DefaultDB
	}
	dbpath = path
}

func BoltClient() (*bolt.DB, error) {
	return _boltClient(dbpath)
}

func _boltClient(path string) (*bolt.DB, error) {
	// Open the my.db data file in your current directory.
	// It will be created if it doesn't exist.
	if db, ok := boltClient.Load("bolt"); ok {
		return db.(*bolt.DB), nil
	}

	lock.Lock()
	defer lock.Unlock()
	db, err := open(path)
	if err != nil {
		return nil, err
	}

	boltClient.Store("bolt", db)
	return db, nil
}

// TODO
func Close() error {
	return nil
}

func open(path string) (*bolt.DB, error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 10 * time.Second})
	if err != nil {
		return nil, err
	}
	if err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(DefaultBucket)
		return err
	}); err != nil {
		return nil, err
	}

	return db, nil
}
