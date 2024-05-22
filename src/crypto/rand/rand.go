// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package rand implements a cryptographically secure
// random number generator.
package rand

import "io"

// Reader is a global, shared instance of a cryptographically
// secure random number generator.
//
//   - On Linux, FreeBSD, Dragonfly, and Solaris, Reader uses getrandom(2)
//     if available, and /dev/urandom otherwise.
//   - On macOS and iOS, Reader uses arc4random_buf(3).
//   - On OpenBSD and NetBSD, Reader uses getentropy(2).
//   - On other Unix-like systems, Reader reads from /dev/urandom.
//   - On Windows, Reader uses the ProcessPrng API.
//   - On js/wasm, Reader uses the Web Crypto API.
//   - On wasip1/wasm, Reader uses random_get from wasi_snapshot_preview1.
var Reader io.Reader = randReader

// Read is a helper function that reads data from the [Reader] and populates
// the entire out byte slice with cryptographically secure random data.
// It has the same behaviour as calling io.ReadFull with the [Reader].
// On return, n == len(b) if and only if err == nil.
func Read(out []byte) (n int, err error) {
	// To avoid escaping the out slice, inline the io.ReadFull function.
	// The following code has the same behaviour as: io.ReadFull(Reader, out).
	for n < len(out) && err == nil {
		var nn int
		nn, err = randReader.Read(out[n:])
		n += nn
	}
	if n >= len(out) {
		err = nil
	} else if n > 0 && err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	return
}
