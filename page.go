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
	return p.value
}

func (p *page) set(value []byte) {
	p.Lock()
	defer p.Unlock()
	wasDirty := p.isDirty
	p.value = value
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
