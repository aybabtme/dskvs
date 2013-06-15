package dskvs

func deleteCollection(coll string) error {
	collectionLock.Lock()
	collections[coll] = nil
	dirtyCollections <- coll
	collectionLock.Unlock()
	return nil
}

func deleteKey(coll, key string) {
	collectionLock.RLock()
	page := collections[coll][key]
	collectionLock.RUnlock()

	page.lock.Lock()
	wasDirty := page.isDirty
	page.isDeleted = true
	page.isDirty = true
	page.value = nil

	if !wasDirty {
		dirtyPages <- page
	}
	page.lock.Unlock()
}
