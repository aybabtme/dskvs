/*

Package dskvs is an embedded, in memory key value store following the REST
convention of adressing resources as being 'collections' and 'members'.

The main features of dskvs are:

	- very high read performance
	- high write performance
	- safe for concurrent use by multiple readers/writers.
	- every read/write is done in memory.  Writes are persisted asynchronously
	- remains usable under heavy concurrent usage, although throughput is not-optimal
	- can load []byte values of any size
	- persisted files are always verified for consistency

In dskvs, a collection contains many members, a member contains a value. A full
key is a combination of both the collection and member identifiers to a value.

Example:

	fullkey := "artist" + CollKeySep + "Daft Punk"

In this example, 'artist' is a collection that contains many artists.  'Daft Punk'
is one of those artists.

Example:

	fullkey := "artist" + CollKeySep + "Daft Punk" + CollKeySep + "Discovery.."


A fullkey can contain many CollKeySep; only the first encountered is considered
for the collection name.

Every entry of dskvs is saved as a file under a path.  If you tell dskvs to use
the base path "/home/aybabtme/dskvs", it will prepare a filename for your key
such as :

	colletion := "This Collection"
	key       := "Is the Greatest!!!"

	escapedKey   := filepathSafe(key)  // escapes all dangerous characters
	truncatedKey := escapedKey[:40]    // truncate the key to 40 runes

	member := truncatedKey + sha1(key) // append the truncated key to the SHA1 of
	                                   // the original key to prevent collisions

dskvs will then write the entry at the path :

	/home/aybabtme/dskvs/<collection>/<member>

Example:

	fullkey    := "artist" + CollKeySep + "Daft Punk" + CollKeySep + "Discovery.."

	collection := "artist"
	member     := "Daft+Punk%2FDiscovery..446166742050756e6b2f446973636f766572792e2eda39a3ee5e6b4b0d3255bfef95601890afd80709"


Given that keys have their value escaped, you can safely use any key:

	fullkey := "My Collection/../../"

Will yield :

	collection == "My Collection"
	member == "..%2F..%2F2e2e2f2e2e2fda39a3ee5e6b4b0d3255bfef95601890afd80709"
*/
package dskvs

const (
	// MAJOR_VERSION is used to ensure that incompatible fileformat versions are
	// not loaded in memory.
	MAJOR_VERSION uint16 = 0
	// MINOR_VERSION is used to differentiate between fileformat versions. It might
	// be used for migrations if a future change to dskvs breaks the original
	// fileformat contract
	MINOR_VERSION uint16 = 2
	// PATCH_VERSION is used for the same reasons as MINOR_VERSION
	PATCH_VERSION uint64 = 1
)
