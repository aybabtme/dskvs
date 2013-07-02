package dskvs

import (
	"sync"
)

type collections struct {
	sync.RWMutex
	basepath string
	members  map[string]*member
}

func newCollections(basepath string) *collections {
	return &collections{
		basepath: basepath,
		members:  make(map[string]*member),
	}
}

func (c *collections) get(coll, key string) ([]byte, error) {
	c.RLock()
	m, ok := c.members[coll]
	c.RUnlock()
	if !ok {
		return nil, errorNoSuchColl(coll)
	}

	return m.get(key)
}

func (c *collections) getCollection(coll string) ([][]byte, error) {
	c.RLock()
	m, ok := c.members[coll]
	c.RUnlock()

	if !ok {
		return nil, errorNoSuchColl(coll)
	}

	return m.getMembers(), nil
}

func (c *collections) put(coll, key string, value []byte) error {
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
			m = newMember(c.basepath, coll)
			c.members[coll] = m
			jan.ToCreate <- m
		}
		c.Unlock()
	}

	m.put(key, value)

	return nil
}

func (c *collections) deleteKey(coll, key string) error {
	c.RLock()
	m, ok := c.members[coll]
	c.RUnlock()

	if !ok {
		return errorNoSuchColl(coll)
	}

	m.delete(key)

	return nil
}

func (c *collections) deleteCollection(coll string) error {
	c.RLock()
	_, ok := c.members[coll]
	c.RUnlock()

	if ok {
		c.Lock()
		m, ok := c.members[coll]
		delete(c.members, coll)
		c.Unlock()
		// Was deleted in between our read-lock and the current write-lock
		if !ok {
			return errorNoSuchColl(coll)
		}

		// TODO : This is not really necessary, can just delete the folder
		// at once and save some IO.
		m.deleteAll()
		jan.ToDelete <- m
	} else {
		return errorNoSuchColl(coll)
	}

	return nil
}
