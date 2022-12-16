package database

import "testing"

var testCases = []struct {
	key string
	val *DataEntity
}{
	{"1", &DataEntity{Data: "hello"}},
	{"2", &DataEntity{Data: 2}},
	{"3", &DataEntity{Data: '3'}},
	{"4", &DataEntity{Data: 4.1}},
	{"5", &DataEntity{Data: "world"}},
}

func TestMapDB(t *testing.T) {
	const firstPutNum = 3

	db := newMapDB()

	for i := 0; i < firstPutNum; i++ {
		db.Put(testCases[i].key, testCases[i].val)
	}

	if db.Size() != firstPutNum {
		t.Error("Put or Size 方法测试失败.")
		return
	}

	for i := 0; i < firstPutNum; i++ {
		val, exists := db.Get(testCases[i].key)
		if !exists || testCases[i].val != val {
			t.Error("Get 方法测试失败.")
			return
		}
	}

	for i := 0; i < firstPutNum; i++ {
		entity := &DataEntity{"PutIfExists"}
		db.PutIfExists(testCases[i].key, entity)

		val, exists := db.Get(testCases[i].key)
		if !exists || entity != val {
			t.Error("PutIfExists 方法测试失败 (when exists).")
			return
		}
	}
	for i := firstPutNum; i < len(testCases); i++ {
		entity := &DataEntity{Data: "PutIfExists"}
		db.PutIfExists(testCases[i].key, entity)

		_, exists := db.Get(testCases[i].key)
		if exists {
			t.Error("PutIfExists 方法测试失败 (when not exists).")
			return
		}
	}

	for i := 0; i < firstPutNum; i++ {
		entity := &DataEntity{Data: "PutIfAbsent"}
		db.PutIfAbsent(testCases[i].key, entity)

		val, _ := db.Get(testCases[i].key)
		if entity == val {
			t.Error("PutIfExists 方法测试失败 (when exists).")
			return
		}
	}
	for i := firstPutNum; i < len(testCases); i++ {
		entity := &DataEntity{Data: "PutIfAbsent"}
		db.PutIfAbsent(testCases[i].key, entity)

		val, exists := db.Get(testCases[i].key)
		if !exists || entity != val {
			t.Error("PutIfExists 方法测试失败 (when not exists).")
			return
		}
	}

	keys := db.Keys()
	t.Log(keys)
	if db.Size() != len(keys) {
		t.Error("Keys 方法测试失败.")
		return
	}

	db.Remove(keys[0])
	if db.Size() != len(keys)-1 {
		t.Error("Remove 方法测试失败.")
		return
	}

	removedNum := db.Removes(keys...)
	if db.Size() != 0 && removedNum != len(keys)-1 {
		t.Error("Removes 方法测试失败.")
		return
	}
}

func TestMapDB_ForEach(t *testing.T) {
	db := newMapDB()

	for _, testCase := range testCases {
		db.Put(testCase.key, testCase.val)
	}

	db.ForEach(func(key string, val *DataEntity) bool {
		t.Log(key, val)
		return true
	})

	randomKeys := db.RandomKeys(10)
	t.Log(randomKeys)

	distinctKeys := db.RandomDistinctKeys(4)
	t.Log(distinctKeys)
	distinctKeys = db.RandomDistinctKeys(10)
	t.Log(distinctKeys)
}

func TestMapDB_Flush(t *testing.T) {
	db := newMapDBWithIndex(15)

	for _, testCase := range testCases {
		db.Put(testCase.key, testCase.val)
	}

	db.Flush()
	if db.Size() != 0 || db.index != 15 {
		t.Error("Flush 方法测试失败.")
		return
	}

	t.Log(db)
}
