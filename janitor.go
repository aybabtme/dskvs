package dskvs

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync/atomic"
)

type janitor struct {
	toWriteChan  chan *page
	toWriteCount int64

	toDeleteChan  chan *member
	toDeleteCount int64

	toCreateChan  chan *member
	toCreateCount int64

	mustDie            chan bool
	blockUntilFinished chan bool
}

func newJanitor() janitor {
	return janitor{
		make(chan *page), 0,
		make(chan *member), 0,
		make(chan *member), 0,
		make(chan bool, 1),
		make(chan bool, 1),
	}
}

func (j *janitor) writePage(p *page) {
	atomic.AddInt64(&j.toWriteCount, 1)
	j.toWriteChan <- p
}

func (j *janitor) createFolder(m *member) {
	atomic.AddInt64(&j.toCreateCount, 1)
	j.toCreateChan <- m
}

func (j *janitor) deleteFolder(m *member) {
	atomic.AddInt64(&j.toDeleteCount, 1)
	j.toDeleteChan <- m
}

func (j *janitor) hasNoFolderOps() chan *page {
	if j.toCreateCount != 0 {
		return nil
	}
	if j.toDeleteCount != 0 {
		return nil
	}
	return j.toWriteChan
}

func (j *janitor) shouldDie() chan bool {

	backlog := j.toCreateCount +
		j.toDeleteCount +
		j.toWriteCount

	if backlog != 0 && len(j.mustDie) != 0 {

		log.Printf("Dying - backlog: write=%d, rmdir=%d, mkdir=%d",
			j.toWriteCount,
			j.toDeleteCount,
			j.toCreateCount)

		return nil
	}
	return j.mustDie
}

func (j *janitor) run() {
	go func() {
		for {
			select {
			case page := <-j.hasNoFolderOps():
				atomic.AddInt64(&j.toWriteCount, -1)
				writeToFile(page)

			case member := <-j.toDeleteChan:
				atomic.AddInt64(&j.toDeleteCount, -1)
				deleteFolder(member)

			case member := <-j.toCreateChan:
				atomic.AddInt64(&j.toCreateCount, -1)
				createFolder(member)

			case <-j.shouldDie():
				j.blockUntilFinished <- false
				return
			}

		}
	}()
}

func (j *janitor) die() {
	j.mustDie <- true
}

func (j *janitor) loadStore(s *Store) error {

	basepath := s.storagePath
	possibleColl, err := ioutil.ReadDir(basepath)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		log.Printf("Can't list directory at path %s: %v", basepath, err)
		return err
	}

	var memberPathList []string
	var memberPath string
	for _, file := range possibleColl {
		if file.IsDir() {
			memberPath = filepath.Join(basepath, file.Name())
			memberPathList = append(memberPathList, memberPath)
			s.coll.members[file.Name()] = newMember(basepath, file.Name())
		}
	}

	var aPage *page
	var pagePath string
	for _, member := range memberPathList {
		possiblePage, err := ioutil.ReadDir(member)
		if err != nil {
			log.Printf("\t... skipping, can't list directory at path <%s>: %v",
				basepath, err)
			continue
		}

		for _, file := range possiblePage {
			if !file.Mode().IsRegular() {
				log.Printf("\t... skipping irregular file <%s>", file.Name())
				continue
			}
			pagePath = filepath.Join(member, file.Name())
			aPage, err = readFromFile(pagePath)
			if err != nil {
				log.Printf("\t... skipping, error reading possible page file: %v",
					err)
				continue
			}
			s.coll.members[aPage.coll].entries[aPage.key] = aPage
		}
	}
	return nil
}

func (j *janitor) unloadStore(s *Store) error {
	j.die()
	<-j.blockUntilFinished
	return nil
}
