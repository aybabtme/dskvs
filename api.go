package dskvs

import (
	"path/filepath"
	"sync"
)

var (
	// CollKeySep is the value used by dskvs to separate the
	// collection part of the full key from the member part.  This is usually
	// '/' on Unix systems.
	CollKeySep = string(filepath.Separator)

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
	storagePath string
	coll        *collections
}

/*
	Meta operations on Store
*/

// Load retrieves previously persisted entries and load them in memory.
// Every file is checked for consistency with a SHA1 checksum.  A file that
// is not consistent will be ignored, a log message emitted and an error
// returned.  This call will block until all collections have been replenished.
func Open(path string) (*Store, error) {

	if !isValidPath(path) {
		return nil, errorPathInvalid(path)
	}

	basepath := expandPath(path)

	storeExistsLock.RLock()
	exists := storeExists[basepath]
	if exists {
		storeExistsLock.RUnlock()
		return nil, errorPathInUse(basepath)
	}
	storeExistsLock.RUnlock()

	storeExistsLock.Lock()
	if !exists && storeExists[basepath] {
		storeExistsLock.Unlock()
		return nil, errorPathInUse(basepath)
	}
	storeExists[basepath] = true
	storeExistsLock.Unlock()

	s := &Store{basepath, newCollections(basepath)}

	err := jan.loadStore(s)
	if err != nil {
		return nil, err
	}
	jan.run()

	return s, nil

}

// Close finishes writing dirty updates and closes all the files. It reports
// any error that occured doing so. This call will block until the writes are
// completed.
func (s *Store) Close() error {

	if s == nil {
		return errorStoreNotLoaded()
	}

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

// Get returns the value referenced by the `fullKey` given in argument. A
// `fullKey` is a string that has a collection identifier and a member
// identifier, separated by `CollKeySep`, Ex:
//
//	val, err := store.Get("artists/daft_punk")
//
// will get the value attached to Daft Punk, from within the Artists
// collection
//
// ATTENTION : do not modify the value of the slices that are returned to
// you.
func (s *Store) Get(fullKey string) ([]byte, error) {

	if err := checkKeyValid(fullKey); err != nil {
		return nil, err
	}

	if isCollectionKey(fullKey) {
		return nil, errorGetIsColl(fullKey)
	}

	coll, key := splitKeys(fullKey)

	return s.coll.get(coll, key)
}

// GetAll returns all the members in the collection `coll`.
//
// ATTENTION : do not modify the value of the slices that are returned to
// you.
func (s *Store) GetAll(coll string) ([][]byte, error) {

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

	if err := checkKeyValid(coll); err != nil {
		return err
	}

	if !isCollectionKey(coll) {
		return errorDeleteAllIsNotColl(coll)
	}

	return s.coll.deleteCollection(coll)
}
