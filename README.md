# DSKVS

[![Build Status](https://drone.io/github.com/aybabtme/dskvs/status.png)]
(https://drone.io/github.com/aybabtme/dskvs/latest)
[![Coverage
Status](https://coveralls.io/repos/aybabtme/dskvs/badge.png?branch=master)](https://coveralls.io/r/aybabtme/dskvs?branch=master)

An in-memory, eventually persisted data store of the simplest design, `dskvs`
stores it's data in two layers of maps, which are routinely persisted to disk
by a janitor.

`dskvs` stands for Dead Simple Key-Value Store.  The aim of this library is to
provide storage solution for _Small Data_™ apps.  If your data-set holds within
the RAM of a single host, `dskvs` is the right thing for you.

## Status
Holds within 952 lines of Go code, tested by an extra 1k lines.

This project is not yet ready for anybody's usage, it is still under development and has not been extensively tested.


## At a glance

```go
// Create a store
store := dskvs.NewStore("/home/aybabtme/music")

// Create persistance artifacts and loads data already on disk
store.Load()

// Get
value, err := store.Get("artist/daft_punk")

// GetAll
values, err := store.GetAll("artist")

// Put
oldValue, err := store.Put("artist/daft_punk", "{ quality:'epic' }")

// Delete
err := store.Delete("artist/celine_dion")

// Delete all
err := store.DeleteAll("artist")

// Finish writing, close files.
store.Close()
```

There is currently no support for replication of any sort.  There are already
dozens of highly specialized data-store providing this sort of features,
`dskvs` is not one of them.

## Usage
`go get` the master branch.

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
acceptable for now.
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

God help you if you load two `Store`s that share some part of the filesystem.
A check for that is not implemented.

I might add an unsafe version of `Store` in the future if there's evidence of
notable performance gains, if there's a use case for it and if it doesn't
uglify the code.

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
The basic API for `dskvs` should be `Get`, `Put` and `Delete`.  However, since
`dskvs` supports 'collection/members', there's no practical way to query for
all members, while `dskvs` has facility to act on aggregates.  So it makes
sense for `dskvs` to provide `GetAll` and `DeleteAll` since it has better
visibility on what `All` represents in those cases.

This is not true for `PutAll`.  There's no added value in having `dskvs`
`Put` all your values for you, one by one, as it would results in it simply
calling `Put` on all your values.

## License
An MIT license, see the LICENSE file.
