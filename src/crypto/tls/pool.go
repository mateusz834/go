package tls

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"hash"
	"sync"

	"golang.org/x/crypto/hkdf"
)

var (
	pool128B   = &sync.Pool{New: func() any { return &[128]byte{} }}
	poolBuffer = &sync.Pool{New: func() any { return &bytes.Buffer{} }}

	sha256Pool     = &hashPool{new: sha256.New}
	sha384Pool     = &hashPool{new: sha512.New384}
	hmacSHA256Pool = &hmacPool{new: sha256Pool.New}
	hmacSHA384Pool = &hmacPool{new: sha384Pool.New}
	hkdfSHA256Pool = &sync.Pool{New: func() any { return hkdf.NewHKDF(sha256Pool.New) }}
	hkdfSHA384Pool = &sync.Pool{New: func() any { return hkdf.NewHKDF(sha384Pool.New) }}
)

type hmacPool struct {
	pool sync.Pool
	new  func() hash.Hash
}

func (h *hmacPool) New(key []byte) hash.Hash {
	if hmac := h.pool.Get(); hmac != nil {
		hm := hmac.(hash.Hash)
		hm.(interface{ ResetKey(key []byte) }).ResetKey(key)
		return hm
	}
	return hmac.New(h.new, key)
}

func (h *hmacPool) Put(hmac hash.Hash) {
	h.pool.Put(hmac)
}

type hashPool struct {
	pool sync.Pool
	new  func() hash.Hash
}

func (h *hashPool) New() hash.Hash {
	if hmac := h.pool.Get(); hmac != nil {
		hm := hmac.(hash.Hash)
		hm.Reset()
		return hm
	}
	return h.new()
}

func (h *hashPool) Put(hash hash.Hash) {
	h.pool.Put(hash)
}
