package dskvs

import (
	"sync"
)

// A member is a map protected by a RW lock to prevent concurrent
// modificiations
type member struct {
	basepath string
	coll     string
	entries  map[string]*page
	sync.RWMutex
}

func newMember(basepath, coll string) *member {
	return &member{
		basepath: basepath,
		coll:     coll,
		entries:  make(map[string]*page),
	}
}

func (m *member) get(key string) ([]byte, error) {
	m.RLock()
	aPage, ok := m.entries[key]
	m.RUnlock()

	if !ok {
		return nil, errorNoSuchKey(key)
	}

	return aPage.get(), nil
}

func (m *member) getMembers() [][]byte {
	// This is tricky because we don't want to lock the whole map for
	// reading while doing this query.  We don't want that for two reasons:
	// 1. If there are many ongoing writes on the pages, chances are we will
	//	  not be able to proceed with our read-lock.
	// 2. If there are writes following our read, they will have to wait
	//    for the duration of our procedure.
	// However, we still want consistent results. The solution follows:

	// Lock the map for read
	m.RLock()
	var pages []*page
	// Get a snapshot of the pages
	for _, aPage := range m.entries {
		pages = append(pages, aPage)
	}
	// Release the lock, everybody is free to use the map again
	m.RUnlock()

	// Now we want to grab the bytes in the pages
	var values [][]byte
	var aVal []byte
	for _, aPage := range pages {
		// As we iterate over every page, other goroutines can do so as well.
		// Since the map is not locked, page[i+1] could be deleted while we
		// read page[i]. In such a case, when we get to read page[i+1], we'll
		// want to discard this value.
		aVal = aPage.get()
		// If the page has been deleted, page.get() contracts says it will
		// return a nil-slice
		if aVal == nil {
			// So we can discard those deleted bytes
		} else {
			// And keep only those that are still valid
			values = append(values, aVal)
		}
	}
	// That's all folks!
	return values
}

func (m *member) put(key string, value []byte) {

	// We'd rather not write-lock the whole map if we don't need to
	m.RLock()
	aPage, ok := m.entries[key]
	m.RUnlock()

	// But if the entry doesn't exist, we need to write one
	if !ok {
		m.Lock()
		// Before writing the entry, we verify that is was not added since
		// last read
		aPage, ok = m.entries[key]
		if !ok {
			// It was not so go ahead and write a new entry
			aPage = newPage(m.basepath, m.coll, key)
			m.entries[key] = aPage
		}
		m.Unlock()
	}

	// Operate on the page itself, which holds a more granular lock
	aPage.set(value)
}

func (m *member) delete(key string) {
	// If the page is already deleted, we don't waste time
	m.RLock()
	aPage, ok := m.entries[key]
	m.RUnlock()

	if ok {
		// Delete the page from the entries first
		m.Lock()
		delete(m.entries, key)
		m.Unlock()
		// Then let the page delete itself on a more granular lock
		aPage.delete()
	}
}

func (m *member) deleteAll() {
	// In this case, it makes sense to just lock the whole map :
	// we're deleting everything...
	m.Lock()
	for _, aPage := range m.entries {
		aPage.delete()
	}
	m.Unlock()
}
