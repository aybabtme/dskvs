package dskvs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type Data struct {
	Descr string
}

/*
	Boilerplate
*/

func setUp(t *testing.T) *Store {
	store, err := NewStore("./db")
	if err != nil {
		t.Fatalf("Error creating store, %v", err)
	}

	err = store.Load()
	if err != nil {
		t.Fatalf("Error loading store, %v", err)
	}
	return store
}

func tearDown(store *Store, t *testing.T) {
	err := store.Close()
	if err != nil {
		t.Fatalf("Error closing store, %v", err)
	}
	err = os.RemoveAll(store.storagePath)
	if err != nil {
		t.Fatalf("Error deleting storage path, %v", err)
	}
}

func generateData(d Data, t *testing.T) []byte {
	dataBytes, err := json.Marshal(d)
	if err != nil {
		t.Fatal("Error with test data, %v", err)
	}
	return dataBytes
}

/*
	Common checks
*/

func checkGetIsEmpty(store *Store, key string, t *testing.T) {
	_, err := store.Get(key)
	if err == nil {
		t.Fatalf("Expected to receive KeyError but no error")
	}
	switch err := err.(type) {
	case KeyError:
		// Expected
		break
	default:
		t.Fatalf("Expected to receive KeyError but got <%s>", err)
	}
}

/*
	Test cases, single goroutine
*/

func TestCreatingStore(t *testing.T) {
	store := setUp(t)
	tearDown(store, t)
}

func TestSingleOperation(t *testing.T) {

	store := setUp(t)
	defer tearDown(store, t)

	key := "artist/daftpunk"
	expected := generateData(Data{"The peak of awesome"}, t)

	err := store.Put(key, expected)
	if err != nil {
		t.Fatalf("Error putting data in", err)
	}

	actual, err := store.Get(key)
	if err != nil {
		t.Fatalf("Error getting data back", err)
	}

	if !bytes.Equal(expected, actual) {
		t.Fatalf("Expected <%s> but was <%s>",
			expected,
			actual)
	}

	err = store.Delete(key)
	if err != nil {
		t.Fatalf("Error deleting data we just put", err)
	}

	checkGetIsEmpty(store, key, t)

}

func TestMultipleOperations(t *testing.T) {

	store := setUp(t)
	defer tearDown(store, t)

	coll := "artist"
	baseKey := "daftpunk"

	var key string
	var expected []byte
	var expectedList [][]byte
	var actual []byte
	for i := int(0); i < 10; i++ {
		key = coll + CollKeySep + baseKey + strconv.Itoa(i)
		expected = generateData(Data{"The peak of awesome" + strconv.Itoa(i)}, t)

		err := store.Put(key, expected)
		if err != nil {
			t.Errorf("Error putting data in", err)
		}
		actual, err = store.Get(key)
		if err != nil {
			t.Errorf("Error getting data back", err)
		}

		if !bytes.Equal(expected, actual) {
			t.Errorf("Expected <%s> but was <%s>",
				expected,
				actual)
		}

		expectedList = append(expectedList, expected)
	}

	actuaList, err := store.GetAll(coll)
	if err != nil {
		t.Fatalf("Error getting all keys with coll <%s>", coll)
	}
	if len(actuaList) != len(expectedList) {
		t.Fatalf("Expected to read %d items, but read %d",
			len(expectedList),
			len(actuaList))
	}

	err = store.DeleteAll("artist")
	if err != nil {
		t.Fatal("Error deleting key", err)
	}
	for i := int(0); i < 10; i++ {
		key = coll + CollKeySep + baseKey + strconv.Itoa(i)
		checkGetIsEmpty(store, key, t)
	}
}

/*
	Test cases, multiple goroutine
*/

func TestMultipleGoroutine(t *testing.T) {
	var kvCount int
	if testing.Short() {
		kvCount = 1000
	} else {
		kvCount = 10000
	}
	coll := "games"
	baseKey := "total annihilation #"

	type KeyValue struct {
		Key   string
		Value []byte
	}
	countGet := int64(0)
	countPut := int64(0)
	countDelete := int64(0)
	start := time.Now()

	store := setUp(t)
	defer tearDown(store, t)
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(runtime.NumCPU()))

	var writeGroup sync.WaitGroup
	var readGroup sync.WaitGroup
	var deleteGroup sync.WaitGroup
	writeErr := make(chan error)
	readErr := make(chan error)
	deleteErr := make(chan error)

	expectedList := make([]KeyValue, 0)
	for i := int(0); i < kvCount; i++ {
		key := coll + CollKeySep + baseKey + strconv.Itoa(i)
		data := generateData(Data{"It's fun!"}, t)
		kv := KeyValue{key, data}
		expectedList = append(expectedList, kv)
	}

	for _, kv := range expectedList {
		// Put all the values concurrently, ensure there's no error
		writeGroup.Add(1)
		go func(pair KeyValue, cErr chan error) {
			defer writeGroup.Done()
			err := store.Put(pair.Key, pair.Value)
			atomic.AddInt64(&countPut, 1)
			if err != nil {
				t.Fatal("Received an error", err)
				cErr <- err
			}

		}(kv, writeErr)
	}

	writeGroup.Wait()

	if len(writeErr) != 0 {
		t.Fatalf("Failed to write values concurrently, got %d errors",
			len(writeErr))
	}

	for _, kv := range expectedList {
		// Get each value, ensure they're the same
		readGroup.Add(1)
		go func(kv KeyValue, cErr chan error) {
			defer readGroup.Done()

			expected := kv.Value
			actual, err := store.Get(kv.Key)
			atomic.AddInt64(&countGet, 1)
			if err != nil {
				cErr <- err
			} else if !bytes.Equal(expected, actual) {
				t.Errorf("Expected <%s> but was <%s>",
					expected,
					actual)
			}

		}(kv, readErr)

		// GetAll values, ensure this slice is as big as all the value
		// we've put
		readGroup.Add(1)
		go func(cErr chan error) {
			defer readGroup.Done()

			actual, err := store.GetAll(coll)
			atomic.AddInt64(&countGet, int64(len(actual)))
			if err != nil {
				cErr <- err
			} else if len(actual) != len(expectedList) {
				t.Errorf("Expected len(store.GetAll)=<%s> but was <%s>",
					len(expectedList),
					len(actual))
			}
		}(readErr)
	}

	readGroup.Wait()

	if len(readErr) != 0 {
		t.Fatalf("Failed to write values concurrently, got %d errors",
			len(readErr))
	}

	for _, kv := range expectedList {
		// Put all the values concurrently, ensure there's no error
		deleteGroup.Add(1)
		go func(pair KeyValue, cErr chan error) {
			defer deleteGroup.Done()
			err := store.Delete(pair.Key)
			atomic.AddInt64(&countDelete, 1)
			if err != nil {
				t.Fatal("Received an error", err)
				cErr <- err
			}

		}(kv, deleteErr)
	}

	deleteGroup.Wait()

	if len(deleteErr) != 0 {
		t.Fatalf("Failed to write values concurrently, got %d errors",
			len(deleteErr))
	}

	err := store.DeleteAll(coll)
	atomic.AddInt64(&countDelete, int64(len(expectedList)))
	if err != nil {
		t.Fatalf("Error deleting all", err)
	}
	for _, kv := range expectedList {
		atomic.AddInt64(&countGet, 1)
		checkGetIsEmpty(store, kv.Key, t)
	}
	duration := time.Now().Sub(start)
	fmt.Printf("Concurrent test\n")
	fmt.Printf("- Get operations=%d\n", countGet)
	fmt.Printf("- Put operations=%d\n", countPut)
	fmt.Printf("- Delete operations=%d\n", countPut)
	fmt.Printf("in %fs ", duration.Seconds())
	fmt.Printf("by %d cpus, using %d concurrent goroutines\n",
		runtime.NumCPU(), kvCount)

}
