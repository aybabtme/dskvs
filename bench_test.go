package dskvs

import (
	"fmt"
	"os"
	"strconv"
	"testing"
)

func setUpBench(b *testing.B) *Store {
	store, err := NewStore("./db")
	if err != nil {
		b.Fatalf("Error creating store, %v", err)
	}

	err = store.Load()
	if err != nil {
		b.Fatalf("Error loading store, %v", err)
	}
	return store
}

func tearDownBench(store *Store, b *testing.B) {
	err := store.Close()
	if err != nil {
		b.Fatalf("Error closing store, %v", err)
	}
	err = os.RemoveAll(store.storagePath)
	if err != nil {
		b.Fatalf("Error deleting storage path, %v", err)
	}
}

func benchPut(valSize int, b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping put benchmarks")
	}

	b.ReportAllocs()
	store := setUpBench(b)
	defer tearDownBench(store, b)
	baseKey := "hello/hello"
	val := make([]byte, valSize)
	for i := int(0); i < valSize; i++ {
		val[i] = uint8(i % 256)
	}
	keyList := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		key := baseKey + strconv.Itoa(i)
		keyList = append(keyList, key)
	}
	b.ResetTimer()
	for _, key := range keyList {
		err := store.Put(key, val)
		if err != nil {
			b.Errorf("Get has error %v", err)
		}
	}
	b.StopTimer()
}

func BenchmarkPut100B(b *testing.B) {
	benchPut(100, b)
}

func BenchmarkPut400B(b *testing.B) {
	benchPut(400, b)
}

func BenchmarkPut1KB(b *testing.B) {
	benchPut(1000, b)
}

func BenchmarkPut4KB(b *testing.B) {
	benchPut(4000, b)
}

func BenchmarkPut10KB(b *testing.B) {
	benchPut(10000, b)
}

func BenchmarkPut40KB(b *testing.B) {
	benchPut(40000, b)
}

func BenchmarkPut100KB(b *testing.B) {
	benchPut(100000, b)
}

func BenchmarkPut400KB(b *testing.B) {
	benchPut(400000, b)
}

func BenchmarkPut1MB(b *testing.B) {
	benchPut(1000000, b)
}

func benchGet(valSize int, b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping get benchmarks")
	}

	fmt.Printf("N=%d\n", b.N)

	b.ReportAllocs()
	store := setUpBench(b)
	defer tearDownBench(store, b)
	baseKey := "hello/hello"
	val := make([]byte, valSize)
	for i := int(0); i < valSize; i++ {
		val[i] = uint8(i % 256)
	}

	keyList := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		key := baseKey + strconv.Itoa(i)
		keyList[i] = key
		store.Put(key, val)
	}
	b.ResetTimer()
	for _, key := range keyList {
		_, err := store.Get(key)
		if err != nil {
			b.Errorf("Get has error %v", err)
		}
	}
	b.StopTimer()
}

func BenchmarkGet100B(b *testing.B) {
	benchGet(100, b)
}

func BenchmarkGet400B(b *testing.B) {
	benchGet(400, b)
}

func BenchmarkGet1KB(b *testing.B) {
	benchGet(1000, b)
}

func BenchmarkGet4KB(b *testing.B) {
	benchGet(4000, b)
}

func BenchmarkGet10KB(b *testing.B) {
	benchGet(10000, b)
}

func BenchmarkGet40KB(b *testing.B) {
	benchGet(40000, b)
}

func BenchmarkGet100KB(b *testing.B) {
	benchGet(100000, b)
}

func BenchmarkGet400KB(b *testing.B) {
	benchGet(400000, b)
}

func BenchmarkGet1MB(b *testing.B) {
	benchGet(1000000, b)
}
