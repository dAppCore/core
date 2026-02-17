package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetGet_Good(t *testing.T) {
	s, err := New(":memory:")
	require.NoError(t, err)
	defer s.Close()

	err = s.Set("config", "theme", "dark")
	require.NoError(t, err)

	val, err := s.Get("config", "theme")
	require.NoError(t, err)
	assert.Equal(t, "dark", val)
}

func TestGet_Bad_NotFound(t *testing.T) {
	s, _ := New(":memory:")
	defer s.Close()

	_, err := s.Get("config", "missing")
	assert.Error(t, err)
}

func TestDelete_Good(t *testing.T) {
	s, _ := New(":memory:")
	defer s.Close()

	_ = s.Set("config", "key", "val")
	err := s.Delete("config", "key")
	require.NoError(t, err)

	_, err = s.Get("config", "key")
	assert.Error(t, err)
}

func TestCount_Good(t *testing.T) {
	s, _ := New(":memory:")
	defer s.Close()

	_ = s.Set("grp", "a", "1")
	_ = s.Set("grp", "b", "2")
	_ = s.Set("other", "c", "3")

	n, err := s.Count("grp")
	require.NoError(t, err)
	assert.Equal(t, 2, n)
}

func TestDeleteGroup_Good(t *testing.T) {
	s, _ := New(":memory:")
	defer s.Close()

	_ = s.Set("grp", "a", "1")
	_ = s.Set("grp", "b", "2")
	err := s.DeleteGroup("grp")
	require.NoError(t, err)

	n, _ := s.Count("grp")
	assert.Equal(t, 0, n)
}

func TestRender_Good(t *testing.T) {
	s, _ := New(":memory:")
	defer s.Close()

	_ = s.Set("user", "pool", "pool.lthn.io:3333")
	_ = s.Set("user", "wallet", "iz...")

	tmpl := `{"pool":"{{ .pool }}","wallet":"{{ .wallet }}"}`
	out, err := s.Render(tmpl, "user")
	require.NoError(t, err)
	assert.Contains(t, out, "pool.lthn.io:3333")
	assert.Contains(t, out, "iz...")
}
