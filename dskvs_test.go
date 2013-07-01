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

///////////////////////////////////////////////////////////////////////////////
// Boilerplate
///////////////////////////////////////////////////////////////////////////////

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

///////////////////////////////////////////////////////////////////////////////
// Common checks
///////////////////////////////////////////////////////////////////////////////

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

///////////////////////////////////////////////////////////////////////////////
// Single goroutine
///////////////////////////////////////////////////////////////////////////////

// Normal cases

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
		t.Fatalf("Error putting data in, %v", err)
	}

	actual, err := store.Get(key)
	if err != nil {
		t.Fatalf("Error getting data back, %v", err)
	}

	if !bytes.Equal(expected, actual) {
		t.Fatalf("Expected <%s> but was <%s>",
			expected,
			actual)
	}

	err = store.Delete(key)
	if err != nil {
		t.Fatalf("Error deleting data we just put, %v", err)
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
			t.Errorf("Error putting data in, %v", err)
		}
		actual, err = store.Get(key)
		if err != nil {
			t.Errorf("Error getting data back, %v", err)
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
		t.Fatal("Error deleting key, %v", err)
	}
	for i := int(0); i < 10; i++ {
		key = coll + CollKeySep + baseKey + strconv.Itoa(i)
		checkGetIsEmpty(store, key, t)
	}
}

func TestStorePersistPutAfterClose(t *testing.T) {
	store := setUp(t)

	key := "artist/the prodigy"
	expected := generateData(Data{"Beyond the peak"}, t)

	if err := store.Put(key, expected); err != nil {
		tearDown(store, t)
		t.Fatalf("Error putting data in, %v", err)
	}

	// Don't use tearDown as it deletes the storage after use
	if err := store.Close(); err != nil {
		t.Fatalf("Error closing store, %v", err)
	}

	otherStore := setUp(t)
	defer tearDown(otherStore, t)

	actual, err := otherStore.Get(key)
	if err != nil {
		t.Fatalf("Error getting data back, %v", err)
	}
	if !bytes.Equal(expected, actual) {
		t.Fatalf("Expected <%s> but was <%s>",
			expected,
			actual)
	}
}

func TestStorePersistDeleteAfterClose(t *testing.T) {
	store := setUp(t)

	key := "artist/the prodigy"
	expected := generateData(Data{"Beyond the peak"}, t)

	if err := store.Put(key, expected); err != nil {
		tearDown(store, t)
		t.Fatalf("Error putting value in, %v", err)
	}

	if err := store.Delete(key); err != nil {
		tearDown(store, t)
		t.Fatalf("Error deleting value, %v", err)
	}

	// Don't use tearDown as it deletes the storage after use
	if err := store.Close(); err != nil {
		t.Fatalf("Error closing store, %v", err)
	}

	otherStore := setUp(t)
	defer tearDown(otherStore, t)

	checkGetIsEmpty(otherStore, key, t)

}

// Error cases

func TestErrorWhenStorePointToNonDirectoryPath(t *testing.T) {
	filename := "test_regular_file"
	_, err := os.Create(filename)
	if err != nil {
		t.Fatalf("Error creating test file, %v", err)
	}
	defer os.Remove(filename)

	store, err := NewStore(filename)
	if _, isRightType := err.(PathError); !isRightType {
		defer tearDown(store, t)
		t.Errorf("Should have returned an error of type PathError")
	}

}

func TestErrorWhenStoreAlreadyUsingPath(t *testing.T) {
	path := "a_busy_path"
	store, err := NewStore(path)
	if err != nil {
		t.Fatalf("Error creating test store, %v", err)
	}
	if err := store.Load(); err != nil {
		t.Fatalf("Error loading test store, %v", err)
	}
	defer tearDown(store, t)

	another, err := NewStore(path)
	if err != nil {
		t.Fatalf("Error creating second test store, %v", err)
	}
	err = another.Load()
	if _, isRightType := err.(PathError); !isRightType {
		defer tearDown(another, t)
		t.Errorf("Should have returned an error of type PathError, was %v",
			err)
	}
}

func TestErrorWhenStoreNotLoaded(t *testing.T) {
	path := "a_busy_path"
	store, err := NewStore(path)
	if err != nil {
		t.Fatalf("Error creating test store, %v", err)
	}
	err = store.Close()
	if _, isRightType := err.(StoreError); !isRightType {
		t.Errorf("Should have returned an error of type StoreError, was %v",
			err)
	}

	_, err = store.Get("coll/key")
	if _, isRightType := err.(StoreError); !isRightType {
		t.Errorf("Should have returned an error of type StoreError, was %v",
			err)
	}

	_, err = store.GetAll("coll")
	if _, isRightType := err.(StoreError); !isRightType {
		t.Errorf("Should have returned an error of type StoreError, was %v",
			err)
	}

	err = store.Put("coll/key", nil)
	if _, isRightType := err.(StoreError); !isRightType {
		t.Errorf("Should have returned an error of type StoreError, was %v",
			err)
	}

	err = store.Delete("coll/key")
	if _, isRightType := err.(StoreError); !isRightType {
		t.Errorf("Should have returned an error of type StoreError, was %v",
			err)
	}

	err = store.DeleteAll("coll")
	if _, isRightType := err.(StoreError); !isRightType {
		t.Errorf("Should have returned an error of type StoreError, was %v",
			err)
	}

}

///////////////////////////////////////////////////////////////////////////////
// Multiple goroutine
///////////////////////////////////////////////////////////////////////////////

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
