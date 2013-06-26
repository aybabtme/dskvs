package dskvs

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"io/ioutil"
	"log"
	"net/url"
	"os"
)

type janitor struct {
	DirtyPages chan *page
	ToDelete   chan *member
	ToCreate   chan *member
	mustDie    chan bool
}

func newJanitor() janitor {
	return janitor{
		make(chan *page),
		make(chan *member),
		make(chan bool),
	}
}

func (j *janitor) loadStore(s *Store) error {
	log.Printf("janitor.loadStore(%s) called but not yet implemented",
		s.storagePath)
	return nil
}

func (j *janitor) unloadStore(s *Store) error {
	log.Printf("janitor.unloadStore(%s) called but not yet implemented",
		s.storagePath)
	return nil
}

func (j *janitor) run() {
	go func() {
		for {
			select {
			case dirty := <-j.DirtyPages:
				writeToFile(dirty)
			case delete := <-j.ToDelete:
				deleteFolder(delete)
			case create := <-j.ToCreate:
				createFolder(create)
			case <-j.mustDie:
				log.Printf("janitor dying")
				return
			}
		}
	}()
}

func (j *janitor) die() {
	j.mustDie <- true
}
