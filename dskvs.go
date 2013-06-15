
package dskvs

import "fmt"

func Get(key string) (string, error) {

	return nil, nil
}

// Puts the given value into the key location.  key should be a member, not a // collection
func Put(key, value string) error {
	return nil
}

// Deletes a key from the storage.  If the key covers a collection, the whole // collection will be deleted.
func Delete(key string) error {
	return nil
}
