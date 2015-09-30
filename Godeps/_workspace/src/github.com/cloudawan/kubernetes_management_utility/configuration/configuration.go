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

package configuration

import (
	"encoding/json"
	"github.com/cloudawan/kubernetes_management/Godeps/_workspace/src/code.google.com/p/log4go"
	"github.com/cloudawan/kubernetes_management/Godeps/_workspace/src/github.com/cloudawan/kubernetes_management_utility/logger"
	"io"
	"os"
)

var log log4go.Logger = logger.GetLogger("utility")

const (
	rootPath = "/etc"
	fileName = "configuration.json"
)

type Configuration struct {
	jsonMap         map[string]interface{}
	fileNameAndPath string
}

func CreateConfiguration(programName string, configurationContent string) (*Configuration, error) {
	configuration := new(Configuration)
	if err := configuration.createConfigurationFileIfNotExist(programName, configurationContent); err != nil {
		return nil, err
	}
	if err := configuration.readConfigurationFromFile(); err != nil {
		return nil, err
	}
	return configuration, nil
}

func (configuration *Configuration) GetNative(key string) interface{} {
	return configuration.jsonMap[key]
}

func (configuration *Configuration) GetString(key string) (string, bool) {
	value, ok := configuration.jsonMap[key].(string)
	return value, ok
}

func (configuration *Configuration) GetStringSlice(key string) ([]string, bool) {
	stringSlice := make([]string, 0)
	valueSlice, ok := configuration.jsonMap[key].([]interface{})
	if ok == false {
		return nil, false
	}
	for _, value := range valueSlice {
		text, ok := value.(string)
		if ok == false {
			return nil, false
		}
		stringSlice = append(stringSlice, text)
	}
	return stringSlice, true
}

func (configuration *Configuration) GetInt(key string) (int, bool) {
	number, ok := configuration.jsonMap[key].(float64)
	if ok {
		return int(number), true
	} else {
		return 0, false
	}
}

func (configuration *Configuration) GetFloat64(key string) (float64, bool) {
	value, ok := configuration.jsonMap[key].(float64)
	return value, ok
}

func (configuration *Configuration) createConfigurationFileIfNotExist(programName string, configurationContent string) (returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("configuration Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedError = err.(error)
		}
	}()

	directoryPath := rootPath + string(os.PathSeparator) + programName
	if err := os.MkdirAll(directoryPath, os.ModePerm); err != nil {
		return err
	}

	fileNameAndPath := directoryPath + string(os.PathSeparator) + fileName

	if _, err := os.Stat(fileNameAndPath); os.IsNotExist(err) {
		outputFile, err := os.Create(fileNameAndPath)
		if err != nil {
			log.Error("createConfigurationFileIfNotExist create file %s with error: %s", fileNameAndPath, err)
			return err
		}
		outputFile.WriteString(configurationContent)
	}

	configuration.fileNameAndPath = fileNameAndPath

	return nil
}

func (configuration *Configuration) Reload() error {
	return configuration.readConfigurationFromFile()
}

func (configuration *Configuration) readConfigurationFromFile() (returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("readConfigurationFromFile Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedError = err.(error)
		}
	}()

	// open input file
	inputFile, err := os.Open(configuration.fileNameAndPath)
	if err != nil {
		log.Error("readConfigurationFromFile open file error: %s", err)
		return err
	}
	defer inputFile.Close()

	byteSlice := make([]byte, 0)
	buffer := make([]byte, 1024)
	for {
		// read a chunk
		n, err := inputFile.Read(buffer)
		if err != nil && err != io.EOF {
			log.Error("readConfigurationFromFile read file error: %s", err)
			return err
		}
		if n == 0 {
			break
		}

		byteSlice = append(byteSlice, buffer[0:n]...)
	}

	jsonMap := make(map[string]interface{})
	err = json.Unmarshal(byteSlice, &jsonMap)
	if err != nil {
		log.Error("readConfigurationFromFile parse file error: %s", err)
		return err
	}

	configuration.jsonMap = jsonMap

	return nil
}
