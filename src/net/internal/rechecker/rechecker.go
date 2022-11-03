// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rechecker

import (
	"io"
	"os"
	"sync/atomic"
	"time"

	"sync"
)

type value[T any] struct {
	v   *T
	err error
}

type Rechecker[T any] struct {
	File             string
	Parse            func(content []byte) (*T, error)
	FileErrorHandler func(err error) (*T, error)
	Duration         time.Duration

	val         atomic.Pointer[value[T]]
	once        sync.Once
	recheckSema atomic.Bool
	lastCheched time.Time
	modTime     time.Time
}

func (r *Rechecker[T]) Get() (*T, error) {
	var val *value[T]

	r.once.Do(func() {
		val = &value[T]{}
		val.v, r.modTime, val.err = r.initialFileParse()
		r.lastCheched = time.Now()
		r.val.Store(val)
	})

	if val != nil {
		return val.v, val.err
	}

	val = r.val.Load()

	// one goroutine at a time
	if r.recheckSema.CompareAndSwap(false, true) {
		defer r.recheckSema.Store(false)

		now := time.Now()
		if now.After(r.lastCheched.Add(r.Duration)) {
			r.lastCheched = now

			stat, err := os.Stat(r.File)
			if err != nil {
				newVal, err := r.fileErrHandle(err)
				val = &value[T]{v: newVal, err: err}
				r.val.Store(val)
				return newVal, err
			}

			if !stat.ModTime().Equal(r.modTime) {
				val = &value[T]{}
				val.v, val.err = r.recheckParse()
				r.modTime = stat.ModTime()
				r.val.Store(val)
				return val.v, val.err
			}
		}
	}

	return val.v, val.err
}

func (r *Rechecker[T]) fileErrHandle(inErr error) (val *T, err error) {
	if r.FileErrorHandler == nil {
		return nil, inErr
	}
	return r.FileErrorHandler(inErr)
}

func (r *Rechecker[T]) recheckParse() (val *T, err error) {
	f, err := os.OpenFile(r.File, os.O_RDONLY, 0)
	if err != nil {
		return r.fileErrHandle(err)
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return r.fileErrHandle(err)
	}

	return r.Parse(data)
}

func (r *Rechecker[T]) initialFileParse() (val *T, modTime time.Time, err error) {
	f, err := os.OpenFile(r.File, os.O_RDONLY, 0)
	if err != nil {
		val, err = r.fileErrHandle(err)
		return val, time.Time{}, err
	}

	stat, err := f.Stat()
	if err != nil {
		val, err = r.fileErrHandle(err)
		return val, time.Time{}, err
	}

	data, err := io.ReadAll(f)
	if err != nil {
		val, err = r.fileErrHandle(err)
		return val, time.Time{}, err
	}

	val, err = r.Parse(data)
	return val, stat.ModTime(), err
}

// ChangeFile should be used ONLY inside tests.
func (r *Rechecker[T]) ChangeFile(file string, lastChecked time.Time) bool {
	r.Get() // call Get(), so that the r.once is compleated.

	for i := 0; i < 10; i++ {
		if r.recheckSema.CompareAndSwap(false, true) {
			defer r.recheckSema.Store(false)
			r.File = file
			val := &value[T]{}
			val.v, r.modTime, val.err = r.initialFileParse()
			r.lastCheched = lastChecked
			r.val.Store(val)
			return true
		}
	}
	return false
}
