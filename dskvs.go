/*

Package dskvs is a key value store.  In this store, there are two level or
mapping.  The store is organized in collections that hold members.  Each member
is represented by a page.  Each collection is represented by a map of members.

dskvs addresses members using a 'collection/member' convention.

To start using dskvs, create a Store object with dskvs.NewStore, specifying
where in the current filesystem to save the Store's files. Then, call store.Load
to load any pre-existing collections and/or members into your store instance.
When you're done using the store, call store.Close, which finishes writing the
last updates package dskvs

*/
package dskvs

import (
	"sync"
)

const (
	MAJOR_VERSION uint16 = 0
	MINOR_VERSION uint16 = 1
	PATCH_VERSION uint64 = 0
)

var (
	storeExistsLock sync.RWMutex
	storeExists     map[string]bool
	jan             janitor
)

func init() {
	storeExists = make(map[string]bool)
	jan = newJanitor()
}

// Store provides methods to manipulate the data held in memory and on disk at
// the path that was specified when you instantiated it.  Every store instance
// points at a different path location on disk.  Beware if you create a store
// that lives within the tree of another store.  There's no garantee to what
// will happen, aside perhaps a garantee that things will go wrong.
type Store struct {
	isLoaded    bool
	storagePath string
	coll        *collections
}

/*
	Meta operations on Store
*/

// NewStore instantiate a new store reading from the specified path
func NewStore(path string) (*Store, error) {

	if !isValidPath(path) {
		return nil, errorPathInvalid(path)
	}

	basepath := expandPath(path)
	return &Store{
		false,                    // isLoaded
		basepath,                 // storagePath
		newCollections(basepath), // collections
	}, nil
}

// Load loads the files in memory. This call will block for disk IO.
func (s *Store) Load() error {
	storeExistsLock.RLock()
	exists := storeExists[s.storagePath]
	storeExistsLock.RUnlock()

	if exists {
		return errorPathInUse(s.storagePath)
	}

	storeExistsLock.Lock()
	storeExists[s.storagePath] = true
	storeExistsLock.Unlock()

	err := jan.loadStore(s)
	if err != nil {
		return err
	}
	jan.run()

	s.isLoaded = true
	return nil
}

// Close finishes writing dirty updates and closes all the files. Report any
// error occuring doing so. This call will block for disk IO.
func (s *Store) Close() error {
	if !s.isLoaded {
		return errorStoreNotLoaded(s)
	}

	s.isLoaded = false

	storeExistsLock.Lock()
	delete(storeExists, s.storagePath)
	storeExistsLock.Unlock()

	err := jan.unloadStore(s)
	if err != nil {
		jan.die()
		return err
	}

	return nil
}

/*
	Storage operations
*/

// Get returns the value refernced by the `fullKey` given in argument. A
// `fullKey` is a string that has a collection identifier and a member
// identifier, separated by `CollKeySep`, Ex:
//
//	val, err := store.Get("artists/daft_punk")
//
// will get the value attached to Daft Punk, from within the Artists
// collection
func (s *Store) Get(fullKey string) ([]byte, error) {
	if !s.isLoaded {
		return nil, errorStoreNotLoaded(s)
	}

	if err := checkKeyValid(fullKey); err != nil {
		return nil, err
	}

	if isCollectionKey(fullKey) {
		return nil, errorGetIsColl(fullKey)
	}

	coll, key := splitKeys(fullKey)

	return s.coll.get(coll, key)
}

// GetAll returns all the members' value in the collection `coll`.
func (s *Store) GetAll(coll string) ([][]byte, error) {
	if !s.isLoaded {
		return nil, errorStoreNotLoaded(s)
	}

	if err := checkKeyValid(coll); err != nil {
		return nil, err
	}

	if !isCollectionKey(coll) {
		return nil, errorGetAllIsNotColl(coll)
	}

	return s.coll.getCollection(coll)
}

// Put saves the given value into the key location.  `fullKey` should be a
// member,  not a collection.  There is no `PutAll` version of this
// call.  If you wish to add a collection all at once, iterate over your
// collection and call `Put` on each member.
func (s *Store) Put(fullKey string, value []byte) error {
	if !s.isLoaded {
		return errorStoreNotLoaded(s)
	}

	if err := checkKeyValid(fullKey); err != nil {
		return err
	}

	if isCollectionKey(fullKey) {
		return errorPutIsColl(fullKey, string(value))
	}

	coll, key := splitKeys(fullKey)

	s.coll.put(coll, key, value)
	return nil
}

// Delete removes member with `fullKey` from the storage.
func (s *Store) Delete(fullKey string) error {
	if !s.isLoaded {
		return errorStoreNotLoaded(s)
	}

	if err := checkKeyValid(fullKey); err != nil {
		return err
	}

	if isCollectionKey(fullKey) {
		return errorDeleteIsColl(fullKey)
	}

	coll, key := splitKeys(fullKey)

	return s.coll.deleteKey(coll, key)
}

// DeleteAll removes all the members in collection `coll`
func (s *Store) DeleteAll(coll string) error {
	if !s.isLoaded {
		return errorStoreNotLoaded(s)
	}

	if err := checkKeyValid(coll); err != nil {
		return err
	}

	if !isCollectionKey(coll) {
		return errorDeleteAllIsNotColl(coll)
	}

	return s.coll.deleteCollection(coll)
}
