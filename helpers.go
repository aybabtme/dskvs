package dskvs

import (
	"log"
	"strings"
)

const CollKeySep = "/"

func checkKeyValid(key string) error {
	idxSeperator := strings.Index(key, CollKeySep)
	if idxSeperator == 0 {
		return errorNoColl(key)
	} else if key == "" {
		return errorEmptyKey()
	}
	return nil
}

// Returns whether a key is a collection key or a collection/member key.
// Returns an error if the key is invalid
func isCollectionKey(key string) bool {
	idxSeperator := strings.Index(key, CollKeySep)
	if idxSeperator < 0 {
		return true
	} else if idxSeperator == len(key)-1 {
		return true
	}
	return false
}

// Takes a fullkey and splits it in a (collection, member) tuple.  If member
// is nil, the fullkey is a request for the collection as a whole
func splitKeys(fullKey string) (string, string, error) {
	if isCollectionKey(fullKey) {
		return "", "", errorNoKey(fullKey)
	}

	keys := strings.SplitN(fullKey, CollKeySep, 2)

	return keys[0], keys[1], nil
}

func isValidPath(path string) bool {
	log.Printf("isValidPath(%s) called but not yet implemented", path)
	return true
}

func expandPath(path string) string {
	log.Printf("expandPath(%s) called but not yet implemented", path)
	return ""
}
