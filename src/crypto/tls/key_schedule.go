// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tls

import (
	"crypto/ecdh"
	"encoding/binary"
	"errors"
	"hash"
	"io"
	"sync"

	"golang.org/x/crypto/hkdf"
)

// This file contains the functions necessary to compute the TLS 1.3 key
// schedule. See RFC 8446, Section 7.

const (
	resumptionBinderLabel         = "res binder"
	clientHandshakeTrafficLabel   = "c hs traffic"
	serverHandshakeTrafficLabel   = "s hs traffic"
	clientApplicationTrafficLabel = "c ap traffic"
	serverApplicationTrafficLabel = "s ap traffic"
	exporterLabel                 = "exp master"
	resumptionLabel               = "res master"
	trafficUpdateLabel            = "traffic upd"
)

var hkdfLabelPool = sync.Pool{
	New: func() any {
		return &[128]byte{}
	},
}

// expandLabel implements HKDF-Expand-Label from RFC 8446, Section 7.1.
func (c *cipherSuiteTLS13) expandLabel(secret []byte, label string, context []byte, length int) []byte {
	tmp := hkdfLabelPool.Get().(*[128]byte)
	defer hkdfLabelPool.Put(tmp)

	hkdfLabel := binary.BigEndian.AppendUint16(tmp[:0], uint16(length))
	hkdfLabel = append(hkdfLabel, uint8(len("tls13 ")+len(label)))
	hkdfLabel = append(hkdfLabel, []byte("tls13 ")...)
	hkdfLabel = append(hkdfLabel, label...)
	hkdfLabel = append(hkdfLabel, uint8(len(context)))
	hkdfLabel = append(hkdfLabel, context...)

	out := make([]byte, length)
	h := c.hkdfPool.Get().(*hkdf.HKDF)
	defer c.hkdfPool.Put(h)
	n, err := h.Expand(secret, hkdfLabel).Read(out)
	if err != nil || n != length {
		panic("tls: HKDF-Expand-Label invocation failed unexpectedly")
	}
	return out
}

// deriveSecret implements Derive-Secret from RFC 8446, Section 7.1.
func (c *cipherSuiteTLS13) deriveSecret(secret []byte, label string, transcript hash.Hash) []byte {
	if transcript == nil {
		transcript = c.hash.New()
	}
	return c.expandLabel(secret, label, transcript.Sum(nil), c.hash.Size())
}

// extract implements HKDF-Extract with the cipher suite hash.
func (c *cipherSuiteTLS13) extract(newSecret, currentSecret []byte) []byte {
	if newSecret == nil {
		newSecret = make([]byte, c.hash.Size())
	}
	h := c.hkdfPool.Get().(*hkdf.HKDF)
	defer c.hkdfPool.Put(h)
	return h.Extract(newSecret, currentSecret)
}

// nextTrafficSecret generates the next traffic secret, given the current one,
// according to RFC 8446, Section 7.2.
func (c *cipherSuiteTLS13) nextTrafficSecret(trafficSecret []byte) []byte {
	return c.expandLabel(trafficSecret, trafficUpdateLabel, nil, c.hash.Size())
}

// trafficKey generates traffic keys according to RFC 8446, Section 7.3.
func (c *cipherSuiteTLS13) trafficKey(trafficSecret []byte) (key, iv []byte) {
	key = c.expandLabel(trafficSecret, "key", nil, c.keyLen)
	iv = c.expandLabel(trafficSecret, "iv", nil, aeadNonceLength)
	return
}

// finishedHash generates the Finished verify_data or PskBinderEntry according
// to RFC 8446, Section 4.4.4. See sections 4.4 and 4.2.11.2 for the baseKey
// selection.
func (c *cipherSuiteTLS13) finishedHash(baseKey []byte, transcript hash.Hash) []byte {
	finishedKey := c.expandLabel(baseKey, "finished", nil, c.hash.Size())
	verifyData := c.hmacPool.New(finishedKey)
	defer c.hmacPool.Put(verifyData)
	verifyData.Write(transcript.Sum(nil))
	return verifyData.Sum(nil)
}

// exportKeyingMaterial implements RFC5705 exporters for TLS 1.3 according to
// RFC 8446, Section 7.5.
func (c *cipherSuiteTLS13) exportKeyingMaterial(masterSecret []byte, transcript hash.Hash) func(string, []byte, int) ([]byte, error) {
	expMasterSecret := c.deriveSecret(masterSecret, exporterLabel, transcript)
	return func(label string, context []byte, length int) ([]byte, error) {
		secret := c.deriveSecret(expMasterSecret, label, nil)
		h := c.hashPool.New()
		defer c.hashPool.Put(h)

		h.Write(context)
		return c.expandLabel(secret, "exporter", h.Sum(nil), length), nil
	}
}

// generateECDHEKey returns a PrivateKey that implements Diffie-Hellman
// according to RFC 8446, Section 4.2.8.2.
func generateECDHEKey(rand io.Reader, curveID CurveID) (*ecdh.PrivateKey, error) {
	curve, ok := curveForCurveID(curveID)
	if !ok {
		return nil, errors.New("tls: internal error: unsupported curve")
	}

	return curve.GenerateKey(rand)
}

func curveForCurveID(id CurveID) (ecdh.Curve, bool) {
	switch id {
	case X25519:
		return ecdh.X25519(), true
	case CurveP256:
		return ecdh.P256(), true
	case CurveP384:
		return ecdh.P384(), true
	case CurveP521:
		return ecdh.P521(), true
	default:
		return nil, false
	}
}

func curveIDForCurve(curve ecdh.Curve) (CurveID, bool) {
	switch curve {
	case ecdh.X25519():
		return X25519, true
	case ecdh.P256():
		return CurveP256, true
	case ecdh.P384():
		return CurveP384, true
	case ecdh.P521():
		return CurveP521, true
	default:
		return 0, false
	}
}
