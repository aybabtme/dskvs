package dskvs

import (
	"sync"
)

type collections struct {
	sync.RWMutex
	members map[string]*member
}

func newCollections() *collections {
	return &collections{
		new(sync.RWMutex),
		make(map[string]*member),
	}
}

func (c *collections) get(coll, key string) (*string, error) {
	c.RLock()
	m, ok := c.members[coll]
	c.RUnlock()
	if !ok {
		return nil, errorNoSuchColl(coll)
	}

	return m.get(key)
}

func (c *collections) getCollection(coll string) ([]*string, error) {
	c.RLock()
	m, ok := c.members[coll]
	c.RUnlock()

	if !ok {
		return nil, errorNoSuchColl(coll)
	}

	return m.getMembers()
}

func (c *collections) put(coll, key string, value *string) error {
	c.RLock()
	m, ok := c.members[coll]
	c.RUnlock()

	if !ok {
		// Another goroutine could have created the entry since our read
		// of ok, so need to Lock and verify again that it's still not
		// an entry.  Not doing so would drop the member that was `put`
		// by the other goroutine
		c.Lock()
		_, stillOk := c.members[coll]
		if !stillOk {
			m := newMember()
			c.members[coll] = m
		}
		c.Unlock()
	}

	return m.put(key, value)
}

func (c *collections) deleteKey(coll, key string) error {
	c.RLock()
	m, ok := c.members[coll]
	c.RUnlock()

	if !ok {
		return errorNoColl(coll)
	}

	m.delete(key)

	return nil
}

func (c *collections) deleteCollection(coll string) error {
	c.Lock()
	m, ok := c.members[coll]
	delete(c.members, coll)
	defer c.Unlock()
	if !ok {
		return errorNoSuchColl(coll)
	}

	m.deleteAll()

	return nil
}
