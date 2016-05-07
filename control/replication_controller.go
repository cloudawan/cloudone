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
	"github.com/cloudawan/cloudone_utility/deepcopy"
	"github.com/cloudawan/cloudone_utility/jsonparse"
	"github.com/cloudawan/cloudone_utility/logger"
	"github.com/cloudawan/cloudone_utility/restclient"
	"strconv"
	"time"
)

type ReplicationController struct {
	Name           string
	ReplicaAmount  int
	Selector       ReplicationControllerSelector
	Label          ReplicationControllerLabel
	ContainerSlice []ReplicationControllerContainer
	ExtraJsonMap   map[string]interface{}
}

type ReplicationControllerSelector struct {
	Name    string
	Version string
}

type ReplicationControllerLabel struct {
	Name string
}

type ReplicationControllerContainer struct {
	Name             string
	Image            string
	PortSlice        []ReplicationControllerContainerPort
	EnvironmentSlice []ReplicationControllerContainerEnvironment
	ResourceMap      map[string]interface{}
}

type ReplicationControllerContainerPort struct {
	Name          string
	ContainerPort int
}

type ReplicationControllerContainerEnvironment struct {
	Name  string
	Value string
}

func CreateReplicationController(kubeapiHost string, kubeapiPort int, namespace string, replicationController ReplicationController) (returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("CreateReplicationController Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedError = err.(error)
		}
	}()

	containerJsonMapSlice := make([]interface{}, 0)
	for _, replicationControllerContainer := range replicationController.ContainerSlice {
		containerJsonMap := make(map[string]interface{})
		containerJsonMap["name"] = replicationControllerContainer.Name
		containerJsonMap["image"] = replicationControllerContainer.Image
		containerJsonMap["resources"] = replicationControllerContainer.ResourceMap

		portJsonMapSlice := make([]interface{}, 0)
		for _, replicationControllerContainerPort := range replicationControllerContainer.PortSlice {
			portJsonMap := make(map[string]interface{})
			portJsonMap["name"] = replicationControllerContainerPort.Name
			portJsonMap["containerPort"] = replicationControllerContainerPort.ContainerPort
			portJsonMapSlice = append(portJsonMapSlice, portJsonMap)
		}
		containerJsonMap["ports"] = portJsonMapSlice

		environmentJsonMapSlice := make([]interface{}, 0)
		for _, environment := range replicationControllerContainer.EnvironmentSlice {
			environmentJsonMap := make(map[string]interface{})
			environmentJsonMap["name"] = environment.Name
			environmentJsonMap["value"] = environment.Value
			environmentJsonMapSlice = append(environmentJsonMapSlice, environmentJsonMap)
		}
		containerJsonMap["env"] = environmentJsonMapSlice

		// FIXME temporarily to use nested docker
		volumeMountJsonMap := make(map[string]interface{})
		volumeMountJsonMap["name"] = "docker"
		volumeMountJsonMap["readOnly"] = true
		volumeMountJsonMap["mountPath"] = "/var/run/docker.sock"
		volumeMountJsonMapSlice := make([]interface{}, 0)
		volumeMountJsonMapSlice = append(volumeMountJsonMapSlice, volumeMountJsonMap)
		containerJsonMap["volumeMounts"] = volumeMountJsonMapSlice

		containerJsonMapSlice = append(containerJsonMapSlice, containerJsonMap)
	}

	bodyJsonMap := make(map[string]interface{})
	bodyJsonMap["kind"] = "ReplicationController"
	bodyJsonMap["apiVersion"] = "v1"
	bodyJsonMap["metadata"] = make(map[string]interface{})
	bodyJsonMap["metadata"].(map[string]interface{})["name"] = replicationController.Name
	bodyJsonMap["metadata"].(map[string]interface{})["labels"] = make(map[string]interface{})
	bodyJsonMap["metadata"].(map[string]interface{})["labels"].(map[string]interface{})["name"] = replicationController.Label.Name
	bodyJsonMap["spec"] = make(map[string]interface{})
	bodyJsonMap["spec"].(map[string]interface{})["replicas"] = replicationController.ReplicaAmount
	bodyJsonMap["spec"].(map[string]interface{})["selector"] = make(map[string]interface{})
	bodyJsonMap["spec"].(map[string]interface{})["selector"].(map[string]interface{})["name"] = replicationController.Selector.Name
	bodyJsonMap["spec"].(map[string]interface{})["selector"].(map[string]interface{})["version"] = replicationController.Selector.Version
	bodyJsonMap["spec"].(map[string]interface{})["template"] = make(map[string]interface{})
	bodyJsonMap["spec"].(map[string]interface{})["template"].(map[string]interface{})["metadata"] = make(map[string]interface{})
	bodyJsonMap["spec"].(map[string]interface{})["template"].(map[string]interface{})["metadata"].(map[string]interface{})["labels"] = make(map[string]interface{})
	bodyJsonMap["spec"].(map[string]interface{})["template"].(map[string]interface{})["metadata"].(map[string]interface{})["labels"].(map[string]interface{})["name"] = replicationController.Selector.Name
	bodyJsonMap["spec"].(map[string]interface{})["template"].(map[string]interface{})["metadata"].(map[string]interface{})["labels"].(map[string]interface{})["version"] = replicationController.Selector.Version
	bodyJsonMap["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"] = make(map[string]interface{})
	bodyJsonMap["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{})["containers"] = containerJsonMapSlice

	// FIXME temporarily to use nested docker
	volumeJsonMap := make(map[string]interface{})
	volumeJsonMap["name"] = "docker"
	volumeJsonMap["hostPath"] = make(map[string]interface{})
	volumeJsonMap["hostPath"].(map[string]interface{})["path"] = "/var/run/docker.sock"
	volumeJsonMapSlice := make([]interface{}, 0)
	volumeJsonMapSlice = append(volumeJsonMapSlice, volumeJsonMap)
	bodyJsonMap["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{})["volumes"] = volumeJsonMapSlice

	// Configure extra json body
	// It is used for user to input any configuration
	if replicationController.ExtraJsonMap != nil {
		deepcopy.DeepOverwriteJsonMap(replicationController.ExtraJsonMap, bodyJsonMap)
	}

	url := "http://" + kubeapiHost + ":" + strconv.Itoa(kubeapiPort) + "/api/v1/namespaces/" + namespace + "/replicationcontrollers/"
	_, err := restclient.RequestPost(url, bodyJsonMap, nil, true)

	if err != nil {
		return err
	} else {
		return nil
	}
}

func DeleteReplicationController(kubeapiHost string, kubeapiPort int, namespace string, replicationControllerName string) (returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("DeleteReplicationController Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedError = err.(error)
		}
	}()

	url := "http://" + kubeapiHost + ":" + strconv.Itoa(kubeapiPort) + "/api/v1/namespaces/" + namespace + "/replicationcontrollers/" + replicationControllerName
	_, err := restclient.RequestDelete(url, nil, nil, true)
	return err
}

func DeleteReplicationControllerAndRelatedPod(kubeapiHost string, kubeapiPort int, namespace string, replicationControllerName string) (returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("DeleteReplicationControllerAndRelatedPod Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedError = err.(error)
		}
	}()

	err := DeleteReplicationController(kubeapiHost, kubeapiPort, namespace, replicationControllerName)
	if err != nil {
		return err
	} else {
		nameSlice, err := GetAllPodNameBelongToReplicationController(kubeapiHost, kubeapiPort, namespace, replicationControllerName)
		if err != nil {
			return err
		} else {
			for _, name := range nameSlice {
				err := DeletePod(kubeapiHost, kubeapiPort, namespace, name)
				if err != nil {
					return err
				}
			}
			return nil
		}
	}
}

func GetReplicationController(kubeapiHost string, kubeapiPort int, namespace string, replicationControllerName string) (replicationController *ReplicationController, returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("GetReplicationController Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			replicationController = nil
			returnedError = err.(error)
		}
	}()

	url := "http://" + kubeapiHost + ":" + strconv.Itoa(kubeapiPort) + "/api/v1/namespaces/" + namespace + "/replicationcontrollers/" + replicationControllerName
	result, err := restclient.RequestGet(url, nil, true)
	jsonMap, _ := result.(map[string]interface{})
	if err != nil {
		return nil, err
	} else {
		replicationController := new(ReplicationController)
		replicationController.Name, _ = jsonMap["metadata"].(map[string]interface{})["name"].(string)
		selector := jsonMap["spec"].(map[string]interface{})["selector"]
		if selector != nil {
			replicationController.Selector.Name, _ = selector.(map[string]interface{})["name"].(string)
			replicationController.Selector.Version, _ = selector.(map[string]interface{})["version"].(string)
		}

		replicas, _ := jsonparse.ConvertToInt64(jsonMap["spec"].(map[string]interface{})["replicas"])
		replicationController.ReplicaAmount = int(replicas)
		replicationControllerLabelMap := jsonMap["metadata"].(map[string]interface{})["labels"]
		if replicationControllerLabelMap != nil {
			replicationController.Label.Name, _ = replicationControllerLabelMap.(map[string]interface{})["name"].(string)
		}

		containerSlice := jsonMap["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{})["containers"]
		if containerSlice != nil {
			replicationController.ContainerSlice = make([]ReplicationControllerContainer, 0)
			for _, container := range containerSlice.([]interface{}) {
				replicationControllerContainer := ReplicationControllerContainer{}
				replicationControllerContainer.Name, _ = container.(map[string]interface{})["name"].(string)
				replicationControllerContainer.Image, _ = container.(map[string]interface{})["image"].(string)
				portSlice, _ := container.(map[string]interface{})["ports"].([]interface{})
				replicationControllerContainer.PortSlice = make([]ReplicationControllerContainerPort, 0)
				for _, port := range portSlice {
					replicationControllerContainerPort := ReplicationControllerContainerPort{}
					replicationControllerContainerPort.Name, _ = port.(map[string]interface{})["name"].(string)
					containerPort, _ := jsonparse.ConvertToInt64(port.(map[string]interface{})["containerPort"])
					replicationControllerContainerPort.ContainerPort = int(containerPort)
					replicationControllerContainer.PortSlice = append(replicationControllerContainer.PortSlice, replicationControllerContainerPort)
				}

				environmentSlice, _ := container.(map[string]interface{})["env"].([]interface{})
				replicationControllerContainer.EnvironmentSlice = make([]ReplicationControllerContainerEnvironment, 0)
				for _, environment := range environmentSlice {
					name := environment.(map[string]interface{})["name"].(string)
					value := environment.(map[string]interface{})["value"].(string)
					replicationControllerContainer.EnvironmentSlice = append(replicationControllerContainer.EnvironmentSlice, ReplicationControllerContainerEnvironment{name, value})
				}

				replicationControllerContainer.ResourceMap, _ = container.(map[string]interface{})["resources"].(map[string]interface{})

				replicationController.ContainerSlice = append(replicationController.ContainerSlice, replicationControllerContainer)
			}
		}

		return replicationController, nil
	}
}

func GetAllReplicationControllerName(kubeapiHost string, kubeapiPort int, namespace string) (returnedNameSlice []string, returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("GetAllReplicationController Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedNameSlice = nil
			returnedError = err.(error)
		}
	}()

	url := "http://" + kubeapiHost + ":" + strconv.Itoa(kubeapiPort) + "/api/v1/namespaces/" + namespace + "/replicationcontrollers/"
	result, err := restclient.RequestGet(url, nil, true)
	jsonMap, _ := result.(map[string]interface{})
	if err != nil {
		return nil, err
	} else {
		nameSlice := make([]string, 0)
		for _, item := range jsonMap["items"].([]interface{}) {
			name, ok := item.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
			if ok {
				nameSlice = append(nameSlice, name)
			}
		}
		return nameSlice, nil
	}
}

func UpdateReplicationControllerSize(kubeapiHost string, kubeapiPort int, namespace string, replicationControllerName string, size int) (returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("UpdateReplicationControllerSize Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedError = err.(error)
		}
	}()

	url := "http://" + kubeapiHost + ":" + strconv.Itoa(kubeapiPort) + "/api/v1/namespaces/" + namespace + "/replicationcontrollers/" + replicationControllerName + "/"
	result, err := restclient.RequestGet(url, nil, true)
	jsonMap, _ := result.(map[string]interface{})
	if err != nil {
		log.Error("Get replication controller information failure where size: %d, kubeapiHost: %s, kubeapiPort: %d, namespace: %s, replicationControllerName: %s, err: %s", size, kubeapiHost, kubeapiPort, namespace, replicationControllerName, err.Error())
		return err
	} else {
		jsonMap["spec"].(map[string]interface{})["replicas"] = float64(size)
		_, err := restclient.RequestPut(url, jsonMap, nil, true)

		if err != nil {
			return err
		} else {
			return nil
		}
	}
}

func ResizeReplicationController(kubeapiHost string, kubeapiPort int, namespace string, replicationControllerName string, delta int, maximumSize int, minimumSize int) (resized bool, size int, returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("ResizeReplicationController Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			resized = false
			size = -1
			returnedError = err.(error)
		}
	}()

	url := "http://" + kubeapiHost + ":" + strconv.Itoa(kubeapiPort) + "/api/v1/namespaces/" + namespace + "/replicationcontrollers/" + replicationControllerName + "/"
	result, err := restclient.RequestGet(url, nil, true)
	jsonMap, _ := result.(map[string]interface{})
	if err != nil {
		log.Error("Get replication controller information failure where delta: %d, kubeapiHost: %s, kubeapiPort: %d, namespace: %s, replicationControllerName: %s, err: %s", delta, kubeapiHost, kubeapiPort, namespace, replicationControllerName, err.Error())
		return false, -1, err
	} else {
		replicas, _ := jsonparse.ConvertToInt64(jsonMap["spec"].(map[string]interface{})["replicas"])
		currentSize := int(replicas)
		newSize := currentSize + delta
		if newSize < minimumSize {
			newSize = minimumSize
		}
		if newSize > maximumSize {
			newSize = maximumSize
		}

		if newSize == currentSize {
			return false, currentSize, nil
		}

		jsonMap["spec"].(map[string]interface{})["replicas"] = float64(newSize)
		result, err := restclient.RequestPut(url, jsonMap, nil, true)
		resultJsonMap, _ := result.(map[string]interface{})
		if err != nil {
			return false, currentSize, err
		} else {
			replicas, _ := jsonparse.ConvertToInt64(resultJsonMap["spec"].(map[string]interface{})["replicas"])
			return true, int(replicas), nil
		}
	}
}

func RollingUpdateReplicationControllerWithSingleContainer(kubeapiHost string,
	kubeapiPort int, namespace string, replicationControllerName string,
	newReplicationControllerName string, newImage string, newVersion string,
	waitingDuration time.Duration,
	environmentSlice []ReplicationControllerContainerEnvironment) (returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("RollingUpdateReplicationController Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedError = err.(error)
		}
	}()

	oldReplicationController, err := GetReplicationController(kubeapiHost, kubeapiPort, namespace, replicationControllerName)
	if err != nil {
		log.Error("Get old replication controller kubeapiHost: %s, kubeapiPort: %d, namespace: %s, replicationControllerName:%s, error: %s", kubeapiHost, kubeapiPort, namespace, replicationControllerName, err)
		return err
	}

	desiredAmount := oldReplicationController.ReplicaAmount

	newReplicationController := ReplicationController{
		newReplicationControllerName,
		0,
		ReplicationControllerSelector{
			oldReplicationController.Selector.Name,
			newVersion,
		},
		ReplicationControllerLabel{
			newReplicationControllerName,
		},
		make([]ReplicationControllerContainer, 1),
		nil,
	}
	newReplicationController.Label.Name = newReplicationControllerName
	newReplicationController.ContainerSlice[0].Name = newReplicationControllerName
	newReplicationController.ContainerSlice[0].Image = newImage
	newReplicationController.ContainerSlice[0].PortSlice = oldReplicationController.ContainerSlice[0].PortSlice
	newReplicationController.ContainerSlice[0].EnvironmentSlice = environmentSlice
	newReplicationController.ContainerSlice[0].ResourceMap = oldReplicationController.ContainerSlice[0].ResourceMap

	err = CreateReplicationController(kubeapiHost, kubeapiPort, namespace, newReplicationController)
	if err != nil {
		log.Error("Create new replication controller error: %s", err)
		return err
	}

	for newReplicationController.ReplicaAmount < desiredAmount || oldReplicationController.ReplicaAmount > 0 {
		time.Sleep(waitingDuration)
		_, newReplicationController.ReplicaAmount, err = ResizeReplicationController(kubeapiHost, kubeapiPort, namespace, newReplicationController.Name, 1, desiredAmount, 0)
		if err != nil {
			log.Error("Resize new replication controller error: %s", err)
			return err
		}
		time.Sleep(waitingDuration)
		_, oldReplicationController.ReplicaAmount, err = ResizeReplicationController(kubeapiHost, kubeapiPort, namespace, oldReplicationController.Name, -1, desiredAmount, 0)
		if err != nil {
			log.Error("Resize old replication controller error: %s", err)
			return err
		}
	}

	time.Sleep(waitingDuration)

	return DeleteReplicationController(kubeapiHost, kubeapiPort, namespace, replicationControllerName)
}

func CreateReplicationControllerWithJson(kubeapiHost string, kubeapiPort int, namespace string, bodyJsonMap map[string]interface{}) (returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("CreateReplicationControllerWithJson Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedError = err.(error)
		}
	}()

	url := "http://" + kubeapiHost + ":" + strconv.Itoa(kubeapiPort) + "/api/v1/namespaces/" + namespace + "/replicationcontrollers/"
	_, err := restclient.RequestPost(url, bodyJsonMap, nil, true)

	if err != nil {
		log.Error(err)
		return err
	} else {
		return nil
	}
}

func UpdateReplicationControllerWithJson(kubeapiHost string, kubeapiPort int, namespace string, replicationControllerName string, bodyJsonMap map[string]interface{}) (returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("UpdateReplicationControllerWithJson Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedError = err.(error)
		}
	}()

	url := "http://" + kubeapiHost + ":" + strconv.Itoa(kubeapiPort) + "/api/v1/namespaces/" + namespace + "/replicationcontrollers/" + replicationControllerName
	result, err := restclient.RequestGet(url, nil, true)
	jsonMap, _ := result.(map[string]interface{})

	metadataJsonMap, _ := jsonMap["metadata"].(map[string]interface{})
	resourceVersion, _ := metadataJsonMap["resourceVersion"].(string)

	// Update requires the resoruce version
	bodyJsonMap["metadata"].(map[string]interface{})["resourceVersion"] = resourceVersion

	url = "http://" + kubeapiHost + ":" + strconv.Itoa(kubeapiPort) + "/api/v1/namespaces/" + namespace + "/replicationcontrollers/" + replicationControllerName
	_, err = restclient.RequestPut(url, bodyJsonMap, nil, true)

	if err != nil {
		log.Error(err)
		return err
	} else {
		return nil
	}
}
