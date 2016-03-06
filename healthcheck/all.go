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
//"github.com/cloudawan/cloudone/control/glusterfs"
)

func GetAllStatus() (map[string]interface{}, error) {
	jsonMap := make(map[string]interface{})
	// Kubernetes
	kubernetesNodeControl, err := CreateKubernetesNodeControl()
	if err != nil {
		log.Error(err)
		return jsonMap, err
	}
	jsonMap["kubernetes"], err = kubernetesNodeControl.GetStatus()
	if err != nil {
		log.Error(err)
		return jsonMap, err
	}
	ipSlice, err := kubernetesNodeControl.GetHostWithinFlannelNetwork()
	if err != nil {
		log.Error(err)
		return jsonMap, err
	}
	for key, _ := range jsonMap["kubernetes"].(map[string]interface{}) {
		jsonMap["kubernetes"].(map[string]interface{})[key].(map[string]interface{})["active"] = false
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
	/*
		glusterfsVolumeControl, err := glusterfs.CreateGlusterfsVolumeControl()
		if err != nil {
			log.Error(err)
			return jsonMap, err
		}
		hostStatusMap := glusterfsVolumeControl.GetHostStatus()
		jsonMap["glusterfs"] = make(map[string]interface{})
		for key, value := range hostStatusMap {
			jsonMap["glusterfs"].(map[string]interface{})[key] = make(map[string]interface{})
			jsonMap["glusterfs"].(map[string]interface{})[key].(map[string]interface{})["active"] = true
			jsonMap["glusterfs"].(map[string]interface{})[key].(map[string]interface{})["service"] = make(map[string]interface{})
			jsonMap["glusterfs"].(map[string]interface{})[key].(map[string]interface{})["service"].(map[string]interface{})["glusterfs"] = value
		}
	*/
	// CloudOne
	cloudoneControl, err := CreateCloudoneControl()
	if err != nil {
		log.Error(err)
		return jsonMap, err
	}
	jsonMap["cloudone"] = cloudoneControl.GetStatus()

	return jsonMap, nil
}
