package benchmark

import (
	"github.com/peterbourgon/diskv"
	"testing"
)

func load(d *diskv.Diskv, keys []string, val []byte) {
	for _, key := range keys {
		d.Write(key, val)
	}
}

func benchRead(b *testing.B, size, cachesz int) {
	b.StopTimer()
	d := diskv.New(diskv.Options{
		BasePath:     "speed-test",
		Transform:    func(string) []string { return []string{} },
		CacheSizeMax: uint64(cachesz),
	})
	defer d.EraseAll()

	keys := genKeys()
	value := genValue(size)
	load(d, keys, value)
	shuffle(keys)
	b.SetBytes(int64(size))

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.Read(keys[i%len(keys)])
	}
	b.StopTimer()
}

func benchWrite(b *testing.B, size int) {
	b.StopTimer()

	options := diskv.Options{
		BasePath:     "speed-test",
		Transform:    func(string) []string { return []string{} },
		CacheSizeMax: 0,
	}

	d := diskv.New(options)
	defer d.EraseAll()
	keys := genKeys()
	value := genValue(size)
	shuffle(keys)
	b.SetBytes(int64(size))

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		d.Write(keys[i%len(keys)], value)
	}
	b.StopTimer()
}

func Benchmark_Diskv_Put_32B(b *testing.B) {
	benchWrite(b, 32)
}

func Benchmark_Diskv_Put_1KB(b *testing.B) {
	benchWrite(b, 1024)
}

func Benchmark_Diskv_Put_4KB(b *testing.B) {
	benchWrite(b, 4096)
}

func Benchmark_Diskv_Put_10KB(b *testing.B) {
	benchWrite(b, 10240)
}

func Benchmark_Diskv_Get_32B(b *testing.B) {
	benchRead(b, 32, KEY_COUNT*32*2)
}

func Benchmark_Diskv_Get_1KB(b *testing.B) {
	benchRead(b, 1024, KEY_COUNT*1024*2)
}

func Benchmark_Diskv_Get_4KB(b *testing.B) {
	benchRead(b, 4096, KEY_COUNT*4096*2)
}

func Benchmark_Diskv_Get_10KB(b *testing.B) {
	benchRead(b, 10240, KEY_COUNT*4096*2)
}
