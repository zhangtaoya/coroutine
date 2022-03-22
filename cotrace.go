// Copyright 2022 Zhangtaoya. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package coroutine

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

const GlsCoParentCallStackStrKey = "GlsCoParentCallStackStrKey"

func GetCurCoTraceStr() (coid uint64, traceStr string) {
	coid, traceParent, _ := GetVal(GlsCoParentCallStackStrKey)
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

func SetCurCoParentTraceStr(traceStr string) {
	SetVal(GlsCoParentCallStackStrKey, traceStr)
}

var whiteFullPathList, blackNameList []string
var replaceMap map[string]string

func SetCallStackStrFilter(whiteFullPathListIn, blackNameListIn []string, replaceMapIn map[string]string) {
	whiteFullPathList = whiteFullPathListIn
	blackNameList = blackNameListIn
	replaceMap = replaceMapIn
}
func GetCallStack(Max int) string {
	return GetCallStackStrWithBaseLvl(Max, 3, true)
}

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
