// Copyright 2022 Zhangtaoya. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package coroutine

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/zhangtaoya/coroutine/g"
)

const (
	ptrSize   = 4 << (^uintptr(0) >> 63) // unsafe.Sizeof(uintptr(0)) but an ideal const
	stackSize = 1024
)

var (
	goidOffset   uintptr
	goversion    string
	anchor       = []byte("goroutine ")
	stackBufPool = sync.Pool{
		New: func() interface{} {
			buf := make([]byte, 64)
			return &buf
		},
	}
	goidOffsetDict = map[string]int64{
		"go1.12": 152,
		"go1.13": 152,
		"go1.14": 152,
		"go1.15": 152,
		"go1.16": 152,
		"go1.17": 152,
	}
)

func init() {
	var off int64
	goversion = runtime.Version()
	for k, v := range goidOffsetDict {
		if goversion == k || strings.HasPrefix(goversion, k) {
			off = v
			break
		}
	}
	goidOffset = uintptr(off)
}

// getGoidByStack parse the current goroutine's id from caller stack.
// This function could be very slow(like 3000us/op), but it's very safe.
func getGoidByStack() (goid uint64) {
	bp := stackBufPool.Get().(*[]byte)
	defer stackBufPool.Put(bp)

	b := *bp
	b = b[:runtime.Stack(b, false)]
	goid, _ = findNextGoid(b, 0)
	return
}

// getGoidByNative parse the current goroutine's id from G.
// This function could be very fast(like 1ns/op), but it could be failed.
func getGoidByNative() (uint64, bool) {
	if goidOffset == 0 {
		return 0, false
	}
	tmp := g.G()
	if tmp == nil {
		return 0, false
	}
	p := (*uint64)(unsafe.Pointer(uintptr(tmp) + goidOffset))
	if p == nil {
		return 0, false
	}
	return *p, true
}

// getAllGoidByStack find all goid through stack; WARNING: This function could be very inefficient
func getAllGoidByStack() (goids []uint64) {
	count := runtime.NumGoroutine()
	size := count * stackSize // it's ok?
	buf := make([]byte, size)
	n := runtime.Stack(buf, true)
	buf = buf[:n]
	// parse all goids
	goids = make([]uint64, 0, count+4)
	for i := 0; i < len(buf); {
		goid, off := findNextGoid(buf, i)
		if goid > 0 {
			goids = append(goids, goid)
		}
		i = off
	}
	return
}

// Find the next goid from `buf[off:]`
func findNextGoid(buf []byte, off int) (goid uint64, next int) {
	i := off
	hit := false
	// skip to anchor
	acr := anchor
	for sb := len(buf) - len(acr); i < sb; {
		if buf[i] == acr[0] && buf[i+1] == acr[1] && buf[i+2] == acr[2] && buf[i+3] == acr[3] &&
			buf[i+4] == acr[4] && buf[i+5] == acr[5] && buf[i+6] == acr[6] &&
			buf[i+7] == acr[7] && buf[i+8] == acr[8] && buf[i+9] == acr[9] {
			hit = true
			i += len(acr)
			break
		}
		for ; i < len(buf) && buf[i] != '\n'; i++ {
		}
		i++
	}
	// return if not hit
	if !hit {
		return 0, len(buf)
	}
	// extract goid
	var done bool
	for ; i < len(buf) && !done; i++ {
		switch buf[i] {
		case '0':
			goid *= 10
		case '1':
			goid = goid*10 + 1
		case '2':
			goid = goid*10 + 2
		case '3':
			goid = goid*10 + 3
		case '4':
			goid = goid*10 + 4
		case '5':
			goid = goid*10 + 5
		case '6':
			goid = goid*10 + 6
		case '7':
			goid = goid*10 + 7
		case '8':
			goid = goid*10 + 8
		case '9':
			goid = goid*10 + 9
		case ' ':
			done = true
			break
		default:
			goid = 0
			fmt.Println("should never be here, any bug happens")
		}
	}
	next = i
	return
}

func getGidNoCache() uint64 {
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
func getGidByCache() uint64 {
	// 该方法内，不要调用别的方法，因为有可能别的方法很可能会在调用本方法(比如它打日志时需要本方法获取coid)
	// 只有容易出现死循环调用
	coidKey := "coid"
	addr := getRoutineAddr()
	if addr == 0 {
		return getGoidByStack()
	}

	if val, ok := addrGid.GetVal(addr, coidKey); ok && val != nil {
		gid := val.(uint64)
		return gid
	}

	if gid, ok := getGoidByNative(); ok {
		defer Info("get new gid by native succeed, addr:0x%x, gid:%d", addr, gid)
		addrGid.SetVal(addr, coidKey, gid)
		return gid
	}

	if gid := getGoidByStack(); gid >= 0 {
		defer Warn("get new gid by stack succeed, addr:0x%x, gid:%d", addr, gid)
		addrGid.SetVal(addr, coidKey, gid)
		return gid
	}

	gid := atomic.AddUint64(&gidFake, 1)
	defer Warn("get new gid use fake mod, addr:0x%x, gid:%d", addr, gid)
	addrGid.SetVal(addr, coidKey, gid)
	return gid
}

func getRoutineAddr() uint64 {
	// 这个函数取的只是当前协程的一个唯一标识
	// 用协程地址值当这个标识
	p := g.G()
	if p == nil {
		return 0
	}
	return uint64(uintptr(p))
	//x, _ := strconv.ParseUint(fmt.Sprintf("%v", p)[2:], 16, 64)
	//return x
}
