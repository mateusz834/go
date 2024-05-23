// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rand

import (
	"internal/chacha8rand"
)

// A ChaCha8 is a ChaCha8-based cryptographically strong
// random number generator.
type ChaCha8 struct {
	state chacha8rand.State
}

// NewChaCha8 returns a new ChaCha8 seeded with the given seed.
func NewChaCha8(seed [32]byte) *ChaCha8 {
	c := new(ChaCha8)
	c.state.Init(seed)
	return c
}

// Seed resets the ChaCha8 to behave the same way as NewChaCha8(seed).
func (c *ChaCha8) Seed(seed [32]byte) {
	c.state.Init(seed)
}

// Uint64 returns a uniformly distributed random uint64 value.
func (c *ChaCha8) Uint64() uint64 {
	for {
		x, ok := c.state.Next()
		if ok {
			return x
		}
		c.state.Refill()
	}
}

// Read reads exactly len(p) bytes into p.
// It always returns len(p) and a nil error.
//
// If calls to Read and Uint64 are interleaved, the order in which bits are
// returned by the two is undefined, and Read may return bits generated before
// the last call to Uint64.
func (c *ChaCha8) Read(p []byte) (n int, err error) {
	c.state.FillRand(p)
	return len(p), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (c *ChaCha8) UnmarshalBinary(data []byte) error {
	return chacha8rand.Unmarshal(&c.state, data)
}

func cutPrefix(s, prefix []byte) (after []byte, found bool) {
	if len(s) < len(prefix) || string(s[:len(prefix)]) != string(prefix) {
		return s, false
	}
	return s[len(prefix):], true
}

func readUint8LengthPrefixed(b []byte) (buf, rest []byte, ok bool) {
	if len(b) == 0 || len(b) < int(1+b[0]) {
		return nil, nil, false
	}
	return b[1 : 1+b[0]], b[1+b[0]:], true
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (c *ChaCha8) MarshalBinary() ([]byte, error) {
	return chacha8rand.Marshal(&c.state), nil
}
