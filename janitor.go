package dskvs

import (
	"log"
)

type janitor struct {
	DirtyPages chan *page
	ToDelete   chan *member
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
				updateFile(dirty)
			case delete := <-j.ToDelete:
				deleteFolder(delete)
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

func updateFile(dirty *page) {
	log.Printf("updateFile(%s) called but not yet implemented",
		dirty.key)
}

func deleteFolder(delete *member) {
	log.Printf("deleteFolder(%s) called but not yet implemented",
		delete.coll)
}
