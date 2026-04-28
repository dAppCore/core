// SPDX-License-Identifier: EUPL-1.2

// Hash operations for the Core framework.
// Provides crypto/sha256 wrappers so downstream packages can use core
// primitives for common digest operations.

package core

import (
	"crypto/hkdf"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"hash"
)

// SHA256 returns the SHA-256 digest of data.
//
//	sum := core.SHA256([]byte("hello"))
func SHA256(data []byte) [32]byte {
	return sha256.Sum256(data)
}

// SHA256Hex returns the SHA-256 digest of data as lowercase hexadecimal.
//
//	sum := core.SHA256Hex([]byte("hello"))
func SHA256Hex(data []byte) string {
	sum := SHA256(data)
	return hex.EncodeToString(sum[:])
}

// SHA256String returns the SHA-256 digest of s.
//
//	sum := core.SHA256String("hello")
func SHA256String(s string) [32]byte {
	return SHA256([]byte(s))
}

// SHA256HexString returns the SHA-256 digest of s as lowercase hexadecimal.
//
//	sum := core.SHA256HexString("hello")
func SHA256HexString(s string) string {
	return SHA256Hex([]byte(s))
}

// HMAC returns the HMAC digest for data using key and algo wrapped in a
// Result. OK=false with Code "crypto.algo.unsupported" when algo isn't
// "sha256" or "sha512".
//
//	r := core.HMAC("sha256", []byte("key"), []byte("payload"))
//	if r.OK { digest := r.Value.([]byte) }
func HMAC(algo string, key, data []byte) Result {
	factory := hashFor(algo)
	if factory == nil {
		return Result{Value: NewCode("crypto.algo.unsupported", Concat("unknown hash algorithm: ", algo)), OK: false}
	}
	mac := hmac.New(factory, key)
	mac.Write(data)
	return Result{Value: mac.Sum(nil), OK: true}
}

// HKDF derives length bytes from secret using salt, info, and algo
// wrapped in a Result. OK=false with Code "crypto.algo.unsupported"
// for unsupported algorithms or "crypto.hkdf.failed" on derivation
// failure.
//
//	r := core.HKDF("sha256", secret, salt, []byte("session"), 32)
//	if r.OK { sessionKey := r.Value.([]byte) }
func HKDF(algo string, secret, salt, info []byte, length int) Result {
	factory := hashFor(algo)
	if factory == nil {
		return Result{Value: NewCode("crypto.algo.unsupported", Concat("unknown hash algorithm: ", algo)), OK: false}
	}
	key, err := hkdf.Key(factory, secret, salt, string(info), length)
	if err != nil {
		return Result{Value: WrapCode(err, "crypto.hkdf.failed", "HKDF", "key derivation failed"), OK: false}
	}
	return Result{Value: key, OK: true}
}

func hashFor(algo string) func() hash.Hash {
	switch algo {
	case "sha256":
		return sha256.New
	case "sha512":
		return sha512.New
	default:
		return nil
	}
}
