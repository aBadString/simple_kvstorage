package datastructure

import "testing"

var testCases = []struct {
	key string
	val interface{}
}{
	{"1", "hello"},
	{"2", 2},
	{"3", '3'},
	{"4", 4.1},
	{"5", "world"},
}

func TestSyncDict(t *testing.T) {
	const firstPutNum = 3

	dict := NewSyncDict()

	for i := 0; i < firstPutNum; i++ {
		dict.Put(testCases[i].key, testCases[i].val)
	}

	if dict.Len() != firstPutNum {
		t.Error("Put or Len 方法测试失败.")
		return
	}

	for i := 0; i < firstPutNum; i++ {
		val, exists := dict.Get(testCases[i].key)
		if !exists || testCases[i].val != val {
			t.Error("Get 方法测试失败.")
			return
		}
	}

	for i := 0; i < firstPutNum; i++ {
		dict.PutIfExists(testCases[i].key, "PutIfExists")

		val, exists := dict.Get(testCases[i].key)
		if !exists || "PutIfExists" != val {
			t.Error("PutIfExists 方法测试失败 (when exists).")
			return
		}
	}
	for i := firstPutNum; i < len(testCases); i++ {
		dict.PutIfExists(testCases[i].key, "PutIfExists")

		_, exists := dict.Get(testCases[i].key)
		if exists {
			t.Error("PutIfExists 方法测试失败 (when not exists).")
			return
		}
	}

	for i := 0; i < firstPutNum; i++ {
		dict.PutIfAbsent(testCases[i].key, "PutIfAbsent")

		val, _ := dict.Get(testCases[i].key)
		if "PutIfAbsent" == val {
			t.Error("PutIfExists 方法测试失败 (when exists).")
			return
		}
	}
	for i := firstPutNum; i < len(testCases); i++ {
		dict.PutIfAbsent(testCases[i].key, "PutIfAbsent")

		val, exists := dict.Get(testCases[i].key)
		if !exists || "PutIfAbsent" != val {
			t.Error("PutIfExists 方法测试失败 (when not exists).")
			return
		}
	}

	keys := dict.Keys()
	t.Log(keys)
	if dict.Len() != len(keys) {
		t.Error("Keys 方法测试失败.")
		return
	}

	for _, key := range keys {
		dict.Remove(key)
	}
	if dict.Len() != 0 {
		t.Error("Remove 方法测试失败.")
		return
	}

	dict.Clear()
	if dict.Len() != 0 {
		t.Error("Clear 方法测试失败.")
		return
	}
}

func TestSyncDict_ForEach(t *testing.T) {
	dict := NewSyncDict()

	for _, testCase := range testCases {
		dict.Put(testCase.key, testCase.val)
	}

	dict.ForEach(func(key string, val interface{}) bool {
		t.Log(key, val)
		return true
	})

	randomKeys := dict.RandomKeys(10)
	t.Log(randomKeys)

	distinctKeys := dict.RandomDistinctKeys(4)
	t.Log(distinctKeys)
	distinctKeys = dict.RandomDistinctKeys(10)
	t.Log(distinctKeys)
}
