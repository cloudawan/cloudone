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

package monitor

import (
	"errors"
	"github.com/cloudawan/cloudone/control"
	"github.com/cloudawan/cloudone_utility/logger"
	"github.com/cloudawan/cloudone_utility/restclient"
	"strconv"
)

type ReplicationControllerMetric struct {
	Namespace                 string
	ReplicationControllerName string
	ValidPodSlice             []bool
	PodMetricSlice            []PodMetric
	Size                      int
}

func ExistReplicationController(kubeapiHost string, kubeapiPort int, namespace string, replicationControllerName string) (returnedExist bool, returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("ExistReplicationController Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedExist = false
			returnedError = err.(error)
		}
	}()

	_, err := restclient.RequestGet("http://"+kubeapiHost+":"+strconv.Itoa(kubeapiPort)+"/api/v1/namespaces/"+namespace+"/replicationcontrollers/"+replicationControllerName, true)
	if err != nil {
		log.Error("Fail to detect replication controller existence with host %s, port: %d, namespace: %s, replication controller name: %s, error %s", kubeapiHost, kubeapiPort, namespace, replicationControllerName, err.Error())
		return false, err
	} else {
		return true, nil
	}
}

func MonitorReplicationController(kubeapiHost string, kubeapiPort int, namespace string, replicationControllerName string) (returnedReplicationControllerMetric *ReplicationControllerMetric, returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("MonitorReplicationController Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedReplicationControllerMetric = nil
			returnedError = err.(error)
		}
	}()

	exist, err := ExistReplicationController(kubeapiHost, kubeapiPort, namespace, replicationControllerName)
	if err != nil {
		log.Error("Fail to get the replication controller with host %s, port: %d, namespace: %s, replication controller name: %s", kubeapiHost, kubeapiPort, namespace, replicationControllerName)
		return nil, err
	}
	if exist == false {
		log.Error("Replication controller doesn't exist with host %s, port: %d, namespace: %s, replication controller name: %s", kubeapiHost, kubeapiPort, namespace, replicationControllerName)
		return nil, err
	}

	podNameSlice, err := control.GetAllPodNameBelongToReplicationController(kubeapiHost, kubeapiPort, namespace, replicationControllerName)
	if err != nil {
		log.Error("Fail to get all pod name belong to the replication controller with host %s, port: %d, namespace: %s, replication controller name: %s", kubeapiHost, kubeapiPort, namespace, replicationControllerName)
		return nil, err
	}

	replicationControllerMetric := &ReplicationControllerMetric{}
	replicationControllerMetric.Namespace = namespace
	replicationControllerMetric.ReplicationControllerName = replicationControllerName
	replicationControllerMetric.Size = len(podNameSlice)
	replicationControllerMetric.ValidPodSlice = make([]bool, replicationControllerMetric.Size)
	replicationControllerMetric.PodMetricSlice = make([]PodMetric, replicationControllerMetric.Size)
	errorMessage := "The following index of pod has error: "
	errorHappened := false
	for index, podName := range podNameSlice {
		podMetric, err := MonitorPod(kubeapiHost, kubeapiPort, namespace, podName)
		if err != nil {
			errorMessage = errorMessage + err.Error()
			errorHappened = true
			replicationControllerMetric.ValidPodSlice[index] = false
		} else {
			replicationControllerMetric.ValidPodSlice[index] = true
			replicationControllerMetric.PodMetricSlice[index] = *podMetric
		}
	}

	if errorHappened {
		log.Error("Fail to get all pod inofrmation with host %s, port: %d, namespace: %s, replication controller name: %s, error %s", kubeapiHost, kubeapiPort, namespace, replicationControllerName, errorMessage)
		return replicationControllerMetric, errors.New(errorMessage)
	} else {
		return replicationControllerMetric, nil
	}
}

func GetReplicationControllerNameFromSelector(kubeapiHost string, kubeapiPort int, namespace string, targetSelectorName string) (returnedReplicationControllerNameSlice []string, returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("GetReplicationControllerNameFromSelector Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedReplicationControllerNameSlice = nil
			returnedError = err.(error)
		}
	}()

	result, err := restclient.RequestGet("http://"+kubeapiHost+":"+strconv.Itoa(kubeapiPort)+
		"/api/v1/namespaces/"+namespace+"/replicationcontrollers/", true)
	jsonMap, _ := result.(map[string]interface{})
	if err != nil {
		log.Error("Fail to get all replication controller with host %s, port: %d, namespace: %s, selector name: %s",
			kubeapiHost, kubeapiPort, namespace, targetSelectorName)
		return nil, err
	}

	nameSlice := make([]string, 0)
	for _, item := range jsonMap["items"].([]interface{}) {
		selector, ok := item.(map[string]interface{})["spec"].(map[string]interface{})["selector"].(map[string]interface{})
		if ok {
			selectorName, ok := selector["name"].(string)
			if ok {
				if targetSelectorName == selectorName {
					name, ok := item.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
					if ok {
						nameSlice = append(nameSlice, name)
					}
				}
			}
		}
	}

	return nameSlice, nil
}
