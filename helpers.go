package dskvs

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

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
func splitKeys(fullKey string) (string, string) {
	idx := strings.Index(fullKey, CollKeySep)
	return fullKey[:idx], fullKey[idx:]
}

func isValidPath(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Printf("Could not get absolute filepath %v", err)
		return false
	}

	stat, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return true
		} else {
			log.Printf("Could not get stat %v", err)
			return false
		}
	}

	return stat.IsDir()
}

func expandPath(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}

	return absPath
}
