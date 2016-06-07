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
)

type Region struct {
	Name           string
	LocationTagged bool
	ZoneSlice      []Zone
}

type Zone struct {
	Name           string
	LocationTagged bool
	NodeSlice      []Node
}

type Node struct {
	Name     string
	Address  string
	Capacity Capacity
}

type Capacity struct {
	Cpu    string
	Memory string
}

func GetAllNodeIP(kubeApiServerEndPoint string, kubeApiServerToken string) (returnedIpSlice []string, returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("GetAllNodeIP Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedIpSlice = nil
			returnedError = err.(error)
		}
	}()

	jsonMap := make(map[string]interface{})

	headerMap := make(map[string]string)
	headerMap["Authorization"] = kubeApiServerToken

	url := kubeApiServerEndPoint + "/api/v1/nodes/"
	_, err := restclient.RequestGetWithStructure(url, &jsonMap, headerMap)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	ipSlice := make([]string, 0)

	itemSlice, _ := jsonMap["items"].([]interface{})
	for _, item := range itemSlice {
		itemJsonMap, _ := item.(map[string]interface{})
		statusJsonMap, _ := itemJsonMap["status"].(map[string]interface{})
		addressesJsonSlice, _ := statusJsonMap["addresses"].([]interface{})
		address := ""
		for _, value := range addressesJsonSlice {
			addressJsonMap, _ := value.(map[string]interface{})
			addressType, _ := addressJsonMap["type"].(string)
			addressAddress, _ := addressJsonMap["address"].(string)
			if addressType == "InternalIP" {
				address = addressAddress
				break
			}
		}

		if len(address) > 0 {
			ipSlice = append(ipSlice, address)
		}
	}

	return ipSlice, nil
}

func GetNodeTopology(kubeApiServerEndPoint string, kubeApiServerToken string) (returnedRegionSlice []Region, returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("GetNodeTopology Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedRegionSlice = nil
			returnedError = err.(error)
		}
	}()

	jsonMap := make(map[string]interface{})

	headerMap := make(map[string]string)
	headerMap["Authorization"] = kubeApiServerToken

	url := kubeApiServerEndPoint + "/api/v1/nodes/"
	_, err := restclient.RequestGetWithStructure(url, &jsonMap, headerMap)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	zoneSlice := make([]Zone, 0)
	zoneSlice = append(zoneSlice, Zone{"Untagged", false, make([]Node, 0)})
	regionSlice := make([]Region, 0)
	regionSlice = append(regionSlice, Region{"Untagged", false, zoneSlice})

	itemSlice, _ := jsonMap["items"].([]interface{})
	for _, item := range itemSlice {
		itemJsonMap, _ := item.(map[string]interface{})
		metadataJsonMap, _ := itemJsonMap["metadata"].(map[string]interface{})
		labelJsonMap, _ := metadataJsonMap["labels"].(map[string]interface{})
		statusJsonMap, _ := itemJsonMap["status"].(map[string]interface{})
		capacityJsonMap, _ := statusJsonMap["capacity"].(map[string]interface{})
		conditionJsonSlice, _ := statusJsonMap["conditions"].([]interface{})
		ready := false
		for _, value := range conditionJsonSlice {
			conditionJsonMap, _ := value.(map[string]interface{})
			conditionType, _ := conditionJsonMap["type"].(string)
			conditionStatus, _ := conditionJsonMap["status"].(string)
			if conditionType == "Ready" && conditionStatus == "True" {
				ready = true
			}
		}

		if ready == false {
			// Skip the non ready node
			continue
		}

		cpu, _ := capacityJsonMap["cpu"].(string)
		memory, _ := capacityJsonMap["memory"].(string)
		addressesJsonSlice, _ := statusJsonMap["addresses"].([]interface{})
		address := ""
		for _, value := range addressesJsonSlice {
			addressJsonMap, _ := value.(map[string]interface{})
			addressType, _ := addressJsonMap["type"].(string)
			addressAddress, _ := addressJsonMap["address"].(string)
			if addressType == "InternalIP" {
				address = addressAddress
				break
			}
		}

		regionName, regionOk := labelJsonMap["region"].(string)
		zoneName, zoneOk := labelJsonMap["zone"].(string)
		nodeName, nodeOk := labelJsonMap["kubernetes.io/hostname"].(string)

		if nodeOk == false {
			log.Error("Node host name is not found where jsonMap: %v", item)
			continue
		}

		node := Node{
			nodeName,
			address,
			Capacity{
				cpu,
				memory,
			},
		}

		if regionOk && zoneOk {
			regionIndex := -1
			for i, region := range regionSlice {
				if region.Name == regionName {
					regionIndex = i
				}
			}

			if regionIndex == -1 {
				// Not found
				nodeSlice := make([]Node, 0)
				nodeSlice = append(nodeSlice, node)
				zone := Zone{zoneName, true, nodeSlice}
				zoneSlice := make([]Zone, 0)
				zoneSlice = append(zoneSlice, zone)
				region := Region{regionName, true, zoneSlice}
				regionSlice = append(regionSlice, region)
			} else {
				zoneIndex := -1
				for i, zone := range regionSlice[regionIndex].ZoneSlice {
					if zone.Name == zoneName {
						zoneIndex = i
					}
				}
				if zoneIndex == -1 {
					// Not found
					nodeSlice := make([]Node, 0)
					nodeSlice = append(nodeSlice, node)
					zone := Zone{zoneName, true, nodeSlice}
					regionSlice[regionIndex].ZoneSlice = append(regionSlice[regionIndex].ZoneSlice, zone)
				} else {
					regionSlice[regionIndex].ZoneSlice[zoneIndex].NodeSlice = append(regionSlice[regionIndex].ZoneSlice[zoneIndex].NodeSlice, node)
				}
			}
		} else {
			// 0 is untagged
			regionSlice[0].ZoneSlice[0].NodeSlice = append(regionSlice[0].ZoneSlice[0].NodeSlice, node)
		}
	}

	return regionSlice, nil
}
