// Copyright 2022 Zhangtaoya. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package coroutine

import (
	"fmt"
	"runtime"
	"time"
)

var maxSize = 100 * 1000
var glsMonitorInterval = 10
var addrGid *LoopBuffer
var glsData *LoopBuffer

var gidFake uint64

func init() {
	addrGid = &LoopBuffer{}
	glsData = &LoopBuffer{}

	addrGid.Reset(maxSize)
	glsData.Reset(maxSize)
	if getRoutineAddr() == 0 {
		Warn("gls using stack mode to get gid. low performance")
	} else {
		Info("gls using cache mode to get gid. high performance")
	}
	go gonumscan()
}

// coid 协程id 保证当前协程下此id是自己的，有可能会复用之前的goid，所以并不是goid
func GetCoid() uint64 {
	return getGidByCache()
	//return GetGidNoCache()
}

// 设置k v
func SetVal(k string, val interface{}) {
	glsData.SetVal(GetCoid(), k, val)
}

// 获取k对应的数据
func GetVal(k string) (gid uint64, val interface{}, ok bool) {
	gid = GetCoid()
	val, ok = glsData.GetVal(gid, k)
	return
}

// 获取所有的gls数据，即所有kv的一个map
func GetGlsDataCpy() (gid uint64, data map[string]interface{}, ok bool) {
	gid = GetCoid()
	data, ok = glsData.GetDataCpy(gid)
	return
}

// 设置当前的gls数据，即覆盖(如果存在)之前的gls数据
func SetGlsData(data map[string]interface{}) () {
	glsData.SetData(GetCoid(), data)
	return
}

// 清除当前的gls数据，在一个协程使用gls数据之前，需要先清掉(或者用SetGlsData覆盖)可能存在的残余(见GetGidByCache函数说明)，否则直接访问可能读取到之前留下的数据
func ClearGlsData() {
	// 设置成nil，不会删除key，让key占位复用老的位置，以优化存储
	glsData.SetData(GetCoid(), nil)
	//glsData.DelData(GetCoid())
	return
}

// 可选功能。外部可以调用这个设置监控间隔，单位为秒，默认为10秒
func SetGlsMonitorInterval(i int) {
	if i <= 0 {
		return
	}
	glsMonitorInterval = i
}

//可选功能。外部可以调用这个设置buffer大小，默认为10万，可支持10万go协程gls，超过10万会出现gls数据丢失问题。可以run不会崩，但结果不可靠
func ResetSize(sz int) {
	maxSize = sz
	addrGid.Reset(maxSize)
	glsData.Reset(maxSize)
}

// 内部监控，不对外调用
func gonumscan() {
	for {
		n := runtime.NumGoroutine()
		addr := getRoutineAddr()
		peftag := fmt.Sprintf("addr:0x%x cache mod gid. high performance.", addr)
		if addr == 0 {
			peftag = "addr:0 stack mode gid, low performance."
		}
		cntAddr := addrGid.Cnt
		cntGls := glsData.Cnt
		peftag += fmt.Sprintf("addr_cnt:%d, gls_cnt:%d", cntAddr, cntGls)
		if n < maxSize*8/10 {
			Info("NumGoroutine:%d <  gls max size:%d * 0.8, gls safe(%s. goidoffset:%d, Go version:%s)",
				n, maxSize, peftag, goidOffset, goversion)
		} else {
			Error("NumGoroutine:%d >= gls max size:%d * 0.8, gls unsafe, plz add sz(ResetSize)(%s.  goidoffset:%d, Go version:%s)",
				n, maxSize, peftag, goidOffset, goversion)
		}
		time.Sleep(time.Second * time.Duration(glsMonitorInterval))
	}
}
