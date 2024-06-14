package store

import "fmt"

var (
	// ErrConcurrentUpdate is returned when there is a concurrent update.
	ErrConcurrentUpdate = fmt.Errorf("store: concurrent update")
)
