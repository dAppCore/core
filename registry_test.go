package core_test

import (
	"sync"

	. "dappco.re/go/core"
)

// --- Set ---

func TestRegistry_Set_Good(t *T) {
	r := NewRegistry[string]()
	res := r.Set("alpha", "first")
	AssertTrue(t, res.OK)
	AssertTrue(t, r.Has("alpha"))
}

func TestRegistry_Set_Good_Update(t *T) {
	r := NewRegistry[string]()
	r.Set("alpha", "first")
	r.Set("alpha", "second")
	res := r.Get("alpha")
	AssertEqual(t, "second", res.Value)
	AssertEqual(t, 1, r.Len(), "update should not increase count")
}

func TestRegistry_Set_Bad_Locked(t *T) {
	r := NewRegistry[string]()
	r.Set("alpha", "first")
	r.Lock()
	res := r.Set("beta", "second")
	AssertFalse(t, res.OK)
}

func TestRegistry_Set_Bad_SealedNewKey(t *T) {
	r := NewRegistry[string]()
	r.Set("alpha", "first")
	r.Seal()
	res := r.Set("beta", "new")
	AssertFalse(t, res.OK, "sealed registry must reject new keys")
}

func TestRegistry_Set_Good_SealedExistingKey(t *T) {
	r := NewRegistry[string]()
	r.Set("alpha", "first")
	r.Seal()
	res := r.Set("alpha", "updated")
	AssertTrue(t, res.OK, "sealed registry must allow updates to existing keys")
	AssertEqual(t, "updated", r.Get("alpha").Value)
}

func TestRegistry_Set_Ugly_ConcurrentWrites(t *T) {
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
	AssertEqual(t, 100, r.Len())
}

// --- Get ---

func TestRegistry_Get_Good(t *T) {
	r := NewRegistry[string]()
	r.Set("alpha", "value")
	res := r.Get("alpha")
	AssertTrue(t, res.OK)
	AssertEqual(t, "value", res.Value)
}

func TestRegistry_Get_Bad_NotFound(t *T) {
	r := NewRegistry[string]()
	res := r.Get("missing")
	AssertFalse(t, res.OK)
}

func TestRegistry_Get_Ugly_EmptyKey(t *T) {
	r := NewRegistry[string]()
	r.Set("", "empty-key")
	res := r.Get("")
	AssertTrue(t, res.OK, "empty string is a valid key")
	AssertEqual(t, "empty-key", res.Value)
}

// --- Has ---

func TestRegistry_Has_Good(t *T) {
	r := NewRegistry[string]()
	r.Set("alpha", "value")
	AssertTrue(t, r.Has("alpha"))
}

func TestRegistry_Has_Bad_NotFound(t *T) {
	r := NewRegistry[string]()
	AssertFalse(t, r.Has("missing"))
}

func TestRegistry_Has_Ugly_AfterDelete(t *T) {
	r := NewRegistry[string]()
	r.Set("alpha", "value")
	r.Delete("alpha")
	AssertFalse(t, r.Has("alpha"))
}

// --- Names ---

func TestRegistry_Names_Good(t *T) {
	r := NewRegistry[int]()
	r.Set("charlie", 3)
	r.Set("alpha", 1)
	r.Set("bravo", 2)
	AssertEqual(t, []string{"charlie", "alpha", "bravo"}, r.Names(), "must preserve insertion order")
}

func TestRegistry_Names_Bad_Empty(t *T) {
	r := NewRegistry[int]()
	AssertEmpty(t, r.Names())
}

func TestRegistry_Names_Ugly_AfterDeleteAndReinsert(t *T) {
	r := NewRegistry[int]()
	r.Set("a", 1)
	r.Set("b", 2)
	r.Set("c", 3)
	r.Delete("b")
	r.Set("d", 4)
	AssertEqual(t, []string{"a", "c", "d"}, r.Names())
}

// --- Each ---

func TestRegistry_Each_Good(t *T) {
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
	AssertEqual(t, []string{"a", "b", "c"}, names)
	AssertEqual(t, 6, sum)
}

func TestRegistry_Each_Bad_Empty(t *T) {
	r := NewRegistry[int]()
	called := false
	r.Each(func(_ string, _ int) { called = true })
	AssertFalse(t, called)
}

func TestRegistry_Each_Ugly_SkipsDisabled(t *T) {
	r := NewRegistry[int]()
	r.Set("a", 1)
	r.Set("b", 2)
	r.Set("c", 3)
	r.Disable("b")
	var names []string
	r.Each(func(name string, _ int) { names = append(names, name) })
	AssertEqual(t, []string{"a", "c"}, names)
}

// --- Len ---

func TestRegistry_Len_Good(t *T) {
	r := NewRegistry[string]()
	AssertEqual(t, 0, r.Len())
	r.Set("a", "1")
	AssertEqual(t, 1, r.Len())
	r.Set("b", "2")
	AssertEqual(t, 2, r.Len())
}

// --- List ---

func TestRegistry_List_Good(t *T) {
	r := NewRegistry[string]()
	r.Set("process.run", "run")
	r.Set("process.start", "start")
	r.Set("agentic.dispatch", "dispatch")
	items := r.List("process.*")
	AssertLen(t, items, 2)
	AssertContains(t, items, "run")
	AssertContains(t, items, "start")
}

func TestRegistry_List_Bad_NoMatch(t *T) {
	r := NewRegistry[string]()
	r.Set("alpha", "1")
	items := r.List("beta.*")
	AssertEmpty(t, items)
}

func TestRegistry_List_Ugly_SkipsDisabled(t *T) {
	r := NewRegistry[string]()
	r.Set("process.run", "run")
	r.Set("process.kill", "kill")
	r.Disable("process.kill")
	items := r.List("process.*")
	AssertLen(t, items, 1)
	AssertEqual(t, "run", items[0])
}

func TestRegistry_List_Good_WildcardAll(t *T) {
	r := NewRegistry[string]()
	r.Set("a", "1")
	r.Set("b", "2")
	items := r.List("*")
	AssertLen(t, items, 2)
}

// --- Delete ---

func TestRegistry_Delete_Good(t *T) {
	r := NewRegistry[string]()
	r.Set("alpha", "value")
	res := r.Delete("alpha")
	AssertTrue(t, res.OK)
	AssertFalse(t, r.Has("alpha"))
	AssertEqual(t, 0, r.Len())
}

func TestRegistry_Delete_Bad_NotFound(t *T) {
	r := NewRegistry[string]()
	res := r.Delete("missing")
	AssertFalse(t, res.OK)
}

func TestRegistry_Delete_Ugly_Locked(t *T) {
	r := NewRegistry[string]()
	r.Set("alpha", "value")
	r.Lock()
	res := r.Delete("alpha")
	AssertFalse(t, res.OK, "locked registry must reject delete")
	AssertTrue(t, r.Has("alpha"), "item must survive failed delete")
}

// --- Disable / Enable ---

func TestRegistry_Disable_Good(t *T) {
	r := NewRegistry[string]()
	r.Set("alpha", "value")
	res := r.Disable("alpha")
	AssertTrue(t, res.OK)
	AssertTrue(t, r.Disabled("alpha"))
	// Still exists via Get/Has
	AssertTrue(t, r.Has("alpha"))
	AssertTrue(t, r.Get("alpha").OK)
}

func TestRegistry_Disable_Bad_NotFound(t *T) {
	r := NewRegistry[string]()
	res := r.Disable("missing")
	AssertFalse(t, res.OK)
}

func TestRegistry_Disable_Ugly_EnableRoundtrip(t *T) {
	r := NewRegistry[string]()
	r.Set("alpha", "value")
	r.Disable("alpha")
	AssertTrue(t, r.Disabled("alpha"))

	res := r.Enable("alpha")
	AssertTrue(t, res.OK)
	AssertFalse(t, r.Disabled("alpha"))

	// Verify Each sees it again
	var seen []string
	r.Each(func(name string, _ string) { seen = append(seen, name) })
	AssertEqual(t, []string{"alpha"}, seen)
}

func TestRegistry_Enable_Bad_NotFound(t *T) {
	r := NewRegistry[string]()
	res := r.Enable("missing")
	AssertFalse(t, res.OK)
}

// --- Lock ---

func TestRegistry_Lock_Good(t *T) {
	r := NewRegistry[string]()
	r.Set("alpha", "value")
	r.Lock()
	AssertTrue(t, r.Locked())
	// Reads still work
	AssertTrue(t, r.Get("alpha").OK)
	AssertTrue(t, r.Has("alpha"))
}

func TestRegistry_Lock_Bad_SetAfterLock(t *T) {
	r := NewRegistry[string]()
	r.Lock()
	res := r.Set("new", "value")
	AssertFalse(t, res.OK)
}

func TestRegistry_Lock_Ugly_UpdateAfterLock(t *T) {
	r := NewRegistry[string]()
	r.Set("alpha", "first")
	r.Lock()
	res := r.Set("alpha", "second")
	AssertFalse(t, res.OK, "locked registry must reject even updates")
	AssertEqual(t, "first", r.Get("alpha").Value, "value must not change")
}

// --- Seal ---

func TestRegistry_Seal_Good(t *T) {
	r := NewRegistry[string]()
	r.Set("alpha", "first")
	r.Seal()
	AssertTrue(t, r.Sealed())
	// Update existing OK
	res := r.Set("alpha", "second")
	AssertTrue(t, res.OK)
	AssertEqual(t, "second", r.Get("alpha").Value)
}

func TestRegistry_Seal_Bad_NewKey(t *T) {
	r := NewRegistry[string]()
	r.Seal()
	res := r.Set("new", "value")
	AssertFalse(t, res.OK)
}

func TestRegistry_Seal_Ugly_DeleteWhileSealed(t *T) {
	r := NewRegistry[string]()
	r.Set("alpha", "value")
	r.Seal()
	// Delete is NOT locked by seal — only Set for new keys
	res := r.Delete("alpha")
	AssertTrue(t, res.OK, "seal blocks new keys, not deletes")
}

// --- Open ---

func TestRegistry_Open_Good(t *T) {
	r := NewRegistry[string]()
	r.Lock()
	AssertTrue(t, r.Locked())
	r.Open()
	AssertFalse(t, r.Locked())
	// Can write again
	res := r.Set("new", "value")
	AssertTrue(t, res.OK)
}

// --- Concurrency ---

func TestRegistry_Ugly_ConcurrentReadWrite(t *T) {
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
	AssertEqual(t, 50, r.Len())
}
