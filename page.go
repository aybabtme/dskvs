package dskvs

import (
	"sync"
)

type page struct {
	isDirty   bool
	isDeleted bool
	basepath  string
	coll      string
	key       string
	value     []byte
	sync.RWMutex
}

func newPage(basepath, coll, key string) *page {
	return &page{
		isDirty:   false,
		isDeleted: false,
		basepath:  basepath,
		coll:      coll,
		key:       key,
		value:     nil,
	}
}

func (p *page) get() []byte {
	p.RLock()
	if p.isDeleted {
		return nil
	}
	p.RUnlock()
	return p.value
}

func (p *page) set(value []byte) {
	p.Lock()
	newBytes := make([]byte, len(value))
	copy(newBytes, value)
	p.value = newBytes
	wasDirty := p.isDirty
	p.isDirty = true
	p.Unlock()
	if !wasDirty {
		jan.writePage(p)
	}
}

func (p *page) delete() {
	p.Lock()
	wasDirty := p.isDirty
	p.value = nil
	p.isDirty = true
	p.isDeleted = true
	p.Unlock()
	if !wasDirty {
		jan.writePage(p)
	}
}
