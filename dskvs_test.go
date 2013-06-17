package dskvs

import (
	"bytes"
	"encoding/json"
	"strconv"
	"testing"
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
		t.Fatalf("Error creating store", err)
	}

	err = store.Load()
	if err != nil {
		t.Fatalf("Error loading store", err)
	}
	return store
}

func tearDown(store *Store, t *testing.T) {
	err := store.Close()
	if err != nil {
		t.Fatalf("Error closing store", err)
	}
}

func getData(d Data, t *testing.T) []byte {
	dataBytes, err := json.Marshal(d)
	if err != nil {
		t.Fatal("Error with test data", err)
	}
	return dataBytes
}

/*
	Test cases
*/

func TestCreatingStore(t *testing.T) {
	store := setUp(t)
	tearDown(store, t)
}

func TestSingleOperation(t *testing.T) {

	store := setUp(t)
	defer tearDown(store, t)

	key := "artist/daftpunk"
	expected := getData(Data{"The peak of awesome"}, t)

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
	_, err = store.Get(key)
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

func TestMultipleOperations(t *testing.T) {

	store := setUp(t)
	defer tearDown(store, t)

	var key string
	var expected []byte
	var actual []byte
	for i := int(0); i < 10; i++ {
		key = "artist/daftpunk" + strconv.Itoa(i)
		expected = getData(Data{"The peak of awesome" + strconv.Itoa(i)}, t)

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
	}

	err := store.DeleteAll("artist")
	if err != nil {
		t.Fatal("Error deleting key", err)
	}
	for i := int(0); i < 10; i++ {
		key = "artist/daftpunk" + strconv.Itoa(i)
		_, err = store.Get(key)
		if err == nil {
			t.Fatalf("Expected to receive KeyError, but got no error at all")
		}
		switch err := err.(type) {
		case KeyError:
			// Expected
			break
		default:
			t.Fatalf("Expected to receive KeyError but got <%s>", err)
		}
	}
}
