package dskvs

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type janitor struct {
	DirtyPages         chan *page
	ToDelete           chan *member
	ToCreate           chan *member
	mustDie            chan bool
	blockUntilFinished chan bool
}

func newJanitor() janitor {
	return janitor{
		make(chan *page),
		make(chan *member),
		make(chan *member),
		make(chan bool),
		make(chan bool),
	}
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
	createBacklog := len(j.ToCreate)
	deleteBacklog := len(j.ToDelete)
	pageBacklog := len(j.DirtyPages)

	if createBacklog != 0 ||
		deleteBacklog != 0 ||
		pageBacklog != 0 {
		log.Printf("Janitor has backlog of length %d",
			createBacklog+deleteBacklog+pageBacklog)
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
				j.blockUntilFinished <- false
				return
			}

		}
	}()
}

func (j *janitor) die() {
	j.mustDie <- true
}
