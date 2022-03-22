// Copyright 2022 Zhangtaoya. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package logger

var fInfo, fWarn, fError func(format string, args ...interface{}) error

func SetLogger(fInfoi, fWarni, fErrori func(format string, args ...interface{}) error) {
	fInfo = fInfoi
	fWarn = fWarni
	fError = fErrori
}

func Info(format string, args ...interface{}) error {
	if fInfo != nil {
		return fInfo(format, args...)
	}
	println(format, args)
	return nil
}

func Warn(format string, args ...interface{}) error {
	if fWarn != nil {
		return fWarn(format, args...)
	}
	println(format, args)
	return nil
}

func Error(format string, args ...interface{}) error {
	if fError != nil {
		return fError(format, args...)
	}
	println(format, args)
	return nil
}
