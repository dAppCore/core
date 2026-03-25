package core_test

import (
	"sync"
	"testing"

	. "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
)

// --- Set ---

func TestRegistry_Set_Good(t *testing.T) {
	r := NewRegistry[string]()
	res := r.Set("alpha", "first")
	assert.True(t, res.OK)
	assert.True(t, r.Has("alpha"))
}

func TestRegistry_Set_Good_Update(t *testing.T) {
	r := NewRegistry[string]()
	r.Set("alpha", "first")
	r.Set("alpha", "second")
	res := r.Get("alpha")
	assert.Equal(t, "second", res.Value)
	assert.Equal(t, 1, r.Len(), "update should not increase count")
}

func TestRegistry_Set_Bad_Locked(t *testing.T) {
	r := NewRegistry[string]()
	r.Set("alpha", "first")
	r.Lock()
	res := r.Set("beta", "second")
	assert.False(t, res.OK)
}

func TestRegistry_Set_Bad_SealedNewKey(t *testing.T) {
	r := NewRegistry[string]()
	r.Set("alpha", "first")
	r.Seal()
	res := r.Set("beta", "new")
	assert.False(t, res.OK, "sealed registry must reject new keys")
}

func TestRegistry_Set_Good_SealedExistingKey(t *testing.T) {
	r := NewRegistry[string]()
	r.Set("alpha", "first")
	r.Seal()
	res := r.Set("alpha", "updated")
	assert.True(t, res.OK, "sealed registry must allow updates to existing keys")
	assert.Equal(t, "updated", r.Get("alpha").Value)
}

func TestRegistry_Set_Ugly_ConcurrentWrites(t *testing.T) {
	r := NewRegistry[int]()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			r.Set(Sprintf("key-%d", n), n)
		}(i)
	}
	wg.Wait()
	assert.Equal(t, 100, r.Len())
}

// --- Get ---

func TestRegistry_Get_Good(t *testing.T) {
	r := NewRegistry[string]()
	r.Set("alpha", "value")
	res := r.Get("alpha")
	assert.True(t, res.OK)
	assert.Equal(t, "value", res.Value)
}

func TestRegistry_Get_Bad_NotFound(t *testing.T) {
	r := NewRegistry[string]()
	res := r.Get("missing")
	assert.False(t, res.OK)
}

func TestRegistry_Get_Ugly_EmptyKey(t *testing.T) {
	r := NewRegistry[string]()
	r.Set("", "empty-key")
	res := r.Get("")
	assert.True(t, res.OK, "empty string is a valid key")
	assert.Equal(t, "empty-key", res.Value)
}

// --- Has ---

func TestRegistry_Has_Good(t *testing.T) {
	r := NewRegistry[string]()
	r.Set("alpha", "value")
	assert.True(t, r.Has("alpha"))
}

func TestRegistry_Has_Bad_NotFound(t *testing.T) {
	r := NewRegistry[string]()
	assert.False(t, r.Has("missing"))
}

func TestRegistry_Has_Ugly_AfterDelete(t *testing.T) {
	r := NewRegistry[string]()
	r.Set("alpha", "value")
	r.Delete("alpha")
	assert.False(t, r.Has("alpha"))
}

// --- Names ---

func TestRegistry_Names_Good(t *testing.T) {
	r := NewRegistry[int]()
	r.Set("charlie", 3)
	r.Set("alpha", 1)
	r.Set("bravo", 2)
	assert.Equal(t, []string{"charlie", "alpha", "bravo"}, r.Names(), "must preserve insertion order")
}

func TestRegistry_Names_Bad_Empty(t *testing.T) {
	r := NewRegistry[int]()
	assert.Empty(t, r.Names())
}

func TestRegistry_Names_Ugly_AfterDeleteAndReinsert(t *testing.T) {
	r := NewRegistry[int]()
	r.Set("a", 1)
	r.Set("b", 2)
	r.Set("c", 3)
	r.Delete("b")
	r.Set("d", 4)
	assert.Equal(t, []string{"a", "c", "d"}, r.Names())
}

// --- Each ---

func TestRegistry_Each_Good(t *testing.T) {
	r := NewRegistry[int]()
	r.Set("a", 1)
	r.Set("b", 2)
	r.Set("c", 3)
	var names []string
	var sum int
	r.Each(func(name string, val int) {
		names = append(names, name)
		sum += val
	})
	assert.Equal(t, []string{"a", "b", "c"}, names)
	assert.Equal(t, 6, sum)
}

func TestRegistry_Each_Bad_Empty(t *testing.T) {
	r := NewRegistry[int]()
	called := false
	r.Each(func(_ string, _ int) { called = true })
	assert.False(t, called)
}

func TestRegistry_Each_Ugly_SkipsDisabled(t *testing.T) {
	r := NewRegistry[int]()
	r.Set("a", 1)
	r.Set("b", 2)
	r.Set("c", 3)
	r.Disable("b")
	var names []string
	r.Each(func(name string, _ int) { names = append(names, name) })
	assert.Equal(t, []string{"a", "c"}, names)
}

// --- Len ---

func TestRegistry_Len_Good(t *testing.T) {
	r := NewRegistry[string]()
	assert.Equal(t, 0, r.Len())
	r.Set("a", "1")
	assert.Equal(t, 1, r.Len())
	r.Set("b", "2")
	assert.Equal(t, 2, r.Len())
}

// --- List ---

func TestRegistry_List_Good(t *testing.T) {
	r := NewRegistry[string]()
	r.Set("process.run", "run")
	r.Set("process.start", "start")
	r.Set("agentic.dispatch", "dispatch")
	items := r.List("process.*")
	assert.Len(t, items, 2)
	assert.Contains(t, items, "run")
	assert.Contains(t, items, "start")
}

func TestRegistry_List_Bad_NoMatch(t *testing.T) {
	r := NewRegistry[string]()
	r.Set("alpha", "1")
	items := r.List("beta.*")
	assert.Empty(t, items)
}

func TestRegistry_List_Ugly_SkipsDisabled(t *testing.T) {
	r := NewRegistry[string]()
	r.Set("process.run", "run")
	r.Set("process.kill", "kill")
	r.Disable("process.kill")
	items := r.List("process.*")
	assert.Len(t, items, 1)
	assert.Equal(t, "run", items[0])
}

func TestRegistry_List_Good_WildcardAll(t *testing.T) {
	r := NewRegistry[string]()
	r.Set("a", "1")
	r.Set("b", "2")
	items := r.List("*")
	assert.Len(t, items, 2)
}

// --- Delete ---

func TestRegistry_Delete_Good(t *testing.T) {
	r := NewRegistry[string]()
	r.Set("alpha", "value")
	res := r.Delete("alpha")
	assert.True(t, res.OK)
	assert.False(t, r.Has("alpha"))
	assert.Equal(t, 0, r.Len())
}

func TestRegistry_Delete_Bad_NotFound(t *testing.T) {
	r := NewRegistry[string]()
	res := r.Delete("missing")
	assert.False(t, res.OK)
}

func TestRegistry_Delete_Ugly_Locked(t *testing.T) {
	r := NewRegistry[string]()
	r.Set("alpha", "value")
	r.Lock()
	res := r.Delete("alpha")
	assert.False(t, res.OK, "locked registry must reject delete")
	assert.True(t, r.Has("alpha"), "item must survive failed delete")
}

// --- Disable / Enable ---

func TestRegistry_Disable_Good(t *testing.T) {
	r := NewRegistry[string]()
	r.Set("alpha", "value")
	res := r.Disable("alpha")
	assert.True(t, res.OK)
	assert.True(t, r.Disabled("alpha"))
	// Still exists via Get/Has
	assert.True(t, r.Has("alpha"))
	assert.True(t, r.Get("alpha").OK)
}

func TestRegistry_Disable_Bad_NotFound(t *testing.T) {
	r := NewRegistry[string]()
	res := r.Disable("missing")
	assert.False(t, res.OK)
}

func TestRegistry_Disable_Ugly_EnableRoundtrip(t *testing.T) {
	r := NewRegistry[string]()
	r.Set("alpha", "value")
	r.Disable("alpha")
	assert.True(t, r.Disabled("alpha"))

	res := r.Enable("alpha")
	assert.True(t, res.OK)
	assert.False(t, r.Disabled("alpha"))

	// Verify Each sees it again
	var seen []string
	r.Each(func(name string, _ string) { seen = append(seen, name) })
	assert.Equal(t, []string{"alpha"}, seen)
}

func TestRegistry_Enable_Bad_NotFound(t *testing.T) {
	r := NewRegistry[string]()
	res := r.Enable("missing")
	assert.False(t, res.OK)
}

// --- Lock ---

func TestRegistry_Lock_Good(t *testing.T) {
	r := NewRegistry[string]()
	r.Set("alpha", "value")
	r.Lock()
	assert.True(t, r.Locked())
	// Reads still work
	assert.True(t, r.Get("alpha").OK)
	assert.True(t, r.Has("alpha"))
}

func TestRegistry_Lock_Bad_SetAfterLock(t *testing.T) {
	r := NewRegistry[string]()
	r.Lock()
	res := r.Set("new", "value")
	assert.False(t, res.OK)
}

func TestRegistry_Lock_Ugly_UpdateAfterLock(t *testing.T) {
	r := NewRegistry[string]()
	r.Set("alpha", "first")
	r.Lock()
	res := r.Set("alpha", "second")
	assert.False(t, res.OK, "locked registry must reject even updates")
	assert.Equal(t, "first", r.Get("alpha").Value, "value must not change")
}

// --- Seal ---

func TestRegistry_Seal_Good(t *testing.T) {
	r := NewRegistry[string]()
	r.Set("alpha", "first")
	r.Seal()
	assert.True(t, r.Sealed())
	// Update existing OK
	res := r.Set("alpha", "second")
	assert.True(t, res.OK)
	assert.Equal(t, "second", r.Get("alpha").Value)
}

func TestRegistry_Seal_Bad_NewKey(t *testing.T) {
	r := NewRegistry[string]()
	r.Seal()
	res := r.Set("new", "value")
	assert.False(t, res.OK)
}

func TestRegistry_Seal_Ugly_DeleteWhileSealed(t *testing.T) {
	r := NewRegistry[string]()
	r.Set("alpha", "value")
	r.Seal()
	// Delete is NOT locked by seal — only Set for new keys
	res := r.Delete("alpha")
	assert.True(t, res.OK, "seal blocks new keys, not deletes")
}

// --- Open ---

func TestRegistry_Open_Good(t *testing.T) {
	r := NewRegistry[string]()
	r.Lock()
	assert.True(t, r.Locked())
	r.Open()
	assert.False(t, r.Locked())
	// Can write again
	res := r.Set("new", "value")
	assert.True(t, res.OK)
}

// --- Concurrency ---

func TestRegistry_Ugly_ConcurrentReadWrite(t *testing.T) {
	r := NewRegistry[int]()
	var wg sync.WaitGroup

	// Concurrent writers
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			r.Set(Sprintf("w-%d", n), n)
		}(i)
	}

	// Concurrent readers
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			r.Has(Sprintf("w-%d", n))
			r.Get(Sprintf("w-%d", n))
			r.Names()
			r.Len()
		}(i)
	}

	wg.Wait()
	assert.Equal(t, 50, r.Len())
}
