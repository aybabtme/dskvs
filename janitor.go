package dskvs

import (
	"log"
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

func (j *janitor) dirtyPageIfNoMember() chan *page {
	if len(j.ToCreate) != 0 {
		return nil
	}
	if len(j.ToDelete) != 0 {
		return nil
	}
	return j.DirtyPages
}

func (j *janitor) shouldDie() chan bool {
	if len(j.ToCreate) != 0 {
		return nil
	}
	if len(j.ToDelete) != 0 {
		return nil
	}
	if len(j.DirtyPages) != 0 {
		return nil
	}
	return j.mustDie
}

func (j *janitor) run() {
	go func() {
		for {
			select {
			case dirty := <-j.dirtyPageIfNoMember():
				writeToFile(dirty)
			case delete := <-j.ToDelete:
				deleteFolder(delete)
			case create := <-j.ToCreate:
				createFolder(create)
			case <-j.shouldDie():
				log.Printf("janitor dying")
				return
			}
		}
	}()
}

func (j *janitor) die() {
	j.mustDie <- true
}
