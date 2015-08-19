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
	"github.com/cloudawan/kubernetes_management_utility/jsonparse"
	"github.com/cloudawan/kubernetes_management_utility/logger"
	"github.com/cloudawan/kubernetes_management_utility/restclient"
	"strconv"
)

type NodeMetric struct {
	Valid                             bool
	KubeletHost                       string
	CpuUsageTotalSlice                []int64
	MemoryUsageSlice                  []int64
	DiskIOServiceBytesStatsTotalSlice []int64
	DiskIOServicedStatsTotalSlice     []int64
	NetworkRXBytesSlice               []int64
	NetworkTXBytesSlice               []int64
	NetworkRXPacketsSlice             []int64
	NetworkTXPacketsSlice             []int64
}

func MonitorNode(kubeapiHost string, kubeapiPort int) (returnedNodeMetricSlice []NodeMetric, returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("MonitorNode Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedNodeMetricSlice = nil
			returnedError = err.(error)
		}
	}()

	result, err := restclient.RequestGet("http://"+kubeapiHost+":"+strconv.Itoa(kubeapiPort)+"/api/v1/nodes", true)
	jsonMap, _ := result.(map[string]interface{})
	if err != nil {
		log.Error("Fail to get node inofrmation with hos:t %s, port: %d, error %s", kubeapiHost, kubeapiPort, err.Error())
		return nil, err
	}
	urlSlice, addressSlice := getNodeLocationFromNodeInformation(jsonMap)
	dataSlice, errorSlice := getNodeMonitorData(urlSlice)

	nodeAmount := len(dataSlice)
	nodeMetricSlice := make([]NodeMetric, nodeAmount)
	errorMessage := "The following index of container has error: "
	errorHappened := false
	for index, data := range dataSlice {
		if errorSlice[index] != nil {
			errorMessage = errorMessage + errorSlice[index].Error()
			nodeMetricSlice[index].Valid = false
			errorHappened = true
		} else {
			nodeMetricSlice[index].Valid = true
			nodeMetricSlice[index].KubeletHost = addressSlice[index]
			length := len(data["stats"].([]interface{}))
			// CPU
			nodeMetricSlice[index].CpuUsageTotalSlice = make([]int64, length)
			// Memory
			nodeMetricSlice[index].MemoryUsageSlice = make([]int64, length)
			// Disk I/O
			nodeMetricSlice[index].DiskIOServiceBytesStatsTotalSlice = make([]int64, length)
			nodeMetricSlice[index].DiskIOServicedStatsTotalSlice = make([]int64, length)
			// Network
			nodeMetricSlice[index].NetworkRXBytesSlice = make([]int64, length)
			nodeMetricSlice[index].NetworkTXBytesSlice = make([]int64, length)
			nodeMetricSlice[index].NetworkRXPacketsSlice = make([]int64, length)
			nodeMetricSlice[index].NetworkTXPacketsSlice = make([]int64, length)
			for i := 0; i < length; i++ {
				// CPU
				nodeMetricSlice[index].CpuUsageTotalSlice[i], _ = jsonparse.ConvertToInt64(data["stats"].([]interface{})[i].(map[string]interface{})["cpu"].(map[string]interface{})["usage"].((map[string]interface{}))["total"])
				// Memory
				nodeMetricSlice[index].MemoryUsageSlice[i], _ = jsonparse.ConvertToInt64(data["stats"].([]interface{})[i].(map[string]interface{})["memory"].(map[string]interface{})["usage"])
				// Disk I/O
				ioServiceBytesSlice := data["stats"].([]interface{})[i].(map[string]interface{})["diskio"].(map[string]interface{})["io_service_bytes"].([]interface{})
				for _, ioServiceBytes := range ioServiceBytesSlice {
					value, _ := jsonparse.ConvertToInt64(ioServiceBytes.(map[string]interface{})["stats"].(map[string]interface{})["Total"])
					nodeMetricSlice[index].DiskIOServiceBytesStatsTotalSlice[i] += value
				}
				ioServicedSlice := data["stats"].([]interface{})[i].(map[string]interface{})["diskio"].(map[string]interface{})["io_serviced"].([]interface{})
				for _, ioServiced := range ioServicedSlice {
					value, _ := jsonparse.ConvertToInt64(ioServiced.(map[string]interface{})["stats"].(map[string]interface{})["Total"])
					nodeMetricSlice[index].DiskIOServicedStatsTotalSlice[i] += value
				}
				// Network
				nodeMetricSlice[index].NetworkRXBytesSlice[i], _ = jsonparse.ConvertToInt64(data["stats"].([]interface{})[i].(map[string]interface{})["network"].(map[string]interface{})["rx_bytes"])
				nodeMetricSlice[index].NetworkTXBytesSlice[i], _ = jsonparse.ConvertToInt64(data["stats"].([]interface{})[i].(map[string]interface{})["network"].(map[string]interface{})["tx_bytes"])
				nodeMetricSlice[index].NetworkRXPacketsSlice[i], _ = jsonparse.ConvertToInt64(data["stats"].([]interface{})[i].(map[string]interface{})["network"].(map[string]interface{})["rx_packets"])
				nodeMetricSlice[index].NetworkTXPacketsSlice[i], _ = jsonparse.ConvertToInt64(data["stats"].([]interface{})[i].(map[string]interface{})["network"].(map[string]interface{})["tx_packets"])
			}
		}
	}

	if errorHappened {
		log.Error("Fail to get all node inofrmation with host %s, port: %d, error %s", kubeapiHost, kubeapiPort, errorMessage)
		return nodeMetricSlice, errors.New(errorMessage)
	} else {
		return nodeMetricSlice, nil
	}
}

func getNodeLocationFromNodeInformation(nodeInformationJsonMap map[string]interface{}) ([]string, []string) {
	urlSlice := make([]string, 0)
	addressSlice := make([]string, 0)

	for _, item := range nodeInformationJsonMap["items"].([]interface{}) {
		address, ok := item.(map[string]interface{})["status"].(map[string]interface{})["addresses"].([]interface{})[0].(map[string]interface{})["address"].(string)
		if ok {
			urlSlice = append(urlSlice, "https://"+address+":10250/stats/")
			addressSlice = append(addressSlice, address)
		}
	}

	return urlSlice, addressSlice
}

func getNodeMonitorData(urlSlice []string) ([]map[string]interface{}, []error) {
	dataMapSlice := make([]map[string]interface{}, 0)
	errorSlice := make([]error, 0)
	for _, url := range urlSlice {
		result, err := restclient.RequestGet(url, true)
		jsonMap, _ := result.(map[string]interface{})
		if err != nil {
			dataMapSlice = append(dataMapSlice, nil)
			errorSlice = append(errorSlice, err)
		} else {
			dataMapSlice = append(dataMapSlice, jsonMap)
			errorSlice = append(errorSlice, nil)
		}
	}

	return dataMapSlice, errorSlice
}
