package dskvs

import (
	"sync"
)

// A member is a map protected by a RW lock to prevent concurrent
// modificiations
type members struct {
	lock    sync.RWMutex
	entries *map[string]*page
}

func newMember() *member {
	return &member{
		new(sync.RWMutex),
		make(map[string]*page),
	}
}

func (m *member) get(key string) (*string, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	val, ok := m.entries[key]

	if ok {
		return val.value, nil
	} else
		return nil, errorNoSuchKey(key)
	}
}

func (m *member) put(key string, value *string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	aPage, ok := m.entries[key]
	if !ok {
		aPage := newPage()
	}
	aPage.set(value)
	m[key] = aPage
}

func (m *member) delete(key string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	aPage, ok := m.entries[key]
	if ok {
		aPage.delete()
		m[key] = nil
	}
}
