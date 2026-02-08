package rsa

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockReader is a reader that returns an error.
type mockReader struct{}

func (r *mockReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

func TestRSA_Good(t *testing.T) {
	s := NewService()

	// Generate a new key pair
	pubKey, privKey, err := s.GenerateKeyPair(2048)
	assert.NoError(t, err)
	assert.NotEmpty(t, pubKey)
	assert.NotEmpty(t, privKey)

	// Encrypt and decrypt a message
	message := []byte("Hello, World!")
	ciphertext, err := s.Encrypt(pubKey, message, nil)
	assert.NoError(t, err)
	plaintext, err := s.Decrypt(privKey, ciphertext, nil)
	assert.NoError(t, err)
	assert.Equal(t, message, plaintext)
}

func TestRSA_Bad(t *testing.T) {
	s := NewService()

	// Decrypt with wrong key
	pubKey, _, err := s.GenerateKeyPair(2048)
	assert.NoError(t, err)
	_, otherPrivKey, err := s.GenerateKeyPair(2048)
	assert.NoError(t, err)
	message := []byte("Hello, World!")
	ciphertext, err := s.Encrypt(pubKey, message, nil)
	assert.NoError(t, err)
	_, err = s.Decrypt(otherPrivKey, ciphertext, nil)
	assert.Error(t, err)

	// Key size too small
	_, _, err = s.GenerateKeyPair(512)
	assert.Error(t, err)
}

func TestRSA_Ugly(t *testing.T) {
	s := NewService()

	// Malformed keys and messages
	_, err := s.Encrypt([]byte("not-a-key"), []byte("message"), nil)
	assert.Error(t, err)
	_, err = s.Decrypt([]byte("not-a-key"), []byte("message"), nil)
	assert.Error(t, err)
	_, err = s.Encrypt([]byte("-----BEGIN PUBLIC KEY-----\nMFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBAJ/6j/y7/r/9/z/8/f/+/v7+/v7+/v7+\nv/7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v4=\n-----END PUBLIC KEY-----"), []byte("message"), nil)
	assert.Error(t, err)
	_, err = s.Decrypt([]byte("-----BEGIN RSA PRIVATE KEY-----\nMIIBOQIBAAJBAL/6j/y7/r/9/z/8/f/+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+\nv/7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v4CAwEAAQJB\nAL/6j/y7/r/9/z/8/f/+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+\nv/7+/v7+/v7+/v7+/v7+/v7+/v7+/v4CgYEA/f8/vLv+v/3/P/z9//7+/v7+/v7+\nvv7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v4C\ngYEA/f8/vLv+v/3/P/z9//7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+\nvv7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v4CgYEA/f8/vLv+v/3/P/z9//7+/v7+\nvv7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+\nv/4CgYEA/f8/vLv+v/3/P/z9//7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+\nvv7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v4CgYEA/f8/vLv+v/3/P/z9//7+/v7+\nvv7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+/v7+\nv/4=\n-----END RSA PRIVATE KEY-----"), []byte("message"), nil)
	assert.Error(t, err)

	// Key generation failure
	oldReader := rand.Reader
	rand.Reader = &mockReader{}
	t.Cleanup(func() { rand.Reader = oldReader })
	_, _, err = s.GenerateKeyPair(2048)
	assert.Error(t, err)

	// Encrypt with non-RSA key
	rand.Reader = oldReader // Restore reader for this test
	ecdsaPrivKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)
	ecdsaPubKeyBytes, err := x509.MarshalPKIXPublicKey(&ecdsaPrivKey.PublicKey)
	assert.NoError(t, err)
	ecdsaPubKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: ecdsaPubKeyBytes,
	})
	_, err = s.Encrypt(ecdsaPubKeyPEM, []byte("message"), nil)
	assert.Error(t, err)
	rand.Reader = &mockReader{} // Set it back for the next test

	// Encrypt message too long
	rand.Reader = oldReader // Restore reader for this test
	pubKey, _, err := s.GenerateKeyPair(2048)
	assert.NoError(t, err)
	message := make([]byte, 2048)
	_, err = s.Encrypt(pubKey, message, nil)
	assert.Error(t, err)
	rand.Reader = &mockReader{} // Set it back
}
