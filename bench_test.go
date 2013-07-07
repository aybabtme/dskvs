package dskvs

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
)

// This benchmark is pretty much verbatim taken from
// "github.com/peterbourgon/diskv", because it's a similar
// key-value store and it's easy to compare to it.

func shuffle(keys []string) {
	ints := rand.Perm(len(keys))
	for i, _ := range keys {
		keys[i], keys[ints[i]] = keys[ints[i]], keys[i]
	}
}

func genValue(size int) []byte {
	v := make([]byte, size)
	for i := 0; i < size; i++ {
		v[i] = uint8((rand.Int() % 26) + 97) // a-z
	}
	return v
}

const (
	KEY_COUNT = 10000
)

func genKeys() []string {
	keys := make([]string, KEY_COUNT)
	for i := 0; i < KEY_COUNT; i++ {
		keys[i] = fmt.Sprintf("speed_test/%d", i)
	}
	return keys
}

func setUpBench(b *testing.B) *Store {
	store, err := Open("./db")
	if err != nil {
		b.Fatalf("Error opening store, %v", err)
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

func benchGet(size int, b *testing.B) {
	b.StopTimer()
	d := setUpBench(b)
	defer tearDownBench(d, b)

	keys := genKeys()
	value := genValue(size)
	for _, key := range keys {

		if err := d.Put(key, value); err != nil {
			b.Fatalf("Failed putting values, %v", err)
		}
	}

	shuffle(keys)
	b.SetBytes(int64(size))

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.Get(keys[i%len(keys)])
	}
	b.StopTimer()
}

func benchPut(size int, b *testing.B) {
	b.StopTimer()

	d := setUpBench(b)
	defer tearDownBench(d, b)

	keys := genKeys()
	value := genValue(size)
	shuffle(keys)
	b.SetBytes(int64(size))

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_ = d.Put(keys[i%len(keys)], value)
	}
	b.StopTimer()
}

// Put
func BenchmarkPut32B(b *testing.B) {
	benchPut(32, b)
}

func BenchmarkPut1KB(b *testing.B) {
	benchPut(1024, b)
}

func BenchmarkPut4KB(b *testing.B) {
	benchPut(4096, b)
}

func BenchmarkPut10KB(b *testing.B) {
	benchPut(10240, b)
}

// Get
func BenchmarkGet32B(b *testing.B) {
	benchGet(32, b)
}

func BenchmarkGet1KB(b *testing.B) {
	benchGet(1024, b)
}

func BenchmarkGet4KB(b *testing.B) {
	benchGet(4096, b)
}

func BenchmarkGet10KB(b *testing.B) {
	benchGet(10240, b)
}
