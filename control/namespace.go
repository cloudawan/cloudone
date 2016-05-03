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

package control

import (
	"github.com/cloudawan/cloudone_utility/logger"
	"github.com/cloudawan/cloudone_utility/restclient"
	"strconv"
)

func GetAllNamespaceName(kubeapiHost string, kubeapiPort int) (returnedNameSlice []string, returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("GetAllNamespaceName Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedNameSlice = nil
			returnedError = err.(error)
		}
	}()

	result, err := restclient.RequestGet("http://"+kubeapiHost+":"+strconv.Itoa(kubeapiPort)+"/api/v1/namespaces/", nil, true)
	if err != nil {
		log.Error("Fail to get all namespace name with host: %s, port: %d, error: %s", kubeapiHost, kubeapiPort, err.Error())
		return nil, err
	}
	jsonMap, _ := result.(map[string]interface{})

	nameSlice := make([]string, 0)
	for _, data := range jsonMap["items"].([]interface{}) {
		name, ok := data.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
		if ok {
			statusJsonMap, _ := data.(map[string]interface{})["status"].(map[string]interface{})
			phase, _ := statusJsonMap["phase"].(string)
			if phase == "Active" {
				nameSlice = append(nameSlice, name)
			}
		}

	}

	return nameSlice, nil
}

func CreateNamespace(kubeapiHost string, kubeapiPort int, name string) (returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("CreateNamespace Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedError = err.(error)
		}
	}()

	jsonMap := make(map[string]interface{})
	jsonMap["metadata"] = make(map[string]interface{})
	jsonMap["metadata"].(map[string]interface{})["name"] = name

	_, err := restclient.RequestPost("http://"+kubeapiHost+":"+strconv.Itoa(kubeapiPort)+"/api/v1/namespaces/", jsonMap, nil, true)
	if err != nil {
		log.Error("Fail to create namespace with host: %s, port: %d, name: %s, error: %s", kubeapiHost, kubeapiPort, name, err.Error())
		return err
	}

	return nil
}

func DeleteNamespace(kubeapiHost string, kubeapiPort int, name string) (returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("DeleteNamespace Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedError = err.(error)
		}
	}()

	_, err := restclient.RequestDelete("http://"+kubeapiHost+":"+strconv.Itoa(kubeapiPort)+"/api/v1/namespaces/"+name, nil, nil, true)
	if err != nil {
		log.Error("Fail to delete namespace with host: %s, port: %d, name: %s, error: %s", kubeapiHost, kubeapiPort, name, err.Error())
		return err
	}

	return nil
}
