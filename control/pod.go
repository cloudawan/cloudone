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
	"github.com/cloudawan/cloudone_utility/jsonparse"
	"github.com/cloudawan/cloudone_utility/logger"
	"github.com/cloudawan/cloudone_utility/restclient"
	"strconv"
)

type Pod struct {
	Name           string
	Namespace      string
	HostIP         string
	PodIP          string
	ContainerSlice []PodContainer
}

type PodContainer struct {
	Name        string
	Image       string
	ContainerID string
	PortSlice   []PodContainerPort
}

type PodContainerPort struct {
	Name          string
	ContainerPort int
	Protocol      string
}

func DeletePod(kubeapiHost string, kubeapiPort int, namespace string, podName string) (returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("DeletePod Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedError = err.(error)
		}
	}()

	url := "http://" + kubeapiHost + ":" + strconv.Itoa(kubeapiPort) + "/api/v1/namespaces/" + namespace + "/pods/" + podName
	_, err := restclient.RequestDelete(url, nil, nil, true)
	if err != nil {
		return err
	} else {
		return nil
	}
}

func GetAllPodNameBelongToReplicationController(kubeapiHost string, kubeapiPort int, namespace string, replicationControllerName string) (returnedNameSlice []string, returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("GetAllPodNameBelongToReplicationController Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedNameSlice = nil
			returnedError = err.(error)
		}
	}()

	result, err := restclient.RequestGet("http://"+kubeapiHost+":"+strconv.Itoa(kubeapiPort)+"/api/v1/namespaces/"+namespace+"/pods/", nil, true)
	if err != nil {
		log.Error("Fail to get replication controller inofrmation with host %s, port: %d, namespace: %s, replication controller name: %s, error %s", kubeapiHost, kubeapiPort, namespace, replicationControllerName, err.Error())
		return nil, err
	}
	jsonMap, _ := result.(map[string]interface{})

	generateName := replicationControllerName + "-"
	podNameSlice := make([]string, 0)
	for _, data := range jsonMap["items"].([]interface{}) {
		generateNameField, _ := data.(map[string]interface{})["metadata"].(map[string]interface{})["generateName"].(string)
		nameField, nameFieldOk := data.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
		if generateName == generateNameField {
			if nameFieldOk {
				podNameSlice = append(podNameSlice, nameField)
			}
		} else if replicationControllerName == nameField {
			if nameFieldOk {
				podNameSlice = append(podNameSlice, nameField)
			}
		}
	}

	return podNameSlice, nil
}

func GetAllPodBelongToReplicationController(kubeapiHost string, kubeapiPort int, namespace string, replicationControllerName string) (returnedPodSlice []Pod, returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("GetAllPodBelongToReplicationController Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedPodSlice = nil
			returnedError = err.(error)
		}
	}()

	result, err := restclient.RequestGet("http://"+kubeapiHost+":"+strconv.Itoa(kubeapiPort)+"/api/v1/namespaces/"+namespace+"/pods/", nil, true)
	jsonMap, _ := result.(map[string]interface{})
	if err != nil {
		log.Error("Fail to get all pod inofrmation with host %s, port: %d, namespace: %s, replication controller name: %s, error %s", kubeapiHost, kubeapiPort, namespace, replicationControllerName, err.Error())
		return nil, err
	}

	return GetAllPodBelongToReplicationControllerFromData(replicationControllerName, jsonMap)
}

func GetAllPodBelongToReplicationControllerFromData(replicationControllerName string, jsonMap map[string]interface{}) (returnedPodSlice []Pod, returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("GetAllPodBelongToReplicationControllerFromData Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedPodSlice = nil
			returnedError = err.(error)
		}
	}()

	generateName := replicationControllerName + "-"
	podSlice := make([]Pod, 0)
	for _, data := range jsonMap["items"].([]interface{}) {
		generateNameField, _ := data.(map[string]interface{})["metadata"].(map[string]interface{})["generateName"].(string)
		nameField, nameFieldOk := data.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
		if generateName == generateNameField || replicationControllerName == nameField {
			if nameFieldOk {
				containerSlice := make([]PodContainer, 0)
				containerFieldSlice, containerFieldSliceOk := data.(map[string]interface{})["spec"].(map[string]interface{})["containers"].([]interface{})
				containerStatusSlice, containerStatusSliceOk := data.(map[string]interface{})["status"].(map[string]interface{})["containerStatuses"].([]interface{})
				if containerFieldSliceOk && containerStatusSliceOk {
					for _, containerField := range containerFieldSlice {
						portSlice := make([]PodContainerPort, 0)
						portFieldSlice, portFieldSliceOk := containerField.(map[string]interface{})["ports"].([]interface{})
						if portFieldSliceOk {
							for _, portField := range portFieldSlice {
								portName, _ := portField.(map[string]interface{})["name"].(string)
								portContainerPort, _ := jsonparse.ConvertToInt64(portField.(map[string]interface{})["containerPort"])
								portProtocol, _ := portField.(map[string]interface{})["protocol"].(string)

								podContainerPort := PodContainerPort{
									portName,
									int(portContainerPort),
									portProtocol,
								}

								portSlice = append(portSlice, podContainerPort)
							}
						}

						containerName, _ := containerField.(map[string]interface{})["name"].(string)
						containerImage, _ := containerField.(map[string]interface{})["image"].(string)

						containerId := ""
						for _, containerStatus := range containerStatusSlice {
							name, _ := containerStatus.(map[string]interface{})["name"].(string)
							if name == containerName {
								containerId, _ = containerStatus.(map[string]interface{})["containerID"].(string)
							}
						}

						podContainer := PodContainer{
							containerName,
							containerImage,
							containerId,
							portSlice,
						}

						containerSlice = append(containerSlice, podContainer)
					}
				}

				namespace, _ := data.(map[string]interface{})["metadata"].(map[string]interface{})["namespace"].(string)
				hostIP, _ := data.(map[string]interface{})["status"].(map[string]interface{})["hostIP"].(string)
				podIP, _ := data.(map[string]interface{})["status"].(map[string]interface{})["podIP"].(string)
				pod := Pod{
					nameField,
					namespace,
					hostIP,
					podIP,
					containerSlice,
				}
				podSlice = append(podSlice, pod)
			}
		}
	}

	return podSlice, nil
}
