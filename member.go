package dskvs

import (
	"sync"
)

// A member is a map protected by a RW lock to prevent concurrent modificiations
type member struct {
	lock    sync.RWMutex
	members *map[string]*page
}

func newMember() *member {
	return &member{
		new(sync.RWMutex),
		make(map[string]*page),
	}
}

func (m *member) get(key string) *page {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m[key]
}

func (m *member) put(key, value string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m[key] = newPage(value)
}

func (m *member) delete(key string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	aPage := m[key]
	aPage.delete()
	m[key] = nil
}
