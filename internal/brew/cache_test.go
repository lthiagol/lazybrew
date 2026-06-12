package brew

import (
	"testing"
	"time"
)

func TestCacheGetSet(t *testing.T) {
	c := NewCache(time.Minute)
	c.Set(KeyFormulaeList, []Formula{{Name: "ripgrep"}})
	val, ok := c.Get(KeyFormulaeList)
	if !ok {
		t.Fatal("expected cache hit")
	}
	formulae, ok := val.([]Formula)
	if !ok {
		t.Fatal("expected []Formula")
	}
	if len(formulae) != 1 || formulae[0].Name != "ripgrep" {
		t.Errorf("got %+v, want [ripgrep]", formulae)
	}
}

func TestCacheMiss(t *testing.T) {
	c := NewCache(time.Minute)
	_, ok := c.Get("nonexistent")
	if ok {
		t.Fatal("expected cache miss")
	}
}

func TestCacheTTL(t *testing.T) {
	c := NewCache(50 * time.Millisecond)
	c.Set(KeyFormulaeList, "data")
	time.Sleep(100 * time.Millisecond)
	_, ok := c.Get(KeyFormulaeList)
	if ok {
		t.Error("expected cache miss after TTL expiry")
	}
}

func TestCacheInvalidate(t *testing.T) {
	c := NewCache(time.Minute)
	c.Set("key1", "value1")
	c.Set("key2", "value2")
	c.Invalidate("key1")
	_, ok1 := c.Get("key1")
	if ok1 {
		t.Error("expected key1 to be invalidated")
	}
	_, ok2 := c.Get("key2")
	if !ok2 {
		t.Error("expected key2 to still be present")
	}
}

func TestCacheInvalidateFor(t *testing.T) {
	c := NewCache(time.Minute)
	c.Set(KeyFormulaeList, "f")
	c.Set(KeyOutdatedFormulae, "o")
	c.InvalidateFor("upgrade")
	_, fOk := c.Get(KeyOutdatedFormulae)
	_, oOk := c.Get(KeyOutdatedCasks)
	if fOk || oOk {
		t.Error("expected both keys invalidated by upgrade group")
	}
}

func TestCacheInvalidateAll(t *testing.T) {
	c := NewCache(time.Minute)
	c.Set("a", 1)
	c.Set("b", 2)
	c.InvalidateAll()
	_, ok := c.Get("a")
	if ok {
		t.Error("expected all keys invalidated")
	}
}

func TestCacheConcurrency(t *testing.T) {
	c := NewCache(time.Minute)
	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			c.Set(KeyFormulaeList, i)
			c.Get(KeyFormulaeList)
			c.Invalidate(KeyFormulaeList)
		}
		done <- struct{}{}
	}()
	go func() {
		for i := 0; i < 100; i++ {
			c.Get(KeyFormulaeList)
			c.InvalidateAll()
			c.Set(KeyOutdatedFormulae, i)
		}
		done <- struct{}{}
	}()
	<-done
	<-done
}

func TestCacheZeroTTL(t *testing.T) {
	c := NewCache(0)
	c.Set(KeyFormulaeList, "data")
	time.Sleep(time.Millisecond)
	_, ok := c.Get(KeyFormulaeList)
	if ok {
		t.Error("expected cache miss with zero TTL")
	}
}

func TestCacheConcurrentExpiry(t *testing.T) {
	c := NewCache(50 * time.Millisecond)
	done := make(chan struct{})
	for i := 0; i < 4; i++ {
		go func(idx int) {
			for j := 0; j < 50; j++ {
				c.Set(KeyFormulaeList, j)
				c.Get(KeyFormulaeList)
				time.Sleep(time.Millisecond)
			}
			done <- struct{}{}
		}(i)
	}
	for i := 0; i < 4; i++ {
		<-done
	}
}

func TestCacheSeparateOutdatedKeys(t *testing.T) {
	c := NewCache(time.Minute)
	c.Set(KeyOutdatedFormulae, []Formula{{Name: "ripgrep"}})
	c.Set(KeyOutdatedCasks, []Cask{{Name: "firefox"}})

	f, fOk := c.Get(KeyOutdatedFormulae)
	if !fOk {
		t.Fatal("expected KeyOutdatedFormulae hit")
	}
	formulae, _ := f.([]Formula)
	if len(formulae) != 1 || formulae[0].Name != "ripgrep" {
		t.Error("wrong formulae cached")
	}

	f2, f2Ok := c.Get(KeyOutdatedCasks)
	if !f2Ok {
		t.Fatal("expected KeyOutdatedCasks hit")
	}
	casks, _ := f2.([]Cask)
	if len(casks) != 1 || casks[0].Name != "firefox" {
		t.Error("wrong casks cached")
	}
}
