package dskvs

func put(coll, key, value string) {
	collectionLock.RLock()
	members := collections[coll]
	collectionLock.RUnlock()

	var aPage *page
	if members == nil {
		aPage = newPage()
	}

	aPage.lock.Lock()
	wasDirty := aPage.isDirty
	aPage.isDirty = true
	aPage.value = value
	if wasDirty {
		dirtyPages <- aPage
	}
	aPage.Unlock()

	collectionLock.Lock()
	collections[coll][key] = aPage
	collectionLock.Unlock()
}
