# DSKVS

[![Build Status](https://drone.io/github.com/aybabtme/dskvs/status.png)]
(https://drone.io/github.com/aybabtme/dskvs/latest)
[![Coverage
Status](https://coveralls.io/repos/aybabtme/dskvs/badge.png?branch=proto)](https://coveralls.io/r/aybabtme/dskvs?branch=proto)

An in-memory, eventually persisted data store of the simplest design, `dskvs`
stores it's data in two layers of maps, which are routinely persisted to disk
by a janitor.

`dskvs` stands for Dead Simple Key-Value Store.  The aim of this library is to
provide storage solution for _Small Data_™ apps.  If your data-set holds within
the RAM of a single host, `dskvs` is the right thing for you.

## Status
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
Goroutines=1, unique key-value=2048, 100 bytes operations
Put - first time
N=2048,
	 8'905 op/sec
	 min   = 29.472us
	 max   = 4.676918ms
	 avg   = 112.285us
	 med   = 97.672us
	 p75   = 109.408us
	 p90   = 134.222us
	 p99   = 189.819us
	 p999  = 2.421366ms
	 p9999 = 4.676918ms
Put - rewrite
N=2048,
	 10'226 op/sec
	 min   = 3.453us
	 max   = 9.886882ms
	 avg   = 97.786us
	 med   = 67.783us
	 p75   = 85.199us
	 p90   = 105.429us
	 p99   = 207.353us
	 p999  = 6.73288ms
	 p9999 = 9.886882ms
Get
N=2048,
	 2'557'544 op/sec
	 min   = 189ns
	 max   = 2.524us
	 avg   = 391ns
	 med   = 318ns
	 p75   = 401ns
	 p90   = 567ns
	 p99   = 1.544us
	 p999  = 2.279us
	 p9999 = 2.524us
Delete
N=2048,
	 20'096 op/sec
	 min   = 2.1us
	 max   = 16.774411ms
	 avg   = 49.76us
	 med   = 35.834us
	 p75   = 48.358us
	 p90   = 59.506us
	 p99   = 111.642us
	 p999  = 617.158us
	 p9999 = 16.774411ms
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
