// dskvs is a key value store.  In this store, there are two level or
// mapping.  The store is organized in collections that hold members.
// Each member is represented by a page.  Each collection is represented
// by a map of members.
//
// dskvs addresses members using a 'collection/member' convention.
//
// To start using dskvs, create a Store object with dskvs.NewStore,
// specifying where in the current filesystem to save the Store's files.
// Then, call store.Load to load any pre-existing collections and/or
// members into your store instance.  When you're done using the store,
// call store.Close, which finishes writing the last updates
package dskvs

import (
	"fmt"
	"strings"
	"sync"
)

const collKeySep = "/"

var existingStore map[string]bool

func init() {
	existingStore := make(map[string]bool)
}

type Store struct {
	isLoaded     bool
	storagePath  string
	dirtyPages   chan *page
	dirtyMembers chan *member
	coll         collections
}

// Instantiate a new store reading from the specified path
func NewStore(path string) (*Store, error) {

	if isInvalidPath(path) {
		return nil, errorPathInvalid(path)
	}

	return &Store{path, make(chan *page), make(chan *members)}, nil
}

// This call will block for disk IO.
// Loads the files in memory.
func (s *Store) Load() error {
	_, ok := existingStore[s.storagePath]

	if ok {
		return errorPathInUse(s.storagePath)
	}

	existingStore[s.storagePath] = true

	// TODO scan the path for files, load them in memory
}

// This call will block for disk IO.
// Finish writing dirty updates and close all the files. Report any
// error occuring doing so.
func (s *Store) Close() error {
	if !s.isLoaded {
		return errorStoreNotLoaded(s)
	}

	existingStore[s.storagePath]
}

func (s *Store) Get(key string) (*string, error) {
	if !s.isLoaded {
		return nil, errorStoreNotLoaded(s)
	}

	isColl, err := isCollectionKey(key)
	if err != nil {
		return nil, err
	}

	if isColl {
		return nil, errorGetIsColl(key)
	}

	coll, key := splitKeys(key)
	return s.coll.get(coll, key)
}

// Gets all the members' value in the collection `coll`.
func (s *Store) GetAll(coll string) ([]*string, error) {
	if !s.isLoaded {
		return nil, errorStoreNotLoaded(s)
	}

	isColl, err := isCollectionKey(coll)
	if err != nil {
		return nil, err
	}

	if !isColl {
		return nil, errorGetAllIsNotColl(coll)
	}

	return s.coll.getCollection(coll)
}

// Puts the given value into the key location.  `key` should be a member,
// not a collection.  There is no `PutAll` version of this call.  If you
// wish to add a collection all at once, iterate over your collection and
// call `Put` on each member.
func (s *Store) Put(key string, value *string) error {
	if !s.isLoaded {
		return errorStoreNotLoaded(s)
	}

	isColl, err := isCollectionKey(key)
	if err != nil {
		return err
	}

	if isColl {
		return errorPutIsColl(key, value)
	}

	coll, key := splitKeys(key)
	return s.coll.put(coll, key, value)
}

// Deletes member with `key` from the storage.
func (s *Store) Delete(key string) error {
	if !s.isLoaded {
		return errorStoreNotLoaded(s)
	}

	isColl, err := isCollectionKey(key)
	if err != nil {
		return err
	}

	if isColl {
		return errorDeleteIsColl(key)
	}

	coll, key := splitKeys(key)
	return s.coll.deleteKey(coll, key)
}

// Deletes all the members in collection `coll`
func (s *Store) DeleteAll(coll string) error {
	if !s.isLoaded {
		return nil, errorStoreNotLoaded(s)
	}

	isColl, err := isCollectionKey(coll)
	if err != nil {
		return err
	}

	if !isColl {
		return errorDeleteAllIsNotColl(coll)
	}

	return s.coll.deleteCollection(coll)
}
