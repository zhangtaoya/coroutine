// Copyright 2022 Zhangtaoya. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package coroutine

import (
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mohae/deepcopy"

	"git.ixiaochuan.cn/mama/server/lib/coroutine/g"
	"git.ixiaochuan.cn/mama/server/lib/coroutine/logger"
)

var maxSize = 100 * 1000
var addrGid *LoopBuffer
var glsData *LoopBuffer
var glsDataLock sync.RWMutex
var glblock sync.RWMutex

var gidFake uint64

func init() {
	addrGid = &LoopBuffer{}
	glsData = &LoopBuffer{}

	addrGid.Reset(maxSize)
	glsData.Reset(maxSize)
	if getRoutineAddr() == 0 {
		logger.Warn("gls using stack mode to get gid. low performance")
	} else {
		logger.Info("gls using cache mode to get gid. high performance")
	}
	go gonumscan()
}

func gonumscan() {
	for range time.Tick(time.Second * time.Duration(10)) {
		n := runtime.NumGoroutine()
		peftag := "addr cache mod gid. high performance"
		if getRoutineAddr() == 0 {
			peftag = "stack mode gid, low performance"
		}
		if n < maxSize*8/10 {
			logger.Info("NumGoroutine:%d <  gls max size:%d * 0.8, gls safe(%s)", n, maxSize, peftag)
		} else {
			logger.Error("NumGoroutine:%d >= gls max size:%d * 0.8, gls unsafe, plz add sz(ResetSize)(%s)", n, maxSize, peftag)
		}

	}
}
func ResetSize(sz int) {
	glblock.Lock()
	defer glblock.Unlock()

	maxSize = sz
	addrGid.Reset(maxSize)
	glsData.Reset(maxSize)
}

func getRoutineAddr() uint64 {
	// 这个函数取的只是当前协程的一个唯一标识
	// 用协程地址值当这个标识
	p := g.G()
	if p == nil {
		return 0
	}
	x, _ := strconv.ParseUint(fmt.Sprintf("%v", p)[2:], 16, 64)
	return x
}

// coid 协程id 保证当前协程下此id是自己的，有可能会复用之前的goid，所以并不是goid
func GetCoid() uint64 {
	return GetGidByCache()
	//return GetGidNoCache()
}

func GetGidNoCache() uint64 {
	if gid, ok := getGoidByNative(); ok {
		return gid
	}
	if gid := getGoidByStack(); gid >= 0 {
		return gid
	}
	return 0
}

// 以g数据地址为k，对gid做缓存。注意由于gc回收g后可能再给下一个协程用，gid会有重复使用。
// 如果使用该方法做gls，使用tls数据前需先重置或覆盖当前gid的tls数据，直接取有可能是之前这个地址的gid留下的脏数据
// warning: gls数据一定在该协程周期内使用，如果脱离该协程生命周期，然后再访问它的gls，有可能它的gls已经被修改

// 当获取g数据地址失败时，该方法会降级为堆栈方式获取gid
// 当获取g数据地址失败，且堆栈方式获取gid也失效时，该方法失效
// 实际取gid时
// 首先采用g结构获取，
// 如果失败，则用堆栈方式获取(性能低)，但由于做了缓存，所以还好
// 如果还失败，则返回一个内部自增的id，只要g地址不变，gid就不会变。
func GetGidByCache() uint64 {
	addr := getRoutineAddr()
	if addr == 0 {
		return getGoidByStack()
	}

	if val, ok := addrGid.GetVal(addr); ok && val != nil {
		gid := val.(uint64)
		return gid
	}

	if gid, ok := getGoidByNative(); ok {
		addrGid.SetVal(addr, gid)
		return gid
	}

	if gid := getGoidByStack(); gid >= 0 {
		addrGid.SetVal(addr, gid)
		return gid
	}

	//
	gid := atomic.AddUint64(&gidFake, 1)
	addrGid.SetVal(addr, gid)
	return gid
}

func SetVal(k string, val interface{}) {
	gid := GetCoid()

	glsDataLock.Lock()
	defer glsDataLock.Unlock()

	dataRaw, ok := glsData.GetVal(gid)
	if !ok {
		dataRaw = make(map[string]interface{})
	}

	data := dataRaw.(map[string]interface{})
	data[k] = val
	glsData.SetVal(gid, data)
}

func GetVal(k string) (gid uint64, val interface{}, ok bool) {
	gid = GetCoid()

	dataRaw, ok := glsData.GetVal(gid)
	if !ok {
		return
	}
	data := dataRaw.(map[string]interface{})
	val, ok = data[k]
	return
}

func GetGlsDataCpy() (val interface{}, ok bool) {
	gid := GetCoid()
	val, ok = glsData.GetVal(gid)
	val = deepcopy.Copy(val)
	return
}

func SetGlsData(val interface{}) () {
	gid := GetCoid()
	glsData.SetVal(gid, val)
	return
}

// 在一个协程使用gls数据之前，需要先清掉可能存在的残余(见GetGidByCache函数说明)
func ClearGlsData() {
	glsData.ClearVal(GetCoid())
	return
}
