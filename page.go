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
	data := make([]byte, len(p.value))
	copy(data, p.value)
	p.RUnlock()
	return data
	//return p.value
}

func (p *page) set(value []byte) {
	p.Lock()
	p.value = make([]byte, len(value))
	copy(p.value, value)
	wasDirty := p.isDirty
	p.isDirty = true
	p.Unlock()
	if !wasDirty {
		jan.DirtyPages <- p
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
		jan.DirtyPages <- p
	}
}
