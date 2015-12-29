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

package healthcheck

import (
	"encoding/json"
	"github.com/cloudawan/cloudone/control/glusterfs"
	"strings"
)

func GetAllStatus() (map[string]interface{}, error) {
	jsonMap := make(map[string]interface{})
	// Kubernetes
	jsonMap["kubernetes"] = make(map[string]interface{})
	kubernetesNodeControl, err := CreateKubernetesNodeControl()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	kubernetesNodeJsonMap, err := kubernetesNodeControl.GetStatus()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	ipSlice, err := kubernetesNodeControl.GetHostWithinFlannelNetwork()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	for _, data := range kubernetesNodeJsonMap["node"].(map[string]interface{})["nodes"].([]interface{}) {
		dataJsonMap, _ := data.(map[string]interface{})
		key, _ := dataJsonMap["key"].(string)
		splitSlice := strings.Split(key, "/")
		ip := splitSlice[len(splitSlice)-1]
		valueText, _ := dataJsonMap["value"].(string)
		valueJsonMap := make(map[string]interface{})
		err := json.Unmarshal([]byte(valueText), &valueJsonMap)
		if err != nil {
			log.Error(err)
		}
		jsonMap["kubernetes"].(map[string]interface{})[ip] = valueJsonMap
	}
	for _, ip := range ipSlice {
		if jsonMap["kubernetes"].(map[string]interface{})[ip] == nil {
			jsonMap["kubernetes"].(map[string]interface{})[ip] = make(map[string]interface{})
			jsonMap["kubernetes"].(map[string]interface{})[ip].(map[string]interface{})["active"] = false
		} else {
			jsonMap["kubernetes"].(map[string]interface{})[ip].(map[string]interface{})["active"] = true
		}
	}
	// Glusterfs
	glusterfsVolumeControl, err := glusterfs.CreateGlusterfsVolumeControl()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	hostStatusMap := glusterfsVolumeControl.GetHostStatus()
	jsonMap["glusterfs"] = make(map[string]interface{})
	for key, value := range hostStatusMap {
		jsonMap["glusterfs"].(map[string]interface{})[key] = make(map[string]interface{})
		jsonMap["glusterfs"].(map[string]interface{})[key].(map[string]interface{})["active"] = true
		jsonMap["glusterfs"].(map[string]interface{})[key].(map[string]interface{})["service"] = make(map[string]interface{})
		jsonMap["glusterfs"].(map[string]interface{})[key].(map[string]interface{})["service"].(map[string]interface{})["glusterfs"] = value
	}
	// CloudOne
	cloudoneControl, err := CreateCloudoneControl()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	jsonMap["cloudone"] = cloudoneControl.GetStatus()

	return jsonMap, nil
}
