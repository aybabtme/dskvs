package dskvs

import (
	"bytes"
	"encoding/json"
	"os"
	"strconv"
	"testing"
)

type Data struct {
	Descr string
}

///////////////////////////////////////////////////////////////////////////////
// Boilerplate
///////////////////////////////////////////////////////////////////////////////

func setUp(t *testing.T) *Store {
	store, err := Open("./db")
	if err != nil {
		t.Fatalf("Error opening store, %v", err)
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

func TestMultipleOperationsPersistAfterClose(t *testing.T) {

	var kvCount int
	if testing.Short() {
		kvCount = 10
	} else {
		kvCount = 100
	}

	store := setUp(t)

	coll := "artist"
	baseKey := "daftpunk"

	type Pair struct {
		key   string
		value []byte
	}
	var expectedList []Pair

	for i := int(0); i < kvCount; i++ {
		for j := int(0); j < kvCount; j++ {

			var pair Pair
			pair.key = coll +
				strconv.Itoa(i) +
				CollKeySep +
				baseKey +
				strconv.Itoa(j)

			pair.value = generateData(Data{"The peak of awesome" + strconv.Itoa(i)}, t)

			err := store.Put(pair.key, pair.value)
			if err != nil {
				t.Errorf("Error putting data in, %v", err)
			}

			expectedList = append(expectedList, pair)
		}
	}

	// Don't use tearDown as it deletes the storage after use
	if err := store.Close(); err != nil {
		t.Fatalf("Error closing store, %v", err)
	}

	anotherStore := setUp(t)
	defer tearDown(anotherStore, t)

	for _, pair := range expectedList {
		actual, err := anotherStore.Get(pair.key)
		if err != nil {
			t.Fatalf("Failed getting value with key <%s> back, %v",
				pair.key, err)
		}

		if !bytes.Equal(pair.value, actual) {
			t.Errorf("Expected <%s> but was <%s>",
				pair.value,
				actual)
		}
	}
}

// Correctness

func TestGivenDataShouldBeCopied(t *testing.T) {
	store := setUp(t)
	defer tearDown(store, t)
	key := "a coll/a key"
	expected := []byte{
		0x00, 0x00, 0x00, 0x00,
	}
	modified := make([]byte, len(expected))
	copy(modified, expected)
	store.Put(key, modified)

	modified[0] = 0xDE
	modified[1] = 0xAD
	modified[2] = 0xBE
	modified[3] = 0xEF

	actual, err := store.Get(key)
	if err != nil {
		t.Fatalf("Couldn't get value back, %v", err)
	}

	if !bytes.Equal(expected, actual) {
		if bytes.Equal(modified, actual) {
			t.Errorf("Value was successfuly modified from outside the store,"+
				" expected <%v> was <%v>",
				expected, actual)
		} else {
			t.Errorf("Value was modified but took unknown value,"+
				" expected <%v> was <%v>",
				expected, actual)
		}
	}

}

/* To be determined, do we wish to provide speed or safety on returned values?
func TestReturnedDataShouldBeCopied(t *testing.T) {
	store := setUp(t)
	defer tearDown(store, t)
	key := "a coll/a key"
	expected := []byte{
		0x00, 0x00, 0x00, 0x00,
	}
	modified := make([]byte, len(expected))
	copy(modified, expected)
	store.Put(key, modified)

	temp, err := store.Get(key)

	temp[0] = 0xDE
	temp[1] = 0xAD
	temp[2] = 0xBE
	temp[3] = 0xEF

	actual, err := store.Get(key)
	if err != nil {
		t.Fatalf("Couldn't get value back, %v", err)
	}

	if !bytes.Equal(expected, actual) {
		if bytes.Equal(temp, actual) {
			t.Errorf("Value was successfuly modified from outside the store,"+
				" expected <%v> was <%v>",
				expected, actual)
		} else {
			t.Errorf("Value was modified from outside the store, took unknown value"+
				" expected <%v> was <%v>",
				expected, actual)
		}
	}
}
*/

// Error cases

func TestErrorWhenStorePointToNonDirectoryPath(t *testing.T) {
	filename := "test_regular_file"
	_, err := os.Create(filename)
	if err != nil {
		t.Fatalf("Error creating test file, %v", err)
	}
	defer os.Remove(filename)

	store, err := Open(filename)
	if _, isRightType := err.(PathError); !isRightType {
		defer tearDown(store, t)
		t.Errorf("Should have returned an error of type PathError")
	}
	err.Error() // Call it to make gocov happy

}

func TestErrorWhenStoreAlreadyUsingPath(t *testing.T) {
	path := "a_busy_path"
	store, err := Open(path)
	if err != nil {
		t.Fatalf("Error creating test store, %v", err)
	}
	defer tearDown(store, t)

	another, err := Open(path)
	if _, isRightType := err.(PathError); !isRightType {
		t.Errorf("Should have returned an error of type PathError, was %v",
			err)

		another.Close()
	}
	err.Error() // Call it to make gocov happy
}

func TestErrorWhenKeyGivenToGetIsMissingMember(t *testing.T) {
	keyWithoutMember := "a collection only"
	store := setUp(t)
	defer tearDown(store, t)

	_, err := store.Get(keyWithoutMember)
	if _, isRightType := err.(KeyError); !isRightType {
		t.Errorf("Should have returned an error of type KeyError, was %v",
			err)
	}
	err.Error() // Call it to make gocov happy
}

func TestErrorWhenKeyGivenToGetAllHasMember(t *testing.T) {
	keyWithMember := "a collection/with a member key"
	store := setUp(t)
	defer tearDown(store, t)

	_, err := store.GetAll(keyWithMember)
	if _, isRightType := err.(KeyError); !isRightType {
		t.Errorf("Should have returned an error of type KeyError, was %v",
			err)
	}
	err.Error() // Call it to make gocov happy
}

func TestErrorWhenKeyGivenToPutIsMissingMember(t *testing.T) {
	keyWithoutMember := "a collection only"
	store := setUp(t)
	defer tearDown(store, t)

	err := store.Put(keyWithoutMember, nil)
	if _, isRightType := err.(KeyError); !isRightType {
		t.Errorf("Should have returned an error of type KeyError, was %v",
			err)
	}
	err.Error() // Call it to make gocov happy
}

func TestErrorWhenKeyGivenToDeleteIsMissingMember(t *testing.T) {
	keyWithoutMember := "a collection only"
	store := setUp(t)
	defer tearDown(store, t)

	err := store.Delete(keyWithoutMember)
	if _, isRightType := err.(KeyError); !isRightType {
		t.Errorf("Should have returned an error of type KeyError, was %v",
			err)
	}
	err.Error() // Call it to make gocov happy
}

func TestErrorWhenKeyGivenToDeleteAllHasMember(t *testing.T) {
	keyWithMember := "a collection/with a member key"
	store := setUp(t)
	defer tearDown(store, t)

	err := store.DeleteAll(keyWithMember)
	if _, isRightType := err.(KeyError); !isRightType {
		t.Errorf("Should have returned an error of type KeyError, was %v",
			err)
	}
	err.Error() // Call it to make gocov happy
}

var invalidKeys = []string{
	"",
	"/a member",
	"acoll/",
}

func TestErrorWhenKeyGivenToGetIsInvalid(t *testing.T) {
	store := setUp(t)
	defer tearDown(store, t)

	for _, key := range invalidKeys {
		_, err := store.Get(key)
		if _, isRightType := err.(KeyError); !isRightType {
			t.Errorf("Should have returned an error of type KeyError"+
				"key was <%v>, error was %v",
				key,
				err)
		}
		err.Error() // Call it to make gocov happy
	}
}

func TestErrorWhenKeyGivenToGetAllIsInvalid(t *testing.T) {
	store := setUp(t)
	defer tearDown(store, t)

	for _, key := range invalidKeys {
		_, err := store.GetAll(key)
		if _, isRightType := err.(KeyError); !isRightType {
			t.Errorf("Should have returned an error of type KeyError"+
				"key was <%v>, error was %v",
				key,
				err)
		}
		err.Error() // Call it to make gocov happy
	}
}

func TestErrorWhenKeyGivenToPutIsInvalid(t *testing.T) {
	store := setUp(t)
	defer tearDown(store, t)

	for _, key := range invalidKeys {
		err := store.Put(key, nil)
		if _, isRightType := err.(KeyError); !isRightType {
			t.Errorf("Should have returned an error of type KeyError"+
				"key was <%v>, error was %v",
				key,
				err)
		}
		err.Error() // Call it to make gocov happy
	}
}

func TestErrorWhenKeyGivenToDeleteIsInvalid(t *testing.T) {

	store := setUp(t)
	defer tearDown(store, t)

	for _, key := range invalidKeys {
		err := store.Delete(key)
		if _, isRightType := err.(KeyError); !isRightType {
			t.Errorf("Should have returned an error of type KeyError"+
				"key was <%v>, error was %v",
				key,
				err)
		}
		err.Error() // Call it to make gocov happy
	}
}

func TestErrorWhenKeyGivenToDeleteAllIsInvalid(t *testing.T) {

	store := setUp(t)
	defer tearDown(store, t)

	for _, key := range invalidKeys {
		err := store.DeleteAll(key)
		if _, isRightType := err.(KeyError); !isRightType {
			t.Errorf("Should have returned an error of type KeyError"+
				"key was <%v>, error was %v",
				key,
				err)
		}
		err.Error() // Call it to make gocov happy
	}
}
