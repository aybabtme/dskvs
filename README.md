I don't recommend you use this. Was just a toy.

# DSKVS

[![Build Status](https://drone.io/github.com/aybabtme/dskvs/status.png)]
(https://drone.io/github.com/aybabtme/dskvs/latest)
[![Coverage
Status](https://coveralls.io/repos/aybabtme/dskvs/badge.png?branch=master)](https://coveralls.io/r/aybabtme/dskvs?branch=master)

An in-memory, eventually persisted data store of the simplest design, `dskvs`
stores its data in two layers of maps, which are routinely persisted to disk
by a janitor.

`dskvs` stands for Dead Simple Key-Value Store.  The aim of this library is to
provide storage solution for _Small Data_™ apps.  If your data set holds within
the RAM of a single host, `dskvs` is the right thing for you.

## Status
Test coverage is good, but this is alpha release. Use at your own risks.

Future plans are to build a few apps using `dskvs` to provide a proof-of-concept
for the current features.

## At a glance
`Put` and `Get` arbitrarily sized slices of bytes.  _Arbitrarily_ is a word that
holds for as long as you have free RAM (=

```go
// Open a store at the given path.  Existing artifacts are loaded in memory
store, err := dskvs.Open("/home/aybabtme/music")

// Get
value, err := store.Get("artist/daft_punk")

// GetAll
values, err := store.GetAll("artist")

// Put
err := store.Put("artist/daft_punk", []byte("{ quality:'epic' }"))

// Delete
err := store.Delete("artist/celine_dion")

// Delete all
err := store.DeleteAll("artist")

// Finish persisting changes, then close the store.
err := store.Close()
```

There is no support for replication of any sort, and there won't be. There are already
dozens of highly specialized data store providing this sort of features,
`dskvs` is not one of them. `dskvs` is an embedded data store.

## Documentation

Look it up on [GoDoc](http://godoc.org/github.com/aybabtme/dskvs)

## Usage

Verify that you have a recent version of Go (>= 1.1):

```
$ go version
```

Then `go get github.com/aybabtme/dskvs`.

To use the library, import it:
```go
import "github.com/aybabtme/dskvs"
```
Then start using the `dskvs` package.

Otherwise, fork this repo and `go get` your fork.  Also update your import
string.  If you make improvements or fix issues, please do submit a
pull-request.

## Performance

`dskvs` is not optimized and requires much work.  However, performance is
acceptable for now.  The following is the results of 10K read/writes by 10
goroutines.  The results for a single goroutine at a time are much higher,
but somehow meaningless in Go World.
```
2.3 GHz Intel Core i7, 8GB 1600 MHz DDR3, OS X 10.8.4
Concurrency Benchmark - Goroutines=10, unique key-value=2048
Sequence of 10000 bytes operations by group, 10 concurrent request in each groups

Put - first time
N=2048,
	 bandwidth : 7.7 MB/s	 rate :       766 qps
	 min   =   256.121us	 max   =  4.984934ms
	 avg   =  1.305153ms	 med   =  1.194395ms
	 p75   =  1.329744ms	 p90   =  1.596718ms
	 p99   =  4.650752ms	 p999  =  4.977261ms
	 p9999 =  4.984934ms
Put - rewrite
N=2048,
	 bandwidth : 8.7 MB/s	 rate :       869 qps
	 min   =     4.484us	 max   =  9.908242ms
	 avg   =  1.150016ms	 med   =   907.134us
	 p75   =  1.023127ms	 p90   =  1.251781ms
	 p99   =  7.656514ms	 p999  =  9.806725ms
	 p9999 =  9.908242ms
Get
N=2048,
	 bandwidth :  11 GB/s	 rate : 1,072,961 qps
	 min   =       216ns	 max   =     9.943us
	 avg   =       932ns	 med   =       899ns
	 p75   =     1.092us	 p90   =     1.289us
	 p99   =     1.735us	 p999  =     3.324us
	 p9999 =     9.943us
Delete
N=2048,
	 bandwidth :  17 MB/s	 rate :     1,745 qps
	 min   =     4.245us	 max   = 22.290594ms
	 avg   =   572.931us	 med   =   446.827us
	 p75   =   532.494us	 p90   =   677.701us
	 p99   =  1.157741ms	 p999  = 22.145781ms
	 p9999 = 22.290594ms
by 8 cpus, using 10 concurrent goroutines

```


## Concurrency
`Store` instances are safe for concurrent use.  You can create stores
concurrently, read and write to stores concurrently.  Safe concurrent access
are part of the implementation because `dskvs` is expected to be used for
concurrent apps.

God help you if you load two `Store` that share some part of their filepath.
Just don't do it.  Two `Store` share a similar path if they have any common files
in their file tree.

## Eventually persisted ?
The term is a pun on 'eventual consistency', but has nothing to do with the
CAP theorem.  This is not a distributed data store.

`dskvs` serves all its request from memory.  All the writes and reads are
responded for from memory.  When a write happens on a key-value, the key is
flagged as dirty and a janitor goroutine will pick it up as soon as possible
and persist it to disk, whenever that happens to be.

Usually, _eventual-persistance_ means that it will be persisted ASAP, but
within a couple of µ-seconds.  Meanwhile, any read subsequent to the write
will be correct, as they are served from memory.

See `dskvs` as big cache that happens to be backed up to disk very frequently.

## Not `PutAll` ?
A `PutAll` method would simply call `Put` for every entry if your slice.  There
is no _special_ way to optimize a `PutAll` to perform better than as many `Put`
calls, so it was not added to the API.

There are good ways to optimize `GetAll` and `DeleteAll`, which explains their
presence and the incongruence of a missing `PutAll`.

## License

An MIT license, see the LICENSE file.
