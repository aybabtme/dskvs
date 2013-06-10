# A Dead Simple Key-Value Store

An in-memory, eventually persisted data store of the simplest design, `dskvs` stores it's data in two layers of maps, which are routinely persisted to disk by a janitor.

`dskvs` stands for _D_ead _S_imple _K_ey-_V_alue _S_tore.  The aim of this library is to provide storage solution for Small-Data (tm) apps.  If your data-set holds within the RAM of a single host.

```go
// Get
value, err := dskvs.Get("artist/daft_punk")

// Set or Update
oldValue, err := dksvs.Set("artist/daft_punk", "{ quality:'epic' }")

// List
values, err := dskvs.List("artist")

// Delete
err := dksvs.Delete("artist/celine_dion")

// Delete all
err := dksvs.DeleteAll("artist")
```

There is currently no support for replication of any sort.  There are already dozens of highly specialized data-store providing this sort of features, `dskvs` is not one of them.

## Status
This project is not yet ready for any real use.

## Usage
'go get' the master branch.

To use the library, import it:
```go
import "github.com/aybabtme/dskvs"
```
Then start using the `dskvs` package.

Otherwise, fork this repo and `go get` your fork.  Also update your import string.  If you make improvements or fix issues, please do submit a pull-request.

## Eventually persisted ?
The term is a pun on 'eventual consistency', but has nothing to do with the CAP theorem.  This is not a distributed data store.

`dskvs` serves all its request from memory.  All the writes and reads are responded for from memory.  When a write happens on a key-value, the key is flagged as dirty and a janitor goroutine will pick it up as soon as possible and persist it to disk, whenever that happens to be.

Usually, _eventual-persistance_ means that it will be persisted ASAP, but within a couple of µ-seconds.  Meanwhile, any read subsequent to the write will be correct.

## License
An MIT license, see LICENSE.md
