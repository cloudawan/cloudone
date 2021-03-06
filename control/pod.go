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
	"time"
)

type Pod struct {
	Name           string
	Namespace      string
	HostIP         string
	PodIP          string
	Phase          string
	Age            string
	ContainerSlice []PodContainer
}

type PodContainer struct {
	Name         string
	Image        string
	ContainerID  string
	RestartCount int
	Ready        bool
	PortSlice    []PodContainerPort
}

type PodContainerPort struct {
	Name          string
	ContainerPort int
	Protocol      string
}

func DeletePod(kubeApiServerEndPoint string, kubeApiServerToken string, namespace string, podName string) (returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("DeletePod Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedError = err.(error)
		}
	}()

	headerMap := make(map[string]string)
	headerMap["Authorization"] = kubeApiServerToken

	url := kubeApiServerEndPoint + "/api/v1/namespaces/" + namespace + "/pods/" + podName
	_, err := restclient.RequestDelete(url, nil, headerMap, true)
	if err != nil {
		log.Error(err)
		return err
	} else {
		return nil
	}
}

func GetAllPodNameBelongToReplicationController(kubeApiServerEndPoint string, kubeApiServerToken string, namespace string, replicationControllerName string) (returnedNameSlice []string, returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("GetAllPodNameBelongToReplicationController Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedNameSlice = nil
			returnedError = err.(error)
		}
	}()

	headerMap := make(map[string]string)
	headerMap["Authorization"] = kubeApiServerToken

	result, err := restclient.RequestGet(kubeApiServerEndPoint+"/api/v1/namespaces/"+namespace+"/pods/", headerMap, true)
	if err != nil {
		log.Error("Fail to get replication controller inofrmation with endpoint %s, token: %s, namespace: %s, replication controller name: %s, error %s", kubeApiServerEndPoint, kubeApiServerToken, namespace, replicationControllerName, err.Error())
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

func GetAllPodBelongToReplicationController(kubeApiServerEndPoint string, kubeApiServerToken string, namespace string, replicationControllerName string) (returnedPodSlice []Pod, returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("GetAllPodBelongToReplicationController Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedPodSlice = nil
			returnedError = err.(error)
		}
	}()

	headerMap := make(map[string]string)
	headerMap["Authorization"] = kubeApiServerToken

	result, err := restclient.RequestGet(kubeApiServerEndPoint+"/api/v1/namespaces/"+namespace+"/pods/", headerMap, true)
	jsonMap, _ := result.(map[string]interface{})
	if err != nil {
		log.Error("Fail to get all pod inofrmation with endpoint %s, token: %s, namespace: %s, replication controller name: %s, error %s", kubeApiServerEndPoint, kubeApiServerToken, namespace, replicationControllerName, err.Error())
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
						containerReady := false
						containerRestartCount := 0
						for _, containerStatus := range containerStatusSlice {
							name, _ := containerStatus.(map[string]interface{})["name"].(string)
							if name == containerName {
								containerId, _ = containerStatus.(map[string]interface{})["containerID"].(string)
								containerReady, _ = containerStatus.(map[string]interface{})["ready"].(bool)
								restartCount, _ := jsonparse.ConvertToInt64(containerStatus.(map[string]interface{})["restartCount"])
								containerRestartCount = int(restartCount)
							}
						}

						podContainer := PodContainer{
							containerName,
							containerImage,
							containerId,
							containerRestartCount,
							containerReady,
							portSlice,
						}

						containerSlice = append(containerSlice, podContainer)
					}
				}

				namespace, _ := data.(map[string]interface{})["metadata"].(map[string]interface{})["namespace"].(string)
				creationTimestamp, _ := data.(map[string]interface{})["metadata"].(map[string]interface{})["creationTimestamp"].(string)
				createdTime, err := time.Parse(time.RFC3339, creationTimestamp)
				age := ""
				if err == nil {
					age = GetTheFirstTimeUnit(time.Now().Sub(createdTime))
				}

				phase, _ := data.(map[string]interface{})["status"].(map[string]interface{})["phase"].(string)
				hostIP, _ := data.(map[string]interface{})["status"].(map[string]interface{})["hostIP"].(string)
				podIP, _ := data.(map[string]interface{})["status"].(map[string]interface{})["podIP"].(string)
				pod := Pod{
					nameField,
					namespace,
					hostIP,
					podIP,
					phase,
					age,
					containerSlice,
				}
				podSlice = append(podSlice, pod)
			}
		}
	}

	return podSlice, nil
}

func GetTheFirstTimeUnit(duration time.Duration) string {
	second := int(duration / time.Second)
	minute := int(duration / time.Minute)
	hour := int(duration / time.Hour)
	day := hour / 24

	if day > 0 {
		return strconv.Itoa(day) + "d"
	} else if hour > 0 {
		return strconv.Itoa(hour) + "h"
	} else if minute > 0 {
		return strconv.Itoa(minute) + "m"
	} else {
		return strconv.Itoa(second) + "s"
	}
}
