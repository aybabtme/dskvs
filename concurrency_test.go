package dskvs

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"
)

var kvCount int64

const (
	coll       = "games"
	baseKey    = "total annihilation #"
	goroutines = 1
	byteSize   = 100
)

func init() {
	flag.Parse()
	if testing.Short() {
		kvCount = 2048 // To be runable with race detector
	} else {
		kvCount = 10000
	}
	log.Printf("Concurrency Benchmark - Goroutines=%d, unique key-value=%d\n", goroutines, kvCount)
}

type keyValue struct {
	Key   string
	Value []byte
}

type Context struct {
	t      *testing.T
	s      *Store
	kv     keyValue
	wg     *sync.WaitGroup
	sem    chan int
	dur    chan time.Duration
	errors chan error
}

func TestOneOperationWithMultipleConcurrentRequest(t *testing.T) {

	log.Printf("Sequence of %d bytes operations by group, "+
		"%d concurrent request in each groups", byteSize, goroutines)

	store := setUp(t)
	defer tearDown(store, t)
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(runtime.NumCPU()))

	expectedList := generateKeyValueList(kvCount, t)

	log.Println("Put - first time")
	putStats := runTest(doPutRequest, 1, goroutines, store, expectedList, t)
	log.Println(putStats.String())

	log.Println("Put - rewrite")
	rePutStats := runTest(doPutRequest, 1, goroutines, store, expectedList, t)
	log.Println(rePutStats.String())

	log.Printf("Get")
	getStats := runTest(doGetRequest, 1, goroutines, store, expectedList, t)
	log.Println(getStats.String())

	log.Printf("Delete")
	deleteStats := runTest(doDeleteRequest, 1, goroutines, store, expectedList, t)
	log.Println(deleteStats.String())

	log.Printf("by %d cpus, using %d concurrent goroutines\n",
		runtime.NumCPU(), goroutines)

}

func TestManyOperationWithMultipleConcurrentRequest(t *testing.T) {

	store := setUp(t)
	defer tearDown(store, t)
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(runtime.NumCPU()))

	expectedList := generateKeyValueList(kvCount/3, t)

	// Start writing/reading concurrently, in random order
	concurrentFunc := func(ctx Context) {
		switch rand.Intn(5) {
		case 0:
			doFailGetRequest(ctx)
			doPutRequest(ctx)
			doDeleteRequest(ctx)
		case 1:
			doPutRequest(ctx)
			doGetRequest(ctx)
			doDeleteRequest(ctx)
		case 2:
			doPutRequest(ctx)
			doGetRequest(ctx)
			doDeleteRequest(ctx)
		case 3:
			doPutRequest(ctx)
			doDeleteRequest(ctx)
			doFailGetRequest(ctx)
		case 4:
			doFailGetRequest(ctx)
			doPutRequest(ctx)
			doGetRequest(ctx)
		}
	}
	_ = runTest(concurrentFunc, 3, goroutines, store, expectedList, t)

	// Cleanup
	err := store.DeleteAll(coll)
	if err != nil {
		t.Fatalf("Error deleting all", err)
	}
	for _, kv := range expectedList {
		checkGetIsEmpty(store, kv.Key, t)
	}

}

func TestConcurrentPutCanBeGetAllAndDeleteAll(t *testing.T) {

	store := setUp(t)
	defer tearDown(store, t)
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(runtime.NumCPU()))

	expectedList := generateKeyValueList(kvCount, t)

	// Don't care about the stats
	_ = runTest(doPutRequest, 1, 100, store, expectedList, t)

	actual, err := store.GetAll(coll)
	if err != nil {
		t.Errorf("Error on GetAll(%s), %v", coll, err)
	} else if len(actual) != len(expectedList) {
		t.Errorf("Expected len(store.GetAll)=<%s> but was <%s>",
			len(expectedList),
			len(actual))
	}

	err = store.DeleteAll(coll)
	if err != nil {
		t.Fatalf("Error deleting all", err)
	}
	for _, kv := range expectedList {
		checkGetIsEmpty(store, kv.Key, t)
	}

}

func runTest(
	testedFunc func(Context),
	nGo int,
	goroutines int,
	store *Store,
	expectedList []keyValue,
	t *testing.T,
) stats {

	cErr := make(chan error)
	var wg sync.WaitGroup
	durations := make(chan time.Duration, len(expectedList)*nGo)

	sem := make(chan int, goroutines)
	for _, kv := range expectedList {
		wg.Add(nGo)
		ctx := Context{
			t:      t,
			s:      store,
			kv:     kv,
			wg:     &wg,
			sem:    sem,
			dur:    durations,
			errors: cErr,
		}
		go testedFunc(ctx)

	}

	wg.Wait()
	close(durations)

	if len(cErr) != 0 {
		t.Fatalf("Failed to write values concurrently, got %d errors",
			len(cErr))
	}

	var dTList []time.Duration
	for dT := range durations {
		dTList = append(dTList, dT)
	}

	return newStats(dTList)
}

func doPutRequest(ctx Context) {
	ctx.sem <- 1
	t0 := time.Now()
	err := ctx.s.Put(ctx.kv.Key, ctx.kv.Value)
	dT := time.Since(t0)

	ctx.dur <- dT

	if err != nil {
		ctx.t.Fatal("Received an error", err)
		ctx.errors <- err
	}
	ctx.wg.Done()
	<-ctx.sem
}

func doGetRequest(ctx Context) {
	ctx.sem <- 1
	expected := ctx.kv.Value

	t0 := time.Now()
	actual, err := ctx.s.Get(ctx.kv.Key)
	dT := time.Since(t0)

	ctx.dur <- dT

	if err != nil {
		ctx.errors <- err
	} else if !bytes.Equal(expected, actual) {
		ctx.t.Errorf("Expected <%s> but was <%s>",
			expected,
			actual)
	}
	ctx.wg.Done()
	<-ctx.sem
}

// Get can fail when there's no such key, not true for Put and Delete
func doFailGetRequest(ctx Context) {
	ctx.sem <- 1
	t0 := time.Now()
	_, err := ctx.s.Get(ctx.kv.Key)
	dT := time.Since(t0)

	ctx.dur <- dT

	if _, ok := err.(KeyError); !ok {
		ctx.errors <- errors.New(fmt.Sprintf("Should have failed on Get(%s)", ctx.kv.Key))
	}
	ctx.wg.Done()
	<-ctx.sem
}

func doDeleteRequest(ctx Context) {
	ctx.sem <- 1
	t0 := time.Now()
	err := ctx.s.Delete(ctx.kv.Key)
	dT := time.Since(t0)

	ctx.dur <- dT

	if err != nil {
		ctx.t.Fatal("Received an error", err)
		ctx.errors <- err
	}
	ctx.wg.Done()
	<-ctx.sem
}

func generateKeyValueList(kvCount int64, t *testing.T) []keyValue {

	coll := "games"
	baseKey := "total annihilation #"

	kvList := make([]keyValue, kvCount)
	for i := int64(0); i < kvCount; i++ {
		key := coll + CollKeySep + baseKey + strconv.FormatInt(i, 10)
		data := make([]byte, byteSize)
		//data := generateData(Data{"It's fun!"}, t)
		kv := keyValue{key, data}
		kvList[i] = kv
	}

	return kvList
}
