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

func GetPodLog(kubeapiHost string, kubeapiPort int, namespace string, podName string) (returnedLog string, returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("GetPodLog Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedLog = ""
			returnedError = err.(error)
		}
	}()

	byteSlice, err := restclient.RequestGetByteSliceResult("http://" + kubeapiHost + ":" + strconv.Itoa(kubeapiPort) + "/api/v1/namespaces/" + namespace + "/pods/" + podName + "/log")
	if err != nil {
		log.Error("Fail to get all namespace name with host: %s, port: %d, error: %s", kubeapiHost, kubeapiPort, err.Error())
		return "", err
	}

	return string(byteSlice), nil
}
