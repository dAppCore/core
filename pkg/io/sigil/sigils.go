package sigil

import (
	"bytes"
	"compress/gzip"
	"crypto"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/blake2s"
	"golang.org/x/crypto/md4"
	"golang.org/x/crypto/ripemd160"
	"golang.org/x/crypto/sha3"
)

// ReverseSigil is a Sigil that reverses the bytes of the payload.
// It is a symmetrical Sigil, meaning that the In and Out methods perform the same operation.
type ReverseSigil struct{}

// In reverses the bytes of the data.
func (s *ReverseSigil) In(data []byte) ([]byte, error) {
	if data == nil {
		return nil, nil
	}
	reversed := make([]byte, len(data))
	for i, j := 0, len(data)-1; i < len(data); i, j = i+1, j-1 {
		reversed[i] = data[j]
	}
	return reversed, nil
}

// Out reverses the bytes of the data.
func (s *ReverseSigil) Out(data []byte) ([]byte, error) {
	return s.In(data)
}

// HexSigil is a Sigil that encodes/decodes data to/from hexadecimal.
// The In method encodes the data, and the Out method decodes it.
type HexSigil struct{}

// In encodes the data to hexadecimal.
func (s *HexSigil) In(data []byte) ([]byte, error) {
	if data == nil {
		return nil, nil
	}
	dst := make([]byte, hex.EncodedLen(len(data)))
	hex.Encode(dst, data)
	return dst, nil
}

// Out decodes the data from hexadecimal.
func (s *HexSigil) Out(data []byte) ([]byte, error) {
	if data == nil {
		return nil, nil
	}
	dst := make([]byte, hex.DecodedLen(len(data)))
	_, err := hex.Decode(dst, data)
	return dst, err
}

// Base64Sigil is a Sigil that encodes/decodes data to/from base64.
// The In method encodes the data, and the Out method decodes it.
type Base64Sigil struct{}

// In encodes the data to base64.
func (s *Base64Sigil) In(data []byte) ([]byte, error) {
	if data == nil {
		return nil, nil
	}
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(dst, data)
	return dst, nil
}

// Out decodes the data from base64.
func (s *Base64Sigil) Out(data []byte) ([]byte, error) {
	if data == nil {
		return nil, nil
	}
	dst := make([]byte, base64.StdEncoding.DecodedLen(len(data)))
	n, err := base64.StdEncoding.Decode(dst, data)
	return dst[:n], err
}

// GzipSigil is a Sigil that compresses/decompresses data using gzip.
// The In method compresses the data, and the Out method decompresses it.
type GzipSigil struct {
	writer io.Writer
}

// In compresses the data using gzip.
func (s *GzipSigil) In(data []byte) ([]byte, error) {
	if data == nil {
		return nil, nil
	}
	var b bytes.Buffer
	w := s.writer
	if w == nil {
		w = &b
	}
	gz := gzip.NewWriter(w)
	if _, err := gz.Write(data); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// Out decompresses the data using gzip.
func (s *GzipSigil) Out(data []byte) ([]byte, error) {
	if data == nil {
		return nil, nil
	}
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}

// JSONSigil is a Sigil that compacts or indents JSON data.
// The Out method is a no-op.
type JSONSigil struct{ Indent bool }

// In compacts or indents the JSON data.
func (s *JSONSigil) In(data []byte) ([]byte, error) {
	if s.Indent {
		var out bytes.Buffer
		err := json.Indent(&out, data, "", "  ")
		return out.Bytes(), err
	}
	var out bytes.Buffer
	err := json.Compact(&out, data)
	return out.Bytes(), err
}

// Out is a no-op for JSONSigil.
func (s *JSONSigil) Out(data []byte) ([]byte, error) {
	// For simplicity, Out is a no-op. The primary use is formatting.
	return data, nil
}

// HashSigil is a Sigil that hashes the data using a specified algorithm.
// The In method hashes the data, and the Out method is a no-op.
type HashSigil struct {
	Hash crypto.Hash
}

// NewHashSigil creates a new HashSigil.
func NewHashSigil(h crypto.Hash) *HashSigil {
	return &HashSigil{Hash: h}
}

// In hashes the data.
func (s *HashSigil) In(data []byte) ([]byte, error) {
	var h io.Writer
	switch s.Hash {
	case crypto.MD4:
		h = md4.New()
	case crypto.MD5:
		h = md5.New()
	case crypto.SHA1:
		h = sha1.New()
	case crypto.SHA224:
		h = sha256.New224()
	case crypto.SHA256:
		h = sha256.New()
	case crypto.SHA384:
		h = sha512.New384()
	case crypto.SHA512:
		h = sha512.New()
	case crypto.RIPEMD160:
		h = ripemd160.New()
	case crypto.SHA3_224:
		h = sha3.New224()
	case crypto.SHA3_256:
		h = sha3.New256()
	case crypto.SHA3_384:
		h = sha3.New384()
	case crypto.SHA3_512:
		h = sha3.New512()
	case crypto.SHA512_224:
		h = sha512.New512_224()
	case crypto.SHA512_256:
		h = sha512.New512_256()
	case crypto.BLAKE2s_256:
		h, _ = blake2s.New256(nil)
	case crypto.BLAKE2b_256:
		h, _ = blake2b.New256(nil)
	case crypto.BLAKE2b_384:
		h, _ = blake2b.New384(nil)
	case crypto.BLAKE2b_512:
		h, _ = blake2b.New512(nil)
	default:
		// MD5SHA1 is not supported as a direct hash
		return nil, errors.New("sigil: hash algorithm not available")
	}

	h.Write(data)
	return h.(interface{ Sum([]byte) []byte }).Sum(nil), nil
}

// Out is a no-op for HashSigil.
func (s *HashSigil) Out(data []byte) ([]byte, error) {
	return data, nil
}

// NewSigil is a factory function that returns a Sigil based on a string name.
// It is the primary way to create Sigil instances.
func NewSigil(name string) (Sigil, error) {
	switch name {
	case "reverse":
		return &ReverseSigil{}, nil
	case "hex":
		return &HexSigil{}, nil
	case "base64":
		return &Base64Sigil{}, nil
	case "gzip":
		return &GzipSigil{}, nil
	case "json":
		return &JSONSigil{Indent: false}, nil
	case "json-indent":
		return &JSONSigil{Indent: true}, nil
	case "md4":
		return NewHashSigil(crypto.MD4), nil
	case "md5":
		return NewHashSigil(crypto.MD5), nil
	case "sha1":
		return NewHashSigil(crypto.SHA1), nil
	case "sha224":
		return NewHashSigil(crypto.SHA224), nil
	case "sha256":
		return NewHashSigil(crypto.SHA256), nil
	case "sha384":
		return NewHashSigil(crypto.SHA384), nil
	case "sha512":
		return NewHashSigil(crypto.SHA512), nil
	case "ripemd160":
		return NewHashSigil(crypto.RIPEMD160), nil
	case "sha3-224":
		return NewHashSigil(crypto.SHA3_224), nil
	case "sha3-256":
		return NewHashSigil(crypto.SHA3_256), nil
	case "sha3-384":
		return NewHashSigil(crypto.SHA3_384), nil
	case "sha3-512":
		return NewHashSigil(crypto.SHA3_512), nil
	case "sha512-224":
		return NewHashSigil(crypto.SHA512_224), nil
	case "sha512-256":
		return NewHashSigil(crypto.SHA512_256), nil
	case "blake2s-256":
		return NewHashSigil(crypto.BLAKE2s_256), nil
	case "blake2b-256":
		return NewHashSigil(crypto.BLAKE2b_256), nil
	case "blake2b-384":
		return NewHashSigil(crypto.BLAKE2b_384), nil
	case "blake2b-512":
		return NewHashSigil(crypto.BLAKE2b_512), nil
	default:
		return nil, errors.New("sigil: unknown sigil name")
	}
}
