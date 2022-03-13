// Add/Draw byffers from a pool of byte buggers
// Copied and modified from "nhooyr.io/websocket/internal/bpool"
package server

import (
	"bytes"
	"sync"
)

var bpool sync.Pool

// Get returns a buffer from the pool or creates a new one if
// the pool is empty
func getBuf() *bytes.Buffer {
	b := bpool.Get()
	if b == nil {
		return &bytes.Buffer{}
	}
	return b.(*bytes.Buffer)
}

// Put returns a buffer into the pool
func putBuf(b *bytes.Buffer) {
	b.Reset()
	bpool.Put(b)
}
