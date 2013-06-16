package dskvs

import (
	"sync"
)

type page struct {
	sync.RWMutex
	isDirty   bool
	isDeleted bool
	value     *string
}

func newPage() *page {
	return &page{
		isDirty:   false,
		isDeleted: false,
		value:     nil,
	}
}

func (p *page) get() *string {
	p.RLock()
	defer p.RUnlock()
	if p.isDeleted {
		return nil
	}
	return p.value
}

func (p *page) set(value *string) {
	p.Lock()
	defer p.Unlock()
	wasDirty := p.isDirty
	p.value = value
	p.isDirty = true
	if !wasDirty {
		dirtyPages <- p
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
		dirtyPages <- p
	}
}