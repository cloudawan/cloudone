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

func GetPodLog(kubeApiServerEndPoint string, kubeApiServerToken string, namespace string, podName string) (returnedLog map[string]interface{}, returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("GetPodLog Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedLog = nil
			returnedError = err.(error)
		}
	}()

	headerMap := make(map[string]string)
	headerMap["Authorization"] = kubeApiServerToken

	result, err := restclient.RequestGet(kubeApiServerEndPoint+"/api/v1/namespaces/"+namespace+"/pods/"+podName, headerMap, false)
	if err != nil {
		log.Error("Fail to get pod information with namespace %s pod %s endpoint: %s, token: %s, error: %s", namespace, podName, kubeApiServerEndPoint, kubeApiServerToken, err.Error())
		return nil, err
	}

	containerNameSlice := make([]string, 0)

	jsonMap, _ := result.(map[string]interface{})
	specJsonMap, _ := jsonMap["spec"].(map[string]interface{})
	containerSlice, _ := specJsonMap["containers"].([]interface{})
	for _, container := range containerSlice {
		containerJsonMap, _ := container.(map[string]interface{})
		containerName, ok := containerJsonMap["name"].(string)
		if ok {
			containerNameSlice = append(containerNameSlice, containerName)
		}
	}

	logJsonMap := make(map[string]interface{})
	for _, containerName := range containerNameSlice {
		byteSlice, err := restclient.RequestGetByteSliceResult(kubeApiServerEndPoint+"/api/v1/namespaces/"+namespace+"/pods/"+podName+"/log?container="+containerName, headerMap)
		if err != nil {
			log.Error("Fail to get log with namespace %s pod %s container %s endpoint: %s, token: %s, error: %s", namespace, podName, containerName, kubeApiServerEndPoint, kubeApiServerToken, err.Error())
			return nil, err
		} else {
			logJsonMap[containerName] = string(byteSlice)
		}
	}

	return logJsonMap, nil
}
