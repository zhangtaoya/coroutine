// Copyright 2022 Zhangtaoya. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package coroutine

import "sync"

type LoopBuffer struct {
	sync.RWMutex
	ValMap   map[uint64]interface{}
	CoidList []uint64
	IdCur    int
}

func (t *LoopBuffer) Reset(sz int) {
	t.CoidList = make([]uint64, sz)
	t.IdCur = 0
	t.ValMap = make(map[uint64]interface{})
}

func (t *LoopBuffer) ClearVal(dataId uint64) {
	if dataId < 0 {
		return
	}
	t.Lock()
	defer t.Unlock()
	if len(t.ValMap) > 0 {
		delete(t.ValMap, dataId)
	}
	return
}

func (t *LoopBuffer) SetVal(dataId uint64, val interface{}) {
	if dataId < 0 {
		return
	}
	t.Lock()
	defer t.Unlock()

	// 得到需要覆盖的IdCur, 删掉对应gid的在map中的数据
	idxCurOverWrite := t.IdCur + 1
	if idxCurOverWrite >= len(t.CoidList) {
		idxCurOverWrite = 0
	}

	dataIdOverwrite := t.CoidList[idxCurOverWrite]
	if len(t.ValMap) > 0 {
		delete(t.ValMap, dataIdOverwrite)
	}

	// 需要存的数据存储进去
	t.CoidList[idxCurOverWrite] = dataId
	t.ValMap[dataId] = val

	// idcur指向新位置
	t.IdCur = idxCurOverWrite
}

func (t *LoopBuffer) GetVal(k uint64) (val interface{}, has bool) {
	if k < 0 {
		return
	}
	t.RLock()
	defer t.RUnlock()
	val, has = t.ValMap[k]
	return
}
