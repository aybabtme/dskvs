package dskvs

import (
	"errors"
	"fmt"
	"time"
)

type KeyError struct {
	When time.Time
	What string
}

func (e KeyError) Error() string {
	return fmt.Sprintf("%v: %v", e.When, e.What)
}

func errorNoColl(key string) error {
	return KeyError{
		time.Now(),
		fmt.Sprintf("key<%s> has no collection identifier", key),
	}
}

func errorEmptyKey() error {
	return KeyError{
		time.Now(),
		fmt.Sprintf("key is empty"),
	}
}

func errorGetIsColl(key string) error {
	return KeyError{
		time.Now(),
		fmt.Sprintf("key <%s> requested a Get on a collection",
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
