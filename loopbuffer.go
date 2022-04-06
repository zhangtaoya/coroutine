// Copyright 2022 Zhangtaoya. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package coroutine

import (
	"github.com/mohae/deepcopy"
	"sync"
)

type LoopBuffer struct {
	sync.RWMutex
	KeyData map[uint64]map[string]interface{}
	KeyList []uint64
	Cnt     int
}

func (t *LoopBuffer) Reset(sz int) {
	t.Lock()
	defer t.Unlock()
	t.KeyList = make([]uint64, sz)
	t.KeyData = make(map[uint64]map[string]interface{})
	t.Cnt = 0
}

func (t *LoopBuffer) SetVal(dataId uint64, k string, val interface{}) {
	if dataId < 0 {
		return
	}
	t.Lock()
	defer t.Unlock()

	// 如果这个dataId对应的数据存在，直接存到这个map里就行
	if data, ok := t.KeyData[dataId]; ok {
		if data == nil {
			t.KeyData[dataId] = map[string]interface{}{k: val}
		} else {
			t.KeyData[dataId][k] = val
		}
		return
	}

	t.setDataUsingBufferNoLock(dataId, map[string]interface{}{k: val})
}

func (t *LoopBuffer) GetVal(dataId uint64, k string) (val interface{}, has bool) {
	if dataId < 0 {
		return
	}
	t.RLock()
	defer t.RUnlock()
	dataMap, has := t.KeyData[dataId]
	if !has {
		return
	}
	if dataMap == nil {
		has = false
		return
	}
	val, has = dataMap[k]
	return
}

func (t *LoopBuffer) setDataUsingBufferNoLock(dataId uint64, data map[string]interface{}) {

	// 得到需要覆盖的IdCur, 删掉对应gid的在map中的数据
	// idcur指向新位置
	t.Cnt += 1
	idxCurOverWrite := t.Cnt % len(t.KeyList)

	dataIdOverwrite := t.KeyList[idxCurOverWrite]
	if len(t.KeyData) > 0 {
		delete(t.KeyData, dataIdOverwrite)
	}

	// 需要存的数据存储进去
	t.KeyList[idxCurOverWrite] = dataId
	t.KeyData[dataId] = data

	return
}

func (t *LoopBuffer) DelData(dataId uint64) {
	if dataId < 0 {
		return
	}
	t.Lock()
	defer t.Unlock()
	if len(t.KeyData) > 0 {
		delete(t.KeyData, dataId)
	}
	return
}

func (t *LoopBuffer) SetData(dataId uint64, data map[string]interface{}) {
	if dataId < 0 {
		return
	}
	t.Lock()
	defer t.Unlock()

	// 如果这个dataId对应的数据存在，直接替换这个map就行
	if _, ok := t.KeyData[dataId]; ok {
		t.KeyData[dataId] = data
		return
	}

	t.setDataUsingBufferNoLock(dataId, data)
	return
}

func (t *LoopBuffer) GetDataCpy(dataId uint64) (data map[string]interface{}, has bool) {
	if dataId < 0 {
		return
	}
	t.RLock()
	defer t.RUnlock()

	dataTmp, has := t.KeyData[dataId]
	data = make(map[string]interface{})
	for k, v := range dataTmp {
		data[k] = deepcopy.Copy(v)
	}
	return
}
