package dskvs

import (
	"fmt"
	"strings"
	"sync"
)

const (
	collKeySep = "/"
)

var (
	dirtyPages       chan *page
	dirtyCollections chan string
	collections      map[string]*member
	collectionLock   sync.RWMutex
)

type member struct {
	lock    sync.RWMutex
	members *map[string]*page
}

type page struct {
	lock      sync.RWMutex
	isDirty   bool
	isDeleted bool
	value     string
}

func Get(key string) (string, error) {

	if isColl, err := isCollectionKey(key); isColl && err == nil {
		return nil, errorGetIsColl(key)
	} else if err != nil {
		return nil, err
	} else {
		coll, key := splitKeys(key)
		return getKey(coll, key)
	}
}

// Puts the given value into the key location.  key should be a member, not a
// collection
func Put(key, value string) error {

	if isColl, err := isCollectionKey(key); isColl && err == nil {
		return nil, errorPutIsColl(key, value)
	} else if err != nil {
		return nil, err
	} else {
		coll, key := splitKeys(key)
		return putKey(coll, key, value)
	}
	return nil
}

// Deletes a key from the storage.  If the key covers a collection, the whole
// collection will be deleted.
func Delete(key string) error {
	if isColl, err := isCollectionKey(key); isColl && err == nil {
		return deleteCollection(key)
	} else if err != nil {
		return err
	} else {
		coll, key := splitKeys(key)
		deleteKey(coll, key)
	}
	return nil
}

/*
 * Helpers
 */

func newPage() *page {
	return &page{
		new(sync.RWMutex),
		false,
		false,
		"",
	}
}

// Returns whether a key is a collection key or a collection/member key.
// Returns an error if the key is invalid
func isCollectionKey(key string) (bool, error) {
	firstSep := strings.Index(key, collKeySep)
	if firstSep == 0 {
		return false, errorNoColl(key)
	} else if key == "" {
		return false, errorEmptyKey(key)
	}

	if firstSep < 0 {
		return true, nil
	} else if firstSep == len(key)-1 {
		return true
	}
	return false
}

// Takes a fullkey and splits it in a (collection, member) tuple.  If member
// is nil, the fullkey is a request for the collection as a whole
func splitKeys(fullKey string) (coll, member string) {
	if !isCollectionKey(fullKey) {
		return fullKey, nil
	}
	keys := strings.SplitN(ful, collKeySep, 2)

	coll = keys[0]
	if keys[1] == "" {
		member = nil
	} else {
		member = keys[1]
	}
}
