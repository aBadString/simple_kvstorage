package datastructure

import "sync"

type SyncDict struct {
	m sync.Map
}

func NewSyncDict() *SyncDict {
	return &SyncDict{}
}

func (d *SyncDict) Get(key string) (val interface{}, exists bool) {
	return d.m.Load(key)
}

func (d *SyncDict) Len() int {
	l := 0
	d.m.Range(func(_, _ any) bool {
		l++
		return true
	})
	return l
}

func (d *SyncDict) Put(key string, val interface{}) (result int) {
	d.m.Store(key, val)
	return 1
}

func (d *SyncDict) PutIfAbsent(key string, val interface{}) (result int) {
	_, exist := d.Get(key)

	if exist {
		return 0
	} else {
		d.m.Store(key, val)
		return 1
	}
}

func (d *SyncDict) PutIfExists(key string, val interface{}) (result int) {
	_, exist := d.Get(key)

	if exist {
		d.m.Store(key, val)
		return 1
	} else {
		return 0
	}
}

func (d *SyncDict) Remove(key string) (result int) {
	_, exist := d.Get(key)

	if exist {
		d.m.Delete(key)
		return 1
	} else {
		return 0
	}
}

func (d *SyncDict) ForEach(traverser Traverser) {
	d.m.Range(func(key, value any) bool {
		return traverser(key.(string), value)
	})
}

func (d *SyncDict) Keys() []string {
	keys := make([]string, d.Len())
	i := 0
	d.m.Range(func(key, value any) bool {
		keys[i] = key.(string)
		i++
		return true
	})
	return keys
}

func (d *SyncDict) RandomKeys(limit int) []string {
	keys := make([]string, limit)
	for i := 0; i < limit; i++ {
		d.m.Range(func(key, value any) bool {
			keys[i] = key.(string)
			return false
		})
	}
	return keys
}

func (d *SyncDict) RandomDistinctKeys(limit int) []string {
	keys := make([]string, limit)
	i := 0
	d.m.Range(func(key, value any) bool {
		keys[i] = key.(string)
		i++
		if i < limit {
			return true
		} else {
			return false
		}
	})
	return keys
}

func (d *SyncDict) Clear() {
	*d = *NewSyncDict()
}
