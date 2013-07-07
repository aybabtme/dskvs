package dskvs

import (
	"fmt"
)

// A StoreError is returned when the store you try to use is not properly
// setup.
type StoreError struct {
	What string
}

func (e StoreError) Error() string {
	return e.What
}

func errorStoreNotLoaded() error {
	return StoreError{
		"Store is nil",
	}
}

// A FileError is returned when a file that was read failed to return
// expected data
type FileError struct {
	What     string
	Filename string
}

func (e FileError) Error() string {
	return fmt.Sprintf("%v, file=%s", e.What, e.Filename)
}

func errorFailedChecksum(name string) error {
	return FileError{
		"Failed checksum check, file might be corrupted, or not a dskvs file",
		name,
	}
}

func errorPayloadWrongSize(name string, expected uint64, actual int) error {
	return FileError{
		fmt.Sprintf("Payload should have length %d but was %d",
			expected, actual),
		name,
	}
}

func errorWrongVersion(major, minor uint16, patch uint64) error {
	return FileError{
		fmt.Sprintf("File version incompatible with this version (%d.%d.%d)",
			MAJOR_VERSION, MINOR_VERSION, PATCH_VERSION),
		fmt.Sprintf("given version %d.%d.%d", major, minor, patch),
	}
}

func errorCreatingHeader(name string, err error) error {
	return FileError{
		fmt.Sprintf("Error creating header, received error <%v>", err),
		name,
	}
}

// A PathError is returned when the path you provided is not suitable
// for storage, either because of its intrisic nature or because it is
// already in use by another storage.  In the latter case, you should
// verify your code to ensure that you are not forgetting a storage
// instance somewhere.
type PathError struct {
	What string
	Path string
}

func (e PathError) Error() string {
	return fmt.Sprintf("%v, path=%v", e.What, e.Path)
}

func errorPathInUse(path string) error {
	return PathError{
		"Path is already used by another Store",
		path,
	}
}

func errorPathInvalid(path string) error {
	return PathError{
		"String is not a valid path",
		path,
	}
}

// A KeyError is returned when the key you provided is either not a valid
// key due to its nature, or is not appropriate for the method on which
// you use it.
type KeyError struct {
	What string
	Key  string
}

func (e KeyError) Error() string {
	return fmt.Sprintf("%v, key=%v", e.What, e.Key)
}

/*
 * Errors when the key is not valid
 */

func errorNoColl(key string) error {
	return KeyError{
		"key has no collection identifier",
		key,
	}
}

func errorEmptyKey() error {
	return KeyError{
		"key is empty",
		"<empty string>",
	}
}

func errorNoSuchKey(key string) error {
	return KeyError{
		"key holds no value in this store",
		key,
	}
}

func errorNoSuchColl(key string) error {
	return KeyError{
		"key does not represent a collection in this store",
		key,
	}
}

/*
 * Errors when the key implies the wrong method
 */

func errorGetIsColl(key string) error {
	return KeyError{
		"key requested a Get on a collection, wrong method",
		key,
	}
}

func errorGetAllIsNotColl(key string) error {
	return KeyError{
		"key requested a GetAll for only a member, wrong method",
		key,
	}
}

func errorPutIsColl(key, val string) error {
	return KeyError{
		"<key,val> requested a Put on a collection, wrong method",
		fmt.Sprintf("<%s,%s>", key, val),
	}
}

func errorDeleteIsColl(key string) error {
	return KeyError{
		"key requested a Delete on a collection, wrong method",
		key,
	}
}

func errorDeleteAllIsNotColl(key string) error {
	return KeyError{
		"key requested a DeleteAll for only a member, wrong method",
		key,
	}
}
