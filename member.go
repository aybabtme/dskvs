package dskvs

import (
	"sync"
)

// A member is a map protected by a RW lock to prevent concurrent
// modificiations
type member struct {
	sync.RWMutex
	entries map[string]*page
}

func newMember() *member {
	return &member{
		entries: make(map[string]*page),
	}
}

func (m *member) get(key string) (*string, error) {
	m.RLock()
	aPage, ok := m.entries[key]
	m.RUnlock()

	if !ok {
		return nil, errorNoSuchKey(key)
	}

	return aPage.get(), nil
}

func (m *member) getMembers() []*string {
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
	defer m.RUnlock()

	// Now we want to grab the strings in the pages
	var values []*string
	var aVal *string
	for _, aPage := range pages {
		// As we iterate to get every page, since the map is not locked,
		// page[i+1] can be deleted while we read page[i], and when we get
		// to read page[i+1], we'll want to discard this value.  We don't
		// want to return deleted pages as if they were part of our
		// collection
		aVal = aPage.get()
		// If the page has been delete, page.get() contracts says it will
		// return a nil-string
		if aVal == nil {
			// So we can discard those deleted strings
		} else {
			// And keep only those that are still valid
			values = append(values, aVal)
		}
	}
	// That's all folks!
	return values
}

func (m *member) put(key string, value *string) {
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
			aPage := newPage(key)
			m.entries[key] = aPage
		}
		m.RUnlock()
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
