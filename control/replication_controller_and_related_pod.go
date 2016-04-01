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

type ReplicationControllerAndRelatedPod struct {
	Name               string
	Namespace          string
	ReplicaAmount      int
	AliveReplicaAmount int
	Selector           map[string]string
	Label              map[string]string
	PodSlice           []Pod
}

func GetAllReplicationControllerAndRelatedPodSlice(kubeapiHost string, kubeapiPort int, namespace string) (
	returnedReplicationControllerAndRelatedPodSlice []ReplicationControllerAndRelatedPod, returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("GetAllReplicationController Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedReplicationControllerAndRelatedPodSlice = nil
			returnedError = err.(error)
		}
	}()

	result, err := restclient.RequestGet("http://"+kubeapiHost+":"+strconv.Itoa(kubeapiPort)+"/api/v1/namespaces/"+namespace+"/replicationcontrollers/", nil, true)
	replicationControllerJsonMap, _ := result.(map[string]interface{})
	if err != nil {
		log.Error("Fail to get all replication controller inofrmation with host %s, port: %d, namespace: %s, error %s", kubeapiHost, kubeapiPort, namespace, err.Error())
		return nil, err
	} else {
		result, err := restclient.RequestGet("http://"+kubeapiHost+":"+strconv.Itoa(kubeapiPort)+"/api/v1/namespaces/"+namespace+"/pods/", nil, true)
		podJsonMap, _ := result.(map[string]interface{})
		if err != nil {
			log.Error("Fail to get all pod information with host %s, port: %d, namespace: %s, error %s", kubeapiHost, kubeapiPort, namespace, err.Error())
			return nil, err
		} else {
			replicationControllerAndRelatedPodSlice := make([]ReplicationControllerAndRelatedPod, 0)

			itemSlice, _ := replicationControllerJsonMap["items"].([]interface{})
			for _, item := range itemSlice {
				name, _ := item.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
				namespace, _ := item.(map[string]interface{})["metadata"].(map[string]interface{})["namespace"].(string)
				replicaAmount, _ := jsonparse.ConvertToInt64(item.(map[string]interface{})["spec"].(map[string]interface{})["replicas"])
				aliveReplicaAmount, _ := jsonparse.ConvertToInt64(item.(map[string]interface{})["status"].(map[string]interface{})["replicas"])
				selectorField, _ := item.(map[string]interface{})["spec"].(map[string]interface{})["selector"].(map[string]interface{})
				labelField, _ := item.(map[string]interface{})["metadata"].(map[string]interface{})["labels"].(map[string]interface{})

				selectorMap := make(map[string]string)
				for key, value := range selectorField {
					selectorMap[key], _ = value.(string)
				}

				labelMap := make(map[string]string)
				for key, value := range labelField {
					labelMap[key], _ = value.(string)
				}

				podSlice, _ := GetAllPodBelongToReplicationControllerFromData(name, podJsonMap)
				replicationControllerAndRelatedPod := ReplicationControllerAndRelatedPod{
					name,
					namespace,
					int(replicaAmount),
					int(aliveReplicaAmount),
					selectorMap,
					labelMap,
					podSlice,
				}

				replicationControllerAndRelatedPodSlice = append(replicationControllerAndRelatedPodSlice, replicationControllerAndRelatedPod)
			}

			return replicationControllerAndRelatedPodSlice, nil
		}
	}
}
