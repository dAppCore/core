package sigil

import (
	"bytes"
	"crypto/rand"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── XORObfuscator ──────────────────────────────────────────────────

func TestXORObfuscator_Good_RoundTrip(t *testing.T) {
	ob := &XORObfuscator{}
	data := []byte("the axioms are in the weights")
	entropy := []byte("deterministic-nonce-24bytes!")

	obfuscated := ob.Obfuscate(data, entropy)
	assert.NotEqual(t, data, obfuscated)
	assert.Len(t, obfuscated, len(data))

	restored := ob.Deobfuscate(obfuscated, entropy)
	assert.Equal(t, data, restored)
}

func TestXORObfuscator_Good_DifferentEntropyDifferentOutput(t *testing.T) {
	ob := &XORObfuscator{}
	data := []byte("same plaintext")

	out1 := ob.Obfuscate(data, []byte("entropy-a"))
	out2 := ob.Obfuscate(data, []byte("entropy-b"))
	assert.NotEqual(t, out1, out2)
}

func TestXORObfuscator_Good_Deterministic(t *testing.T) {
	ob := &XORObfuscator{}
	data := []byte("reproducible")
	entropy := []byte("fixed-seed")

	out1 := ob.Obfuscate(data, entropy)
	out2 := ob.Obfuscate(data, entropy)
	assert.Equal(t, out1, out2)
}

func TestXORObfuscator_Good_LargeData(t *testing.T) {
	ob := &XORObfuscator{}
	// Larger than one SHA-256 block (32 bytes) to test multi-block key stream.
	data := make([]byte, 256)
	for i := range data {
		data[i] = byte(i)
	}
	entropy := []byte("test-entropy")

	obfuscated := ob.Obfuscate(data, entropy)
	restored := ob.Deobfuscate(obfuscated, entropy)
	assert.Equal(t, data, restored)
}

func TestXORObfuscator_Good_EmptyData(t *testing.T) {
	ob := &XORObfuscator{}
	result := ob.Obfuscate([]byte{}, []byte("entropy"))
	assert.Equal(t, []byte{}, result)

	result = ob.Deobfuscate([]byte{}, []byte("entropy"))
	assert.Equal(t, []byte{}, result)
}

func TestXORObfuscator_Good_SymmetricProperty(t *testing.T) {
	ob := &XORObfuscator{}
	data := []byte("XOR is its own inverse")
	entropy := []byte("nonce")

	// XOR is symmetric: Obfuscate(Obfuscate(x)) == x
	double := ob.Obfuscate(ob.Obfuscate(data, entropy), entropy)
	assert.Equal(t, data, double)
}

// ── ShuffleMaskObfuscator ──────────────────────────────────────────

func TestShuffleMaskObfuscator_Good_RoundTrip(t *testing.T) {
	ob := &ShuffleMaskObfuscator{}
	data := []byte("shuffle and mask protect patterns")
	entropy := []byte("deterministic-entropy")

	obfuscated := ob.Obfuscate(data, entropy)
	assert.NotEqual(t, data, obfuscated)
	assert.Len(t, obfuscated, len(data))

	restored := ob.Deobfuscate(obfuscated, entropy)
	assert.Equal(t, data, restored)
}

func TestShuffleMaskObfuscator_Good_DifferentEntropy(t *testing.T) {
	ob := &ShuffleMaskObfuscator{}
	data := []byte("same data")

	out1 := ob.Obfuscate(data, []byte("entropy-1"))
	out2 := ob.Obfuscate(data, []byte("entropy-2"))
	assert.NotEqual(t, out1, out2)
}

func TestShuffleMaskObfuscator_Good_Deterministic(t *testing.T) {
	ob := &ShuffleMaskObfuscator{}
	data := []byte("reproducible shuffle")
	entropy := []byte("fixed")

	out1 := ob.Obfuscate(data, entropy)
	out2 := ob.Obfuscate(data, entropy)
	assert.Equal(t, out1, out2)
}

func TestShuffleMaskObfuscator_Good_LargeData(t *testing.T) {
	ob := &ShuffleMaskObfuscator{}
	data := make([]byte, 512)
	for i := range data {
		data[i] = byte(i % 256)
	}
	entropy := []byte("large-data-test")

	obfuscated := ob.Obfuscate(data, entropy)
	restored := ob.Deobfuscate(obfuscated, entropy)
	assert.Equal(t, data, restored)
}

func TestShuffleMaskObfuscator_Good_EmptyData(t *testing.T) {
	ob := &ShuffleMaskObfuscator{}
	result := ob.Obfuscate([]byte{}, []byte("entropy"))
	assert.Equal(t, []byte{}, result)

	result = ob.Deobfuscate([]byte{}, []byte("entropy"))
	assert.Equal(t, []byte{}, result)
}

func TestShuffleMaskObfuscator_Good_SingleByte(t *testing.T) {
	ob := &ShuffleMaskObfuscator{}
	data := []byte{0x42}
	entropy := []byte("single")

	obfuscated := ob.Obfuscate(data, entropy)
	restored := ob.Deobfuscate(obfuscated, entropy)
	assert.Equal(t, data, restored)
}

// ── NewChaChaPolySigil ─────────────────────────────────────────────

func TestNewChaChaPolySigil_Good(t *testing.T) {
	key := make([]byte, 32)
	_, _ = rand.Read(key)

	s, err := NewChaChaPolySigil(key)
	require.NoError(t, err)
	assert.NotNil(t, s)
	assert.Equal(t, key, s.Key)
	assert.NotNil(t, s.Obfuscator)
}

func TestNewChaChaPolySigil_Good_KeyIsCopied(t *testing.T) {
	key := make([]byte, 32)
	_, _ = rand.Read(key)
	original := make([]byte, 32)
	copy(original, key)

	s, err := NewChaChaPolySigil(key)
	require.NoError(t, err)

	// Mutating the original key should not affect the sigil.
	key[0] ^= 0xFF
	assert.Equal(t, original, s.Key)
}

func TestNewChaChaPolySigil_Bad_ShortKey(t *testing.T) {
	_, err := NewChaChaPolySigil([]byte("too short"))
	assert.ErrorIs(t, err, ErrInvalidKey)
}

func TestNewChaChaPolySigil_Bad_LongKey(t *testing.T) {
	_, err := NewChaChaPolySigil(make([]byte, 64))
	assert.ErrorIs(t, err, ErrInvalidKey)
}

func TestNewChaChaPolySigil_Bad_EmptyKey(t *testing.T) {
	_, err := NewChaChaPolySigil(nil)
	assert.ErrorIs(t, err, ErrInvalidKey)
}

// ── NewChaChaPolySigilWithObfuscator ───────────────────────────────

func TestNewChaChaPolySigilWithObfuscator_Good(t *testing.T) {
	key := make([]byte, 32)
	_, _ = rand.Read(key)

	ob := &ShuffleMaskObfuscator{}
	s, err := NewChaChaPolySigilWithObfuscator(key, ob)
	require.NoError(t, err)
	assert.Equal(t, ob, s.Obfuscator)
}

func TestNewChaChaPolySigilWithObfuscator_Good_NilObfuscator(t *testing.T) {
	key := make([]byte, 32)
	_, _ = rand.Read(key)

	s, err := NewChaChaPolySigilWithObfuscator(key, nil)
	require.NoError(t, err)
	// Falls back to default XORObfuscator.
	assert.IsType(t, &XORObfuscator{}, s.Obfuscator)
}

func TestNewChaChaPolySigilWithObfuscator_Bad_InvalidKey(t *testing.T) {
	_, err := NewChaChaPolySigilWithObfuscator([]byte("bad"), &XORObfuscator{})
	assert.ErrorIs(t, err, ErrInvalidKey)
}

// ── ChaChaPolySigil In/Out (encrypt/decrypt) ───────────────────────

func TestChaChaPolySigil_Good_RoundTrip(t *testing.T) {
	key := make([]byte, 32)
	_, _ = rand.Read(key)

	s, err := NewChaChaPolySigil(key)
	require.NoError(t, err)

	plaintext := []byte("consciousness does not merely avoid causing harm")
	ciphertext, err := s.In(plaintext)
	require.NoError(t, err)
	assert.NotEqual(t, plaintext, ciphertext)
	assert.Greater(t, len(ciphertext), len(plaintext)) // nonce + tag overhead

	decrypted, err := s.Out(ciphertext)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestChaChaPolySigil_Good_WithShuffleMask(t *testing.T) {
	key := make([]byte, 32)
	_, _ = rand.Read(key)

	s, err := NewChaChaPolySigilWithObfuscator(key, &ShuffleMaskObfuscator{})
	require.NoError(t, err)

	plaintext := []byte("shuffle mask pre-obfuscation layer")
	ciphertext, err := s.In(plaintext)
	require.NoError(t, err)

	decrypted, err := s.Out(ciphertext)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestChaChaPolySigil_Good_NilData(t *testing.T) {
	key := make([]byte, 32)
	_, _ = rand.Read(key)

	s, err := NewChaChaPolySigil(key)
	require.NoError(t, err)

	enc, err := s.In(nil)
	require.NoError(t, err)
	assert.Nil(t, enc)

	dec, err := s.Out(nil)
	require.NoError(t, err)
	assert.Nil(t, dec)
}

func TestChaChaPolySigil_Good_EmptyPlaintext(t *testing.T) {
	key := make([]byte, 32)
	_, _ = rand.Read(key)

	s, err := NewChaChaPolySigil(key)
	require.NoError(t, err)

	ciphertext, err := s.In([]byte{})
	require.NoError(t, err)
	assert.NotEmpty(t, ciphertext) // Has nonce + tag even for empty plaintext.

	decrypted, err := s.Out(ciphertext)
	require.NoError(t, err)
	assert.Equal(t, []byte{}, decrypted)
}

func TestChaChaPolySigil_Good_DifferentCiphertextsPerCall(t *testing.T) {
	key := make([]byte, 32)
	_, _ = rand.Read(key)

	s, err := NewChaChaPolySigil(key)
	require.NoError(t, err)

	plaintext := []byte("same input")
	ct1, _ := s.In(plaintext)
	ct2, _ := s.In(plaintext)

	// Different nonces → different ciphertexts.
	assert.NotEqual(t, ct1, ct2)
}

func TestChaChaPolySigil_Bad_NoKey(t *testing.T) {
	s := &ChaChaPolySigil{}

	_, err := s.In([]byte("data"))
	assert.ErrorIs(t, err, ErrNoKeyConfigured)

	_, err = s.Out([]byte("data"))
	assert.ErrorIs(t, err, ErrNoKeyConfigured)
}

func TestChaChaPolySigil_Bad_WrongKey(t *testing.T) {
	key1 := make([]byte, 32)
	key2 := make([]byte, 32)
	_, _ = rand.Read(key1)
	_, _ = rand.Read(key2)

	s1, _ := NewChaChaPolySigil(key1)
	s2, _ := NewChaChaPolySigil(key2)

	ciphertext, err := s1.In([]byte("secret"))
	require.NoError(t, err)

	_, err = s2.Out(ciphertext)
	assert.ErrorIs(t, err, ErrDecryptionFailed)
}

func TestChaChaPolySigil_Bad_TruncatedCiphertext(t *testing.T) {
	key := make([]byte, 32)
	_, _ = rand.Read(key)

	s, _ := NewChaChaPolySigil(key)
	_, err := s.Out([]byte("too short"))
	assert.ErrorIs(t, err, ErrCiphertextTooShort)
}

func TestChaChaPolySigil_Bad_TamperedCiphertext(t *testing.T) {
	key := make([]byte, 32)
	_, _ = rand.Read(key)

	s, _ := NewChaChaPolySigil(key)
	ciphertext, _ := s.In([]byte("authentic data"))

	// Flip a bit in the ciphertext body (after nonce).
	ciphertext[30] ^= 0xFF

	_, err := s.Out(ciphertext)
	assert.ErrorIs(t, err, ErrDecryptionFailed)
}

// failReader returns an error on read — for testing nonce generation failure.
type failReader struct{}

func (f *failReader) Read([]byte) (int, error) {
	return 0, errors.New("entropy source failed")
}

func TestChaChaPolySigil_Bad_RandReaderFailure(t *testing.T) {
	key := make([]byte, 32)
	_, _ = rand.Read(key)

	s, _ := NewChaChaPolySigil(key)
	s.randReader = &failReader{}

	_, err := s.In([]byte("data"))
	assert.Error(t, err)
}

// ── ChaChaPolySigil without obfuscator ─────────────────────────────

func TestChaChaPolySigil_Good_NoObfuscator(t *testing.T) {
	key := make([]byte, 32)
	_, _ = rand.Read(key)

	s, _ := NewChaChaPolySigil(key)
	s.Obfuscator = nil // Disable pre-obfuscation.

	plaintext := []byte("raw encryption without pre-obfuscation")
	ciphertext, err := s.In(plaintext)
	require.NoError(t, err)

	decrypted, err := s.Out(ciphertext)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

// ── GetNonceFromCiphertext ─────────────────────────────────────────

func TestGetNonceFromCiphertext_Good(t *testing.T) {
	key := make([]byte, 32)
	_, _ = rand.Read(key)

	s, _ := NewChaChaPolySigil(key)
	ciphertext, _ := s.In([]byte("nonce extraction test"))

	nonce, err := GetNonceFromCiphertext(ciphertext)
	require.NoError(t, err)
	assert.Len(t, nonce, 24) // XChaCha20 nonce is 24 bytes.

	// Nonce should match the prefix of the ciphertext.
	assert.Equal(t, ciphertext[:24], nonce)
}

func TestGetNonceFromCiphertext_Good_NonceCopied(t *testing.T) {
	key := make([]byte, 32)
	_, _ = rand.Read(key)

	s, _ := NewChaChaPolySigil(key)
	ciphertext, _ := s.In([]byte("data"))

	nonce, _ := GetNonceFromCiphertext(ciphertext)
	original := make([]byte, len(nonce))
	copy(original, nonce)

	// Mutating the nonce should not affect the ciphertext.
	nonce[0] ^= 0xFF
	assert.Equal(t, original, ciphertext[:24])
}

func TestGetNonceFromCiphertext_Bad_TooShort(t *testing.T) {
	_, err := GetNonceFromCiphertext([]byte("short"))
	assert.ErrorIs(t, err, ErrCiphertextTooShort)
}

func TestGetNonceFromCiphertext_Bad_Empty(t *testing.T) {
	_, err := GetNonceFromCiphertext(nil)
	assert.ErrorIs(t, err, ErrCiphertextTooShort)
}

// ── ChaChaPolySigil in Transmute pipeline ──────────────────────────

func TestChaChaPolySigil_Good_InTransmutePipeline(t *testing.T) {
	key := make([]byte, 32)
	_, _ = rand.Read(key)

	s, _ := NewChaChaPolySigil(key)
	hexSigil, _ := NewSigil("hex")

	chain := []Sigil{s, hexSigil}
	plaintext := []byte("encrypt then hex encode")

	encoded, err := Transmute(plaintext, chain)
	require.NoError(t, err)

	// Result should be hex-encoded ciphertext.
	assert.True(t, isHex(encoded))

	decoded, err := Untransmute(encoded, chain)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decoded)
}

func isHex(data []byte) bool {
	for _, b := range data {
		if !((b >= '0' && b <= '9') || (b >= 'a' && b <= 'f')) {
			return false
		}
	}
	return len(data) > 0
}

// ── Transmute error propagation ────────────────────────────────────

type failSigil struct{}

func (f *failSigil) In([]byte) ([]byte, error)  { return nil, errors.New("fail in") }
func (f *failSigil) Out([]byte) ([]byte, error) { return nil, errors.New("fail out") }

func TestTransmute_Bad_ErrorPropagation(t *testing.T) {
	_, err := Transmute([]byte("data"), []Sigil{&failSigil{}})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fail in")
}

func TestUntransmute_Bad_ErrorPropagation(t *testing.T) {
	_, err := Untransmute([]byte("data"), []Sigil{&failSigil{}})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fail out")
}

// ── GzipSigil with custom writer (edge case) ──────────────────────

func TestGzipSigil_Good_CustomWriter(t *testing.T) {
	var buf bytes.Buffer
	s := &GzipSigil{writer: &buf}

	// With custom writer, compressed data goes to buf, returned bytes will be empty
	// because the internal buffer 'b' is unused when s.writer is set.
	_, err := s.In([]byte("test data"))
	require.NoError(t, err)
	assert.Greater(t, buf.Len(), 0)
}

// ── deriveKeyStream edge: exactly 32 bytes ─────────────────────────

func TestDeriveKeyStream_Good_ExactBlockSize(t *testing.T) {
	ob := &XORObfuscator{}
	data := make([]byte, 32) // Exactly one SHA-256 block.
	for i := range data {
		data[i] = byte(i)
	}
	entropy := []byte("block-boundary")

	obfuscated := ob.Obfuscate(data, entropy)
	restored := ob.Deobfuscate(obfuscated, entropy)
	assert.Equal(t, data, restored)
}

// ── io.Reader fallback in In ───────────────────────────────────────

func TestChaChaPolySigil_Good_NilRandReader(t *testing.T) {
	key := make([]byte, 32)
	_, _ = rand.Read(key)

	s, _ := NewChaChaPolySigil(key)
	s.randReader = nil // Should fall back to crypto/rand.Reader.

	ciphertext, err := s.In([]byte("fallback reader"))
	require.NoError(t, err)

	decrypted, err := s.Out(ciphertext)
	require.NoError(t, err)
	assert.Equal(t, []byte("fallback reader"), decrypted)
}

// limitReader returns exactly N bytes then EOF — for deterministic tests.
type limitReader struct {
	data []byte
	pos  int
}

func (l *limitReader) Read(p []byte) (int, error) {
	if l.pos >= len(l.data) {
		return 0, io.EOF
	}
	n := copy(p, l.data[l.pos:])
	l.pos += n
	return n, nil
}
