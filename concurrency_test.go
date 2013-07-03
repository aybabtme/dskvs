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

func testPut(store *Store, expectedList []keyValue, t *testing.T) stats {

	putErr := make(chan error)
	var putGroup sync.WaitGroup
	durations := make(chan time.Duration, len(expectedList))

	for _, kv := range expectedList {

		// Put all the values concurrently, ensure there's no error
		putGroup.Add(1)
		go func(pair keyValue, durations chan time.Duration, cErr chan error) {
			defer putGroup.Done()

			t0 := time.Now()
			err := store.Put(pair.Key, pair.Value)
			dT := time.Since(t0)

			durations <- dT

			if err != nil {
				t.Fatal("Received an error", err)
				cErr <- err
			}

		}(kv, durations, putErr)
	}

	putGroup.Wait()
	close(durations)

	if len(putErr) != 0 {
		t.Fatalf("Failed to write values concurrently, got %d errors",
			len(putErr))
	}

	var dTList []time.Duration
	for dT := range durations {
		dTList = append(dTList, dT)
	}

	return newStats(dTList)
}

func testGet(store *Store, expectedList []keyValue, t *testing.T) stats {

	getErr := make(chan error)
	var getGroup sync.WaitGroup
	durations := make(chan time.Duration, len(expectedList))

	for _, kv := range expectedList {
		getGroup.Add(1)
		go func(kv keyValue, durations chan time.Duration, cErr chan error) {
			defer getGroup.Done()

			expected := kv.Value

			t0 := time.Now()
			actual, err := store.Get(kv.Key)
			dT := time.Since(t0)

			durations <- dT

			if err != nil {
				cErr <- err
			} else if !bytes.Equal(expected, actual) {
				t.Errorf("Expected <%s> but was <%s>",
					expected,
					actual)
			}

		}(kv, durations, getErr)
	}

	getGroup.Wait()
	close(durations)

	if len(getErr) != 0 {
		t.Fatalf("Failed to read values concurrently, got %d errors",
			len(getErr))
	}

	var dTList []time.Duration
	for dT := range durations {
		dTList = append(dTList, dT)
	}

	return newStats(dTList)
}

func testDelete(store *Store, expectedList []keyValue, t *testing.T) stats {

	deleteErr := make(chan error)
	var deleteGroup sync.WaitGroup

	durations := make(chan time.Duration, len(expectedList))

	for _, kv := range expectedList {
		deleteGroup.Add(1)
		go func(kv keyValue, durations chan time.Duration, cErr chan error) {
			defer deleteGroup.Done()

			t0 := time.Now()
			err := store.Delete(kv.Key)
			dT := time.Since(t0)

			durations <- dT

			if err != nil {
				t.Fatal("Received an error", err)
				cErr <- err
			}

		}(kv, durations, deleteErr)
	}

	deleteGroup.Wait()
	close(durations)

	if len(deleteErr) != 0 {
		t.Fatalf("Failed to delete values concurrently, got %d errors",
			len(deleteErr))
	}
	var dTList []time.Duration
	for dT := range durations {
		dTList = append(dTList, dT)
	}

	return newStats(dTList)
}

func TestMultipleGoroutine(t *testing.T) {
	var kvCount int64
	if testing.Short() {
		kvCount = 1000
	} else {
		kvCount = 100000
	}

	fmt.Printf("Concurrent test, keyValQty=%d\n", kvCount)

	coll := "games"
	baseKey := "total annihilation #"

	store := setUp(t)
	defer tearDown(store, t)
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(runtime.NumCPU()))

	/* GENERATE DATA */

	expectedList := make([]keyValue, 0)
	for i := int64(0); i < kvCount; i++ {
		key := coll + CollKeySep + baseKey + strconv.FormatInt(i, 10)
		data := generateData(Data{"It's fun!"}, t)
		kv := keyValue{key, data}
		expectedList = append(expectedList, kv)
	}

	/* PUT */
	log.Println("Put operations")
	putStats := testPut(store, expectedList, t)
	log.Println(putStats.String())

	/* GET */
	log.Printf("- Get operations=%d\n", kvCount)
	getStats := testGet(store, expectedList, t)
	log.Println(getStats.String())

	/* GETALL */
	actual, err := store.GetAll(coll)
	if err != nil {
		t.Errorf("Error on GetAll(%s), %v", coll, err)
	} else if len(actual) != len(expectedList) {
		t.Errorf("Expected len(store.GetAll)=<%s> but was <%s>",
			len(expectedList),
			len(actual))
	}

	/* DELETE */
	log.Printf("- Delete operations=%d\n", len(expectedList))
	deleteStats := testDelete(store, expectedList, t)
	log.Println(deleteStats.String())

	/* DELETEALL */
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
