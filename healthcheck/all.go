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
	hostControl, err := CreateHostControl()
	if err != nil {
		log.Error(err)
		return jsonMap, err
	}
	// Kubernetes
	jsonMap["kubernetes"], err = hostControl.GetKubernetesHostStatus()
	if err != nil {
		log.Error(err)
		return jsonMap, err
	}
	// Glusterfs
	jsonMap["glusterfs"], err = hostControl.GetGlusterfsHostStatus()
	if err != nil {
		log.Error(err)
		return jsonMap, err
	}
	// SLB
	jsonMap["slb"], err = hostControl.GetSLBHostStatus()
	if err != nil {
		log.Error(err)
		return jsonMap, err
	}
	// CloudOne
	cloudoneControl, err := CreateCloudoneControl()
	if err != nil {
		log.Error(err)
		return jsonMap, err
	}
	jsonMap["cloudone"] = cloudoneControl.GetStatus()

	return jsonMap, nil
}
