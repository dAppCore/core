// This file implements the Pre-Obfuscation Layer Protocol with
// XChaCha20-Poly1305 encryption. The protocol applies a reversible transformation
// to plaintext BEFORE it reaches CPU encryption routines, providing defense-in-depth
// against side-channel attacks.
//
// The encryption flow is:
//
//	plaintext -> obfuscate(nonce) -> encrypt -> [nonce || ciphertext || tag]
//
// The decryption flow is:
//
//	[nonce || ciphertext || tag] -> decrypt -> deobfuscate(nonce) -> plaintext
package sigil

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"io"

	"golang.org/x/crypto/chacha20poly1305"
)

var (
	// ErrInvalidKey is returned when the encryption key is invalid.
	ErrInvalidKey = errors.New("sigil: invalid key size, must be 32 bytes")
	// ErrCiphertextTooShort is returned when the ciphertext is too short to decrypt.
	ErrCiphertextTooShort = errors.New("sigil: ciphertext too short")
	// ErrDecryptionFailed is returned when decryption or authentication fails.
	ErrDecryptionFailed = errors.New("sigil: decryption failed")
	// ErrNoKeyConfigured is returned when no encryption key has been set.
	ErrNoKeyConfigured = errors.New("sigil: no encryption key configured")
)

// PreObfuscator applies a reversible transformation to data before encryption.
// This ensures that raw plaintext patterns are never sent directly to CPU
// encryption routines, providing defense against side-channel attacks.
//
// Implementations must be deterministic: given the same entropy, the transformation
// must be perfectly reversible: Deobfuscate(Obfuscate(x, e), e) == x
type PreObfuscator interface {
	// Obfuscate transforms plaintext before encryption using the provided entropy.
	// The entropy is typically the encryption nonce, ensuring the transformation
	// is unique per-encryption without additional random generation.
	Obfuscate(data []byte, entropy []byte) []byte

	// Deobfuscate reverses the transformation after decryption.
	// Must be called with the same entropy used during Obfuscate.
	Deobfuscate(data []byte, entropy []byte) []byte
}

// XORObfuscator performs XOR-based obfuscation using an entropy-derived key stream.
//
// The key stream is generated using SHA-256 in counter mode:
//
//	keyStream[i*32:(i+1)*32] = SHA256(entropy || BigEndian64(i))
//
// This provides a cryptographically uniform key stream that decorrelates
// plaintext patterns from the data seen by the encryption routine.
// XOR is symmetric, so obfuscation and deobfuscation use the same operation.
type XORObfuscator struct{}

// Obfuscate XORs the data with a key stream derived from the entropy.
func (x *XORObfuscator) Obfuscate(data []byte, entropy []byte) []byte {
	if len(data) == 0 {
		return data
	}
	return x.transform(data, entropy)
}

// Deobfuscate reverses the XOR transformation (XOR is symmetric).
func (x *XORObfuscator) Deobfuscate(data []byte, entropy []byte) []byte {
	if len(data) == 0 {
		return data
	}
	return x.transform(data, entropy)
}

// transform applies XOR with an entropy-derived key stream.
func (x *XORObfuscator) transform(data []byte, entropy []byte) []byte {
	result := make([]byte, len(data))
	keyStream := x.deriveKeyStream(entropy, len(data))
	for i := range data {
		result[i] = data[i] ^ keyStream[i]
	}
	return result
}

// deriveKeyStream creates a deterministic key stream from entropy.
func (x *XORObfuscator) deriveKeyStream(entropy []byte, length int) []byte {
	stream := make([]byte, length)
	h := sha256.New()

	// Generate key stream in 32-byte blocks
	blockNum := uint64(0)
	offset := 0
	for offset < length {
		h.Reset()
		h.Write(entropy)
		var blockBytes [8]byte
		binary.BigEndian.PutUint64(blockBytes[:], blockNum)
		h.Write(blockBytes[:])
		block := h.Sum(nil)

		copyLen := min(len(block), length-offset)
		copy(stream[offset:], block[:copyLen])
		offset += copyLen
		blockNum++
	}
	return stream
}

// ShuffleMaskObfuscator provides stronger obfuscation through byte shuffling and masking.
//
// The obfuscation process:
//  1. Generate a mask from entropy using SHA-256 in counter mode
//  2. XOR the data with the mask
//  3. Generate a deterministic permutation using Fisher-Yates shuffle
//  4. Reorder bytes according to the permutation
//
// This provides both value transformation (XOR mask) and position transformation
// (shuffle), making pattern analysis more difficult than XOR alone.
type ShuffleMaskObfuscator struct{}

// Obfuscate shuffles bytes and applies a mask derived from entropy.
func (s *ShuffleMaskObfuscator) Obfuscate(data []byte, entropy []byte) []byte {
	if len(data) == 0 {
		return data
	}

	result := make([]byte, len(data))
	copy(result, data)

	// Generate permutation and mask from entropy
	perm := s.generatePermutation(entropy, len(data))
	mask := s.deriveMask(entropy, len(data))

	// Apply mask first, then shuffle
	for i := range result {
		result[i] ^= mask[i]
	}

	// Shuffle using Fisher-Yates with deterministic seed
	shuffled := make([]byte, len(data))
	for i, p := range perm {
		shuffled[i] = result[p]
	}

	return shuffled
}

// Deobfuscate reverses the shuffle and mask operations.
func (s *ShuffleMaskObfuscator) Deobfuscate(data []byte, entropy []byte) []byte {
	if len(data) == 0 {
		return data
	}

	result := make([]byte, len(data))

	// Generate permutation and mask from entropy
	perm := s.generatePermutation(entropy, len(data))
	mask := s.deriveMask(entropy, len(data))

	// Unshuffle first
	for i, p := range perm {
		result[p] = data[i]
	}

	// Remove mask
	for i := range result {
		result[i] ^= mask[i]
	}

	return result
}

// generatePermutation creates a deterministic permutation from entropy.
func (s *ShuffleMaskObfuscator) generatePermutation(entropy []byte, length int) []int {
	perm := make([]int, length)
	for i := range perm {
		perm[i] = i
	}

	// Use entropy to seed a deterministic shuffle
	h := sha256.New()
	h.Write(entropy)
	h.Write([]byte("permutation"))
	seed := h.Sum(nil)

	// Fisher-Yates shuffle with deterministic randomness
	for i := length - 1; i > 0; i-- {
		h.Reset()
		h.Write(seed)
		var iBytes [8]byte
		binary.BigEndian.PutUint64(iBytes[:], uint64(i))
		h.Write(iBytes[:])
		jBytes := h.Sum(nil)
		j := int(binary.BigEndian.Uint64(jBytes[:8]) % uint64(i+1))
		perm[i], perm[j] = perm[j], perm[i]
	}

	return perm
}

// deriveMask creates a mask byte array from entropy.
func (s *ShuffleMaskObfuscator) deriveMask(entropy []byte, length int) []byte {
	mask := make([]byte, length)
	h := sha256.New()

	blockNum := uint64(0)
	offset := 0
	for offset < length {
		h.Reset()
		h.Write(entropy)
		h.Write([]byte("mask"))
		var blockBytes [8]byte
		binary.BigEndian.PutUint64(blockBytes[:], blockNum)
		h.Write(blockBytes[:])
		block := h.Sum(nil)

		copyLen := min(len(block), length-offset)
		copy(mask[offset:], block[:copyLen])
		offset += copyLen
		blockNum++
	}
	return mask
}

// ChaChaPolySigil is a Sigil that encrypts/decrypts data using ChaCha20-Poly1305.
// It applies pre-obfuscation before encryption to ensure raw plaintext never
// goes directly to CPU encryption routines.
//
// The output format is:
// [24-byte nonce][encrypted(obfuscated(plaintext))]
//
// Unlike demo implementations, the nonce is ONLY embedded in the ciphertext,
// not exposed separately in headers.
type ChaChaPolySigil struct {
	Key        []byte
	Obfuscator PreObfuscator
	randReader io.Reader // for testing injection
}

// NewChaChaPolySigil creates a new encryption sigil with the given key.
// The key must be exactly 32 bytes.
func NewChaChaPolySigil(key []byte) (*ChaChaPolySigil, error) {
	if len(key) != 32 {
		return nil, ErrInvalidKey
	}

	keyCopy := make([]byte, 32)
	copy(keyCopy, key)

	return &ChaChaPolySigil{
		Key:        keyCopy,
		Obfuscator: &XORObfuscator{},
		randReader: rand.Reader,
	}, nil
}

// NewChaChaPolySigilWithObfuscator creates a new encryption sigil with custom obfuscator.
func NewChaChaPolySigilWithObfuscator(key []byte, obfuscator PreObfuscator) (*ChaChaPolySigil, error) {
	sigil, err := NewChaChaPolySigil(key)
	if err != nil {
		return nil, err
	}
	if obfuscator != nil {
		sigil.Obfuscator = obfuscator
	}
	return sigil, nil
}

// In encrypts the data with pre-obfuscation.
// The flow is: plaintext -> obfuscate -> encrypt
func (s *ChaChaPolySigil) In(data []byte) ([]byte, error) {
	if s.Key == nil {
		return nil, ErrNoKeyConfigured
	}
	if data == nil {
		return nil, nil
	}

	aead, err := chacha20poly1305.NewX(s.Key)
	if err != nil {
		return nil, err
	}

	// Generate nonce
	nonce := make([]byte, aead.NonceSize())
	reader := s.randReader
	if reader == nil {
		reader = rand.Reader
	}
	if _, err := io.ReadFull(reader, nonce); err != nil {
		return nil, err
	}

	// Pre-obfuscate the plaintext using nonce as entropy
	// This ensures CPU encryption routines never see raw plaintext
	obfuscated := data
	if s.Obfuscator != nil {
		obfuscated = s.Obfuscator.Obfuscate(data, nonce)
	}

	// Encrypt the obfuscated data
	// Output: [nonce | ciphertext | auth tag]
	ciphertext := aead.Seal(nonce, nonce, obfuscated, nil)

	return ciphertext, nil
}

// Out decrypts the data and reverses obfuscation.
// The flow is: decrypt -> deobfuscate -> plaintext
func (s *ChaChaPolySigil) Out(data []byte) ([]byte, error) {
	if s.Key == nil {
		return nil, ErrNoKeyConfigured
	}
	if data == nil {
		return nil, nil
	}

	aead, err := chacha20poly1305.NewX(s.Key)
	if err != nil {
		return nil, err
	}

	minLen := aead.NonceSize() + aead.Overhead()
	if len(data) < minLen {
		return nil, ErrCiphertextTooShort
	}

	// Extract nonce from ciphertext
	nonce := data[:aead.NonceSize()]
	ciphertext := data[aead.NonceSize():]

	// Decrypt
	obfuscated, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	// Deobfuscate using the same nonce as entropy
	plaintext := obfuscated
	if s.Obfuscator != nil {
		plaintext = s.Obfuscator.Deobfuscate(obfuscated, nonce)
	}

	if len(plaintext) == 0 {
		return []byte{}, nil
	}

	return plaintext, nil
}

// GetNonceFromCiphertext extracts the nonce from encrypted output.
// This is provided for debugging/logging purposes only.
// The nonce should NOT be stored separately in headers.
func GetNonceFromCiphertext(ciphertext []byte) ([]byte, error) {
	nonceSize := chacha20poly1305.NonceSizeX
	if len(ciphertext) < nonceSize {
		return nil, ErrCiphertextTooShort
	}
	nonceCopy := make([]byte, nonceSize)
	copy(nonceCopy, ciphertext[:nonceSize])
	return nonceCopy, nil
}
