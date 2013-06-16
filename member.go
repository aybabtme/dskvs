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
		new(sync.RWMutex),
		make(map[string]*page),
	}
}

func (m *member) get(key string) (*string, error) {
	m.RLock()
	aPage, ok := m.entries[key]
	m.RUnlock()

	if !ok {
		return nil, errorNoSuchKey(key)
	}

	aPage.RLock()
	defer aPage.RUnlock()
	return aPage.value, nil
}

func (m *member) getMembers() ([]*string, error) {

}

func (m *member) put(key string, value *string) {
	m.Lock()
	defer m.Unlock()
	aPage, ok := m.entries[key]
	if !ok {
		aPage := newPage()
	}
	aPage.set(value)
	m[key] = aPage
}

func (m *member) delete(key string) {
	m.Lock()
	defer m.Unlock()
	aPage, ok := m.entries[key]
	if ok {
		aPage.delete()
		m[key] = nil
	}
}

func (m *member) deleteAll() error {
	m.Lock()
	for _, aPage := range m.entries {
		aPage.delete()
	}
	m.Unlock()
}
