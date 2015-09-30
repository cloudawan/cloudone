// Copyright 2015 CloudAwan LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logger

import (
	"github.com/cloudawan/kubernetes_management/Godeps/_workspace/src/code.google.com/p/log4go"
	"os"
	"runtime"
)

const (
	rootPath  = "/var/log"
	logSuffix = ".log"
)

type Log struct {
	loggerMap     map[string]log4go.Logger
	directoryPath string
}

func CreateLog(programName string) (*Log, error) {
	directoryPath, err := createDirectoryIfNotExist(programName)
	if err != nil {
		return nil, err
	}
	return &Log{make(map[string]log4go.Logger), directoryPath}, nil
}

func createDirectoryIfNotExist(programName string) (string, error) {
	directoryPath := rootPath + string(os.PathSeparator) + programName
	return directoryPath, os.MkdirAll(directoryPath, os.ModePerm)
}

func (log *Log) GetLogger(moduleName string) log4go.Logger {
	filePath := log.directoryPath + string(os.PathSeparator) + moduleName + logSuffix
	return log.getLoggerWithFileName(filePath)
}

func (log *Log) getLoggerWithFileName(filePath string) log4go.Logger {
	logger, ok := log.loggerMap[filePath]
	if ok {
		return logger
	} else {
		// Create the empty logger
		logger = make(log4go.Logger)
		fileWriter := log4go.NewFileLogWriter(filePath, false)
		fileWriter.SetFormat("[%D %T] [%L] (%S) %M")
		fileWriter.SetRotate(false)
		fileWriter.SetRotateSize(100 * 1024 * 1024)
		fileWriter.SetRotateLines(0)
		fileWriter.SetRotateDaily(false)
		logger.AddFilter("file", log4go.DEBUG, fileWriter)

		log.loggerMap[filePath] = logger
		return logger
	}
}

// Self logger
var log *Log

func init() {
	var err error
	log, err = CreateLog("kubernetes_management_utility")
	if err != nil {
		panic(err)
	}
}

func GetLogger(moduleName string) log4go.Logger {
	return log.GetLogger(moduleName)
}

func GetStackTrace(maxByteAmount int, allRoutines bool) string {
	trace := make([]byte, maxByteAmount)
	count := runtime.Stack(trace, allRoutines)
	return string(trace[:count])
}
