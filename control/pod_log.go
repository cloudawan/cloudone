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

func GetPodLog(kubeapiHost string, kubeapiPort int, namespace string, podName string) (returnedLog map[string]interface{}, returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("GetPodLog Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedLog = nil
			returnedError = err.(error)
		}
	}()

	result, err := restclient.RequestGet("http://"+kubeapiHost+":"+strconv.Itoa(kubeapiPort)+"/api/v1/namespaces/"+namespace+"/pods/"+podName, nil, false)
	if err != nil {
		log.Error("Fail to get pod information with namespace %s pod %s host: %s, port: %d, error: %s", namespace, podName, kubeapiHost, kubeapiPort, err.Error())
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
		byteSlice, err := restclient.RequestGetByteSliceResult("http://"+kubeapiHost+":"+strconv.Itoa(kubeapiPort)+"/api/v1/namespaces/"+namespace+"/pods/"+podName+"/log?container="+containerName, nil)
		if err != nil {
			log.Error("Fail to get log with namespace %s pod %s container %s host: %s, port: %d, error: %s", namespace, podName, containerName, kubeapiHost, kubeapiPort, err.Error())
			return nil, err
		} else {
			logJsonMap[containerName] = string(byteSlice)
		}
	}

	return logJsonMap, nil
}
