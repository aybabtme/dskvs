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
	defer p.RUnlock()
	if p.isDeleted {
		return nil
	}
	data := make([]byte, len(p.value))
	copy(data, p.value)
	return data
}

func (p *page) set(value []byte) {
	data := make([]byte, len(value))
	copy(data, value)
	p.Lock()
	defer p.Unlock()
	wasDirty := p.isDirty
	p.value = data
	p.isDirty = true
	if !wasDirty {
		jan.DirtyPages <- p
	}
}

func (p *page) delete() {
	p.Lock()
	defer p.Unlock()
	wasDirty := p.isDirty
	p.value = nil
	p.isDirty = true
	p.isDeleted = true
	if !wasDirty {
		jan.DirtyPages <- p
	}
}
