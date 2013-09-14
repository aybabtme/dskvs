package benchmark

import (
	"fmt"
	"github.com/aybabtme/dskvs"
	"math/rand"
	"os"
	"testing"
)

// This benchmark is pretty much verbatim taken from
// "github.com/peterbourgon/diskv", because it's a similar
// key-value store and it's easy to compare to it.

func shuffle(keys []string) {
	ints := rand.Perm(len(keys))
	for i := range keys {
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
	KEY_COUNT   = 10000
	storagePath = "./db"
)

func genKeys() []string {
	keys := make([]string, KEY_COUNT)
	for i := 0; i < KEY_COUNT; i++ {
		keys[i] = fmt.Sprintf("speed_test/%d", i)
	}
	return keys
}

func setUpBench(b *testing.B) *dskvs.Store {
	store, err := dskvs.Open(storagePath)
	if err != nil {
		b.Fatalf("Error opening store, %v", err)
	}
	return store
}

func tearDownBench(store *dskvs.Store, b *testing.B) {
	err := store.Close()
	if err != nil {
		b.Fatalf("Error closing store, %v", err)
	}
	err = os.RemoveAll(storagePath)
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
		_, _, _ = d.Get(keys[i%len(keys)])
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
func Benchmark_Dskvs_Put_32B(b *testing.B) {
	benchPut(32, b)
}

func Benchmark_Dskvs_Put_1KB(b *testing.B) {
	benchPut(1024, b)
}

func Benchmark_Dskvs_Put_4KB(b *testing.B) {
	benchPut(4096, b)
}

func Benchmark_Dskvs_Put_10KB(b *testing.B) {
	benchPut(10240, b)
}

// Get
func Benchmark_Dskvs_Get_32B(b *testing.B) {
	benchGet(32, b)
}

func Benchmark_Dskvs_Get_1KB(b *testing.B) {
	benchGet(1024, b)
}

func Benchmark_Dskvs_Get_4KB(b *testing.B) {
	benchGet(4096, b)
}

func Benchmark_Dskvs_Get_10KB(b *testing.B) {
	benchGet(10240, b)
}
