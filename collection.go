package dskvs

import (
	"sync"
)

type collections struct {
	lock    sync.RWMutex
	members map[string]*member
}

func (c *collections) get(fullKey string) (string, error) {

}

func (c *collections) getCollection(coll string) ([]string, error) {

}

func (c *collections) put(fullKey, value string) error {

}

func (c *collections) deleteKey(coll, key string) error {

}

func (c *collections) deleteCollection(coll string) error {

}
