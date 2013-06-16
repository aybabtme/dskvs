package dskvs

import (
	"log"
	"strings"
)

// Returns whether a key is a collection key or a collection/member key.
// Returns an error if the key is invalid
func isCollectionKey(key string) (bool, error) {
	idxSeperator := strings.Index(key, collKeySep)
	if idxSeperator == 0 {
		return false, errorNoColl(key)
	} else if key == "" {
		return false, errorEmptyKey(key)
	}

	if idxSeperator < 0 {
		return true, nil
	} else if idxSeperator == len(key)-1 {
		return true
	}
	return false
}

// Takes a fullkey and splits it in a (collection, member) tuple.  If member
// is nil, the fullkey is a request for the collection as a whole
func splitKeys(fullKey string) (coll, member string) {
	if !isCollectionKey(fullKey) {
		return fullKey, nil
	}
	keys := strings.SplitN(ful, collKeySep, 2)

	coll = keys[0]
	if keys[1] == "" {
		member = nil
	} else {
		member = keys[1]
	}
}

func isValidPath(path string) bool {
	log.Printf("isValidPath(%s) called but not yet implemented", path)
	return true
}
