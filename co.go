// Copyright 2022 Zhangtaoya. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package coroutine

import (
	"sync"
)

type Co struct {
	wg sync.WaitGroup
}

func (g *Co) Start(f func()) {
	glsDataCpy, _ := GetGlsDataCpy()

	// 当函数调用到这个地方后，即将启新的协程，调用堆栈信息在这里会断掉
	// 这里先将该协程的调用堆栈信息取出来传给子协程
	// 以此实现跨协程调用堆栈跟踪
	_, traceStr := GetCurCoTraceStr()

	g.wg.Add(1)
	go func() {
		defer g.wg.Done()

		// 将父协程的gls数据拷贝过来，实现跨协程gls数据传递
		SetGlsData(glsDataCpy)

		// 这里已经是新协程了，将父协程的trace信息存起来
		// 后面可以实现跨协程调用跟踪
		if traceStr != "" {
			SetCurCoParentTraceStr(traceStr)
		}

		f()
	}()
}

func (g *Co) StartWithParam(f func(...interface{}), param ...interface{}) {
	glsDataCpy, _ := GetGlsDataCpy()
	_, traceStr := GetCurCoTraceStr()
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		SetGlsData(glsDataCpy)
		if traceStr != "" {
			SetCurCoParentTraceStr(traceStr)
		}

		f(param...)
	}()
}

func (g *Co) Wait() {
	g.wg.Wait()
}

func Go(f func()) { //  just go, no wait
	glsDataCpy, _ := GetGlsDataCpy()
	_, traceStr := GetCurCoTraceStr()
	go func() {
		SetGlsData(glsDataCpy)
		if traceStr != "" {
			SetCurCoParentTraceStr(traceStr)
		}
		f()
	}()
}

func GoWithParam(f func(...interface{}), param ...interface{}) { //  just go, no wait
	glsDataCpy, _ := GetGlsDataCpy()
	_, traceStr := GetCurCoTraceStr()
	go func() {
		SetGlsData(glsDataCpy)
		if traceStr != "" {
			SetCurCoParentTraceStr(traceStr)
		}
		f(param...)
	}()
}
