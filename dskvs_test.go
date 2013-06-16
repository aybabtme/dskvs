package dskvs

import (
	"bytes"
	"encoding/json"
	"testing"
)

type TestData struct {
	Descr string
}

func TestCreatingStore(t *testing.T) {
	store, err := NewStore("./db")
	if err != nil {
		t.Fatalf("Error creating store", err)
	}

	err = store.Load()
	if err != nil {
		t.Fatalf("Error loading store", err)
	}

	err = store.Close()
	if err != nil {
		t.Fatalf("Error closing store", err)
	}
}

func TestGet(t *testing.T) {

	store, err := NewStore("./db")
	if err != nil {
		t.Fatalf("Error creating store", err)
	}

	err = store.Load()
	if err != nil {
		t.Fatalf("Error loading store", err)
	}

	expected, err := json.Marshal(TestData{"the great and epics"})
	if err != nil {
		t.Fatal("Error with test data", err)
	}

	store.Put("artist/daftpunk", expected)

	actual, err := store.Get("artist/daftpunk")
	if err != nil {
		t.Fatalf("Error getting data back", err)
	}

	if bytes.Equal(expected, actual) {
		t.Fatalf("Expected <%s> but was <%s>",
			expected,
			actual)
	}

	err = store.Close()
	if err != nil {
		t.Fatalf("Error closing store", err)
	}
}
