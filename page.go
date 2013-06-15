package dskvs

import (
	"sync"
)

type page struct {
	lock      sync.RWMutex
	isDirty   bool
	isDeleted bool
	value     *string
}

func newPage(value string) *page {
	return &page{
		new(sync.RWMutex),
		false,
		false,
		value,
	}
}

func (p *page) get() *string {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.value
}

func (p *page) set(value *string) {
	p.lock.Lock()
	defer p.lock.Unlock()
	wasDirty := p.isDirty
	p.value = value
	p.isDirty = true
	if !wasDirty {
		dirtyPages <- p
	}
}

func (p *page) delete() {
	p.lock.Lock()
	defer p.lock.Unlock()
	wasDirty := p.isDirty
	p.value = nil
	p.isDirty = true
	p.isDeleted = true
	if !wasDirty {
		dirtyPages <- p
	}
}
