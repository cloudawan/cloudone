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
)

func GetAllNamespaceName(kubeApiServerEndPoint string, kubeApiServerToken string) (returnedNameSlice []string, returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("GetAllNamespaceName Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedNameSlice = nil
			returnedError = err.(error)
		}
	}()

	headerMap := make(map[string]string)
	headerMap["Authorization"] = kubeApiServerToken

	result, err := restclient.RequestGet(kubeApiServerEndPoint+"/api/v1/namespaces/", headerMap, true)
	if err != nil {
		log.Error("Fail to get all namespace name with endpoint: %s, token: %s, error: %s", kubeApiServerEndPoint, kubeApiServerToken, err.Error())
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

func CreateNamespace(kubeApiServerEndPoint string, kubeApiServerToken string, name string) (returnedError error) {
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

	headerMap := make(map[string]string)
	headerMap["Authorization"] = kubeApiServerToken

	_, err := restclient.RequestPost(kubeApiServerEndPoint+"/api/v1/namespaces/", jsonMap, headerMap, true)
	if err != nil {
		log.Error("Fail to create namespace with endpoint: %s, token: %s, name: %s, error: %s", kubeApiServerEndPoint, kubeApiServerToken, name, err.Error())
		return err
	}

	return nil
}

func DeleteNamespace(kubeApiServerEndPoint string, kubeApiServerToken string, name string) (returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("DeleteNamespace Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedError = err.(error)
		}
	}()

	headerMap := make(map[string]string)
	headerMap["Authorization"] = kubeApiServerToken

	_, err := restclient.RequestDelete(kubeApiServerEndPoint+"/api/v1/namespaces/"+name, nil, headerMap, true)
	if err != nil {
		log.Error("Fail to delete namespace with endpoint: %s, token: %s, name: %s, error: %s", kubeApiServerEndPoint, kubeApiServerToken, name, err.Error())
		return err
	}

	return nil
}
