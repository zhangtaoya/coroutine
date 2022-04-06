// Copyright 2022 Zhangtaoya. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package coroutine

import (
	"code.google.com/p/go-uuid/uuid"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

var glsCoParentCallStackStrKey = "CSKEY"

func init() {
	glsCoParentCallStackStrKey += uuid.New()
}

// 获取当前协程级联的调用堆栈，如果用该包里的方法启协程，这个堆栈会追踪到所有的父协程
// 一个协程里面，追踪5级调用
func GetCurCoTraceStr() (coid uint64, traceStr string) {
	coid, traceParent, _ := GetVal(glsCoParentCallStackStrKey)
	traceParentStr := ""
	if traceParent != nil {
		traceParentStr = traceParent.(string)
	}

	traceFile := GetCallStackStrWithBaseLvl(5, 5, false)
	traceFileWithParent := fmt.Sprintf("%s-coid:%d-%s", traceParentStr, coid, traceFile)
	traceFileWithParent = strings.ReplaceAll(traceFileWithParent, "-----", "--")
	traceFileWithParent = strings.ReplaceAll(traceFileWithParent, "----", "--")
	traceFileWithParent = strings.ReplaceAll(traceFileWithParent, "---", "--")
	return coid, traceFileWithParent
}

// 设置调用堆栈的黑白名单，以及替换名单，可选
func SetCallStackStrFilter(whiteFullPathListIn, blackNameListIn []string, replaceMapIn map[string]string) {
	whiteFullPathList = whiteFullPathListIn
	blackNameList = blackNameListIn
	replaceMap = replaceMapIn
}

// 获取直接的调用堆栈，无协程级联
func GetCallStack(Max int) string {
	return GetCallStackStrWithBaseLvl(Max, 3, true)
}

// 获取直接的调用堆栈，无协程级联
func GetCallStackStrWithBaseLvl(Max, baselvl int, allFullPathOpen bool) string {
	callStackStr := ""
	fileNameMap := map[string]bool{}
	for i := Max + baselvl - 1; i >= baselvl; i-- {
		_, fPath, l, _ := runtime.Caller(i)
		if fPath == "" {
			continue
		}
		fName := fPath

		if allFullPathOpen == false {
			fName = strings.ReplaceAll(filepath.Base(fPath), ".go", "")

			// 黑名单文件名不展示，主要过滤框架类文件名
			isBlack := false
			for _, blackName := range blackNameList {
				blackName = strings.ReplaceAll(blackName, ".go", "")
				if fName == blackName {
					isBlack = true
					break
				}
			}
			if isBlack {
				callStackStr += "-"
				continue
			}

			for _, matchStart := range whiteFullPathList {
				if idx := strings.Index(fPath, matchStart); idx > 0 {
					fName = fPath[idx:]
					break
				}
			}

			// 处理重复的带路径文件名
			// 如果都是重复路径，只给文件名就行了
			if _, ok := fileNameMap[fName]; ok {
				fName = strings.ReplaceAll(filepath.Base(fPath), ".go", "")
			}

			// 长名字替换
			for k, v := range replaceMap {
				fName = strings.ReplaceAll(fName, k, v)
			}

			fileNameMap[fName] = true
		}

		if len(callStackStr) == 0 {
			callStackStr = fmt.Sprintf("%s:%d", fName, l)
		} else {
			callStackStr += fmt.Sprintf("-%s:%d", fName, l)
		}
	}
	return callStackStr
}

func setCurCoParentTraceStr(traceStr string) {
	SetVal(glsCoParentCallStackStrKey, traceStr)
}

var whiteFullPathList, blackNameList []string
var replaceMap map[string]string
