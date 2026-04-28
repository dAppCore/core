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

// HMAC returns the HMAC digest for data using key and algo.
// Supported algorithms are "sha256" and "sha512"; unknown algorithms panic.
//
//	digest := core.HMAC("sha256", []byte("key"), []byte("payload"))
func HMAC(algo string, key, data []byte) []byte {
	mac := hmac.New(hashFor(algo), key)
	mac.Write(data)
	return mac.Sum(nil)
}

// HKDF derives length bytes from secret using salt, info, and algo.
// Supported algorithms are "sha256" and "sha512"; unknown algorithms panic.
//
//	key := core.HKDF("sha256", secret, salt, []byte("session"), 32)
func HKDF(algo string, secret, salt, info []byte, length int) []byte {
	key, err := hkdf.Key(hashFor(algo), secret, salt, string(info), length)
	if err != nil {
		panic(err)
	}
	return key
}

func hashFor(algo string) func() hash.Hash {
	switch algo {
	case "sha256":
		return sha256.New
	case "sha512":
		return sha512.New
	default:
		panic(Concat("unknown hash algorithm: ", algo))
	}
}
