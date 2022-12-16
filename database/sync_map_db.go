package database

import (
	"sync"
)

type MapDB struct {
	index int
	// key -> DataEntity
	data sync.Map
}

func newMapDB() *MapDB {
	return &MapDB{}
}

func newMapDBWithIndex(index int) *MapDB {
	return &MapDB{index: index}
}

func (db *MapDB) Get(key string) (*DataEntity, bool) {
	raw, exist := db.data.Load(key)
	if !exist {
		return nil, false
	}
	return raw.(*DataEntity), true
}

func (db *MapDB) Put(key string, val *DataEntity) int {
	db.data.Store(key, val)
	return 1
}

func (db *MapDB) PutIfExists(key string, val *DataEntity) int {
	_, exist := db.Get(key)

	if exist {
		db.data.Store(key, val)
		return 1
	} else {
		return 0
	}
}

func (db *MapDB) PutIfAbsent(key string, val *DataEntity) int {
	_, exist := db.Get(key)

	if exist {
		return 0
	} else {
		db.data.Store(key, val)
		return 1
	}
}

func (db *MapDB) remove0(key string) (result int) {
	_, exist := db.Get(key)

	if exist {
		db.data.Delete(key)
		return 1
	} else {
		return 0
	}
}

func (db *MapDB) Remove(key string) {
	db.remove0(key)
}

func (db *MapDB) Removes(keys ...string) int {
	deleted := 0
	for _, key := range keys {
		deleted += db.remove0(key)
	}
	return deleted
}

func (db *MapDB) Flush() {
	db.data = sync.Map{}
}

func (db *MapDB) Size() int {
	l := 0
	db.data.Range(func(_, _ any) bool {
		l++
		return true
	})
	return l
}

func (db *MapDB) ForEach(traverser func(key string, val *DataEntity) bool) {
	db.data.Range(func(key, value any) bool {
		return traverser(key.(string), value.(*DataEntity))
	})
}

func (db *MapDB) Keys() []string {
	keys := make([]string, db.Size())
	i := 0
	db.data.Range(func(key, value any) bool {
		keys[i] = key.(string)
		i++
		return true
	})
	return keys
}

func (db *MapDB) RandomKeys(limit int) []string {
	keys := make([]string, limit)
	for i := 0; i < limit; i++ {
		db.data.Range(func(key, value any) bool {
			keys[i] = key.(string)
			return false
		})
	}
	return keys
}

func (db *MapDB) RandomDistinctKeys(limit int) []string {
	keys := make([]string, limit)
	i := 0
	db.data.Range(func(key, value any) bool {
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
