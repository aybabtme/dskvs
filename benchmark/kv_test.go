package benchmark

import (
	"github.com/cznic/kv"
	"os"
	"runtime"
	"testing"
)

var (
	kvDir     = "kvDir"
	kvPrefix  = "kvPre"
	kvSuffix  = "kvSuf"
	kvOpts    = &kv.Options{}
	goMaxProc int
)

func setUpKv(b *testing.B) *kv.DB {
	goMaxProc = runtime.GOMAXPROCS(2)
	os.MkdirAll(kvDir, 0770)
	store, err := kv.CreateTemp(kvDir, kvPrefix, kvSuffix, kvOpts)
	if err != nil {
		b.Fatalf("Error opening store, %v", err)
	}
	return store
}

func tearDownKv(store *kv.DB, b *testing.B) {
	if err := store.Close(); err != nil {
		b.Fatalf("Error closing kv store, %v", err)
	}

	if err := os.RemoveAll(kvDir); err != nil {
		b.Fatalf("Error deleting kv store, %v", err)
	}
	runtime.GOMAXPROCS(goMaxProc)
}

func benchKvGet(size int, b *testing.B) {
	b.StopTimer()
	d := setUpKv(b)
	defer tearDownKv(d, b)

	keys := genKeys()
	value := genValue(size)
	for _, key := range keys {

		if err := d.Set([]byte(key), value); err != nil {
			b.Fatalf("Failed putting values, %v", err)
		}
	}

	shuffle(keys)
	b.SetBytes(int64(size))

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.Get([]byte(keys[i%len(keys)]), nil)
	}
	b.StopTimer()
}

func benchKvSet(size int, b *testing.B) {
	b.StopTimer()

	d := setUpKv(b)
	defer tearDownKv(d, b)

	keys := genKeys()
	value := genValue(size)
	shuffle(keys)
	b.SetBytes(int64(size))

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_ = d.Set([]byte(keys[i%len(keys)]), value)
	}
	b.StopTimer()
}

// Put
func Benchmark_KV_Put_32B(b *testing.B) {
	benchKvSet(32, b)
}

func Benchmark_KV_Put_1KB(b *testing.B) {
	benchKvSet(1024, b)
}

func Benchmark_KV_Put_4KB(b *testing.B) {
	benchKvSet(4096, b)
}

func Benchmark_KV_Put_10KB(b *testing.B) {
	benchKvSet(10240, b)
}

// Get
func Benchmark_KV_Get_32B(b *testing.B) {
	benchKvGet(32, b)
}

func Benchmark_KV_Get_1KB(b *testing.B) {
	benchKvGet(1024, b)
}

func Benchmark_KV_Get_4KB(b *testing.B) {
	benchKvGet(4096, b)
}

func Benchmark_KV_Get_10KB(b *testing.B) {
	benchKvGet(10240, b)
}
