package dskvs

import (
	"fmt"
	"time"
)

// A StoreError is returned when the store you try to use is not properly
// setup.
type StoreError struct {
	When time.Time
	What string
}

func (e StoreError) Error() string {
	return fmt.Sprintf("%v: %v", e.When, e.What)
}

func errorStoreNotLoaded(s *Store) error {
	return StoreError{
		time.Now(),

		fmt.Sprintf("Store with path <%s> has not yet been loaded",
			s.storagePath),
	}
}

// A PathError is returned when the path you provided is not suitable
// for storage, either because of its intrisic nature or because it is
// already in use by another storage.  In the latter case, you should
// verify your code to ensure that you are not forgetting a storage
// instance somewhere.
type PathError struct {
	When time.Time
	What string
}

func (e PathError) Error() string {
	return fmt.Sprintf("%v: %v", e.When, e.What)
}

func errorPathInUse(path string) error {
	return PathError{
		time.Now(),
		fmt.Sprintf("Path <%s> is already used by another Store", path),
	}
}

func errorPathInvalid(path string) error {
	return PathError{
		time.Now(),
		fmt.Sprintf("String <%s> is not a valid path", path),
	}
}

// A KeyError is returned when the key you provided is either not a valid
// key due to its nature, or is not appropriate for the method on which
// you use it.
type KeyError struct {
	When time.Time
	What string
}

func (e KeyError) Error() string {
	return fmt.Sprintf("%v: %v", e.When, e.What)
}

/*
 * Errors when the key is not valid
 */

func errorNoColl(key string) error {
	return KeyError{
		time.Now(),
		fmt.Sprintf("key<%s> has no collection identifier", key),
	}
}

func errorNoKey(key string) error {
	return KeyError{
		time.Now(),
		fmt.Sprintf("key<%s> has no member identifier", key),
	}
}

func errorEmptyKey() error {
	return KeyError{
		time.Now(),
		fmt.Sprintf("key is empty"),
	}
}

func errorNoSuchKey(key string) error {
	return KeyError{
		time.Now(),
		fmt.Sprintf("key<%s> has no value", key),
	}
}

func errorNoSuchColl(key string) error {
	return KeyError{
		time.Now(),
		fmt.Sprintf("key<%s> has no value", key),
	}
}

/*
 * Errors when the key implies the wrong method
 */

func errorGetIsColl(key string) error {
	return KeyError{
		time.Now(),
		fmt.Sprintf("key <%s> requested a Get on a collection",
			key),
	}
}

func errorGetAllIsNotColl(key string) error {
	return KeyError{
		time.Now(),
		fmt.Sprintf("key <%s> requested a GetAll for only a member",
			key),
	}
}

func errorPutIsColl(key, val string) error {
	return KeyError{
		time.Now(),
		fmt.Sprintf("key,val <%s,%s> requested a Put on a collection",
			key),
	}
}

func errorDeleteIsColl(key string) error {
	return KeyError{
		time.Now(),
		fmt.Sprintf("key <%s> requested a Delete on a collection",
			key),
	}
}

func errorDeleteAllIsNotColl(key string) error {
	return KeyError{
		time.Now(),
		fmt.Sprintf("key <%s> requested a DeleteAll for only a member",
			key),
	}
}
