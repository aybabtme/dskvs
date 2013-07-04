package dskvs

import (
	"bytes"
	"fmt"
	"log"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"
)

type keyValue struct {
	Key   string
	Value []byte
}

type Context struct {
	t      *testing.T
	s      *Store
	kv     keyValue
	wg     *sync.WaitGroup
	dur    chan time.Duration
	errors chan error
}

func TestMultipleGoroutine(t *testing.T) {
	var kvCount int64
	if testing.Short() {
		kvCount = 1000
	} else {
		kvCount = 5000
	}

	fmt.Printf("Concurrent test, keyValQty=%d\n", kvCount)

	coll := "games"
	baseKey := "total annihilation #"

	store := setUp(t)
	defer tearDown(store, t)
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(runtime.NumCPU()))

	// Generate
	expectedList := make([]keyValue, 0)
	for i := int64(0); i < kvCount; i++ {
		key := coll + CollKeySep + baseKey + strconv.FormatInt(i, 10)
		data := generateData(Data{"It's fun!"}, t)
		kv := keyValue{key, data}
		expectedList = append(expectedList, kv)
	}

	// Put
	log.Println("Put operations")
	putStats := runTest(doPutRequest, store, expectedList, t)
	log.Println(putStats.String())

	// Get
	log.Printf("- Get operations=%d\n", kvCount)
	getStats := runTest(doGetRequest, store, expectedList, t)
	log.Println(getStats.String())

	// GetAll
	actual, err := store.GetAll(coll)
	if err != nil {
		t.Errorf("Error on GetAll(%s), %v", coll, err)
	} else if len(actual) != len(expectedList) {
		t.Errorf("Expected len(store.GetAll)=<%s> but was <%s>",
			len(expectedList),
			len(actual))
	}

	// Delete
	log.Printf("- Delete operations=%d\n", len(expectedList))
	deleteStats := runTest(doDeleteRequest, store, expectedList, t)
	log.Println(deleteStats.String())

	// DeleteAll
	err = store.DeleteAll(coll)
	if err != nil {
		t.Fatalf("Error deleting all", err)
	}
	for _, kv := range expectedList {
		checkGetIsEmpty(store, kv.Key, t)
	}

	log.Printf("by %d cpus, using %d concurrent goroutines\n",
		runtime.NumCPU(), kvCount)

}

func runTest(doer func(Context), store *Store, expectedList []keyValue, t *testing.T) stats {

	cErr := make(chan error)
	var wg sync.WaitGroup
	durations := make(chan time.Duration, len(expectedList))

	for _, kv := range expectedList {
		wg.Add(1)
		ctx := Context{
			t:      t,
			s:      store,
			kv:     kv,
			wg:     &wg,
			dur:    durations,
			errors: cErr,
		}
		go doer(ctx)
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
	defer ctx.wg.Done()

	t0 := time.Now()
	err := ctx.s.Put(ctx.kv.Key, ctx.kv.Value)
	dT := time.Since(t0)

	ctx.dur <- dT

	if err != nil {
		ctx.t.Fatal("Received an error", err)
		ctx.errors <- err
	}
}

func doGetRequest(ctx Context) {
	defer ctx.wg.Done()

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
}

func doDeleteRequest(ctx Context) {
	defer ctx.wg.Done()

	t0 := time.Now()
	err := ctx.s.Delete(ctx.kv.Key)
	dT := time.Since(t0)

	ctx.dur <- dT

	if err != nil {
		ctx.t.Fatal("Received an error", err)
		ctx.errors <- err
	}
}
