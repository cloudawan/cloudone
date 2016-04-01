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
	"github.com/cloudawan/cloudone_utility/jsonparse"
	"github.com/cloudawan/cloudone_utility/logger"
	"github.com/cloudawan/cloudone_utility/restclient"
	"strconv"
)

type PodMetric struct {
	KubeletHost          string
	Namespace            string
	PodName              string
	ValidContainerSlice  []bool
	ContainerMetricSlice []ContainerMetric
}

type ContainerMetric struct {
	ContainerName                     string
	CpuUsageTotalSlice                []int64
	MemoryUsageSlice                  []int64
	DiskIOServiceBytesStatsTotalSlice []int64
	DiskIOServicedStatsTotalSlice     []int64
	NetworkRXBytesSlice               []int64
	NetworkTXBytesSlice               []int64
	NetworkRXPacketsSlice             []int64
	NetworkTXPacketsSlice             []int64
}

func MonitorPod(kubeapiHost string, kubeapiPort int, namespace string, podName string) (returnedPodMetric *PodMetric, returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("MonitorPod Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedPodMetric = nil
			returnedError = err.(error)
		}
	}()

	result, err := restclient.RequestGet("http://"+kubeapiHost+":"+strconv.Itoa(kubeapiPort)+"/api/v1/namespaces/"+namespace+"/pods/"+podName+"/", nil, true)
	jsonMap, _ := result.(map[string]interface{})
	if err != nil {
		log.Error("Fail to get pod inofrmation with host %s, port: %d, namespace: %s, pod name: %s, error %s", kubeapiHost, kubeapiPort, namespace, podName, err.Error())
		return nil, err
	}
	urlSlice, containerNameSlice, kubeletHost := getContainerLocationFromPodInformation(jsonMap)
	jsonMap["container_url_slice"] = urlSlice
	dataSlice, errorSlice := getContainerMonitorData(urlSlice)
	jsonMap["container_monitor_data_slice"] = dataSlice
	jsonMap["container_monitor_error_slice"] = errorSlice

	podMetric := &PodMetric{}
	podMetric.KubeletHost = kubeletHost
	podMetric.Namespace = namespace
	podMetric.PodName = podName
	podMetric.ValidContainerSlice = make([]bool, len(dataSlice))
	podMetric.ContainerMetricSlice = make([]ContainerMetric, len(dataSlice))
	errorMessage := "The following index of container has error: "
	errorHappened := false
	for index, data := range dataSlice {
		if errorSlice[index] != nil {
			errorMessage = errorMessage + errorSlice[index].Error()
			podMetric.ValidContainerSlice[index] = false
			errorHappened = true
		} else {
			podMetric.ValidContainerSlice[index] = true
			podMetric.ContainerMetricSlice[index] = ContainerMetric{}
			podMetric.ContainerMetricSlice[index].ContainerName = containerNameSlice[index]
			length := len(data["stats"].([]interface{}))
			// CPU
			podMetric.ContainerMetricSlice[index].CpuUsageTotalSlice = make([]int64, length)
			// Memory
			podMetric.ContainerMetricSlice[index].MemoryUsageSlice = make([]int64, length)
			// Disk I/O
			podMetric.ContainerMetricSlice[index].DiskIOServiceBytesStatsTotalSlice = make([]int64, length)
			podMetric.ContainerMetricSlice[index].DiskIOServicedStatsTotalSlice = make([]int64, length)
			// Network
			podMetric.ContainerMetricSlice[index].NetworkRXBytesSlice = make([]int64, length)
			podMetric.ContainerMetricSlice[index].NetworkTXBytesSlice = make([]int64, length)
			podMetric.ContainerMetricSlice[index].NetworkRXPacketsSlice = make([]int64, length)
			podMetric.ContainerMetricSlice[index].NetworkTXPacketsSlice = make([]int64, length)
			for i := 0; i < length; i++ {
				// CPU
				podMetric.ContainerMetricSlice[index].CpuUsageTotalSlice[i], _ = jsonparse.ConvertToInt64(data["stats"].([]interface{})[i].(map[string]interface{})["cpu"].(map[string]interface{})["usage"].((map[string]interface{}))["total"])
				// Memory
				podMetric.ContainerMetricSlice[index].MemoryUsageSlice[i], _ = jsonparse.ConvertToInt64(data["stats"].([]interface{})[i].(map[string]interface{})["memory"].(map[string]interface{})["usage"])
				// Disk I/O
				ioServiceBytesSlice := data["stats"].([]interface{})[i].(map[string]interface{})["diskio"].(map[string]interface{})["io_service_bytes"]
				if ioServiceBytesSlice != nil {
					for _, ioServiceBytes := range ioServiceBytesSlice.([]interface{}) {
						value, _ := jsonparse.ConvertToInt64(ioServiceBytes.(map[string]interface{})["stats"].(map[string]interface{})["Total"])
						podMetric.ContainerMetricSlice[index].DiskIOServiceBytesStatsTotalSlice[i] += value
					}
				}
				ioServicedSlice := data["stats"].([]interface{})[i].(map[string]interface{})["diskio"].(map[string]interface{})["io_serviced"]
				if ioServicedSlice != nil {
					for _, ioServiced := range ioServicedSlice.([]interface{}) {
						value, _ := jsonparse.ConvertToInt64(ioServiced.(map[string]interface{})["stats"].(map[string]interface{})["Total"])
						podMetric.ContainerMetricSlice[index].DiskIOServicedStatsTotalSlice[i] += value
					}
				}
				// Network
				podMetric.ContainerMetricSlice[index].NetworkRXBytesSlice[i], _ = jsonparse.ConvertToInt64(data["stats"].([]interface{})[i].(map[string]interface{})["network"].(map[string]interface{})["rx_bytes"])
				podMetric.ContainerMetricSlice[index].NetworkTXBytesSlice[i], _ = jsonparse.ConvertToInt64(data["stats"].([]interface{})[i].(map[string]interface{})["network"].(map[string]interface{})["tx_bytes"])
				podMetric.ContainerMetricSlice[index].NetworkRXPacketsSlice[i], _ = jsonparse.ConvertToInt64(data["stats"].([]interface{})[i].(map[string]interface{})["network"].(map[string]interface{})["rx_packets"])
				podMetric.ContainerMetricSlice[index].NetworkTXPacketsSlice[i], _ = jsonparse.ConvertToInt64(data["stats"].([]interface{})[i].(map[string]interface{})["network"].(map[string]interface{})["tx_packets"])
			}
		}
	}

	if errorHappened {
		log.Error("Fail to get all container inofrmation with host %s, port: %d, namespace: %s, pod name: %s, error %s", kubeapiHost, kubeapiPort, namespace, podName, errorMessage)
		return podMetric, errors.New(errorMessage)
	} else {
		return podMetric, nil
	}
}

func getContainerLocationFromPodInformation(podInformationJsonMap map[string]interface{}) ([]string, []string, string) {
	urlSlice := make([]string, 0)
	containerNameSlice := make([]string, 0)

	kubeletHost, _ := podInformationJsonMap["status"].(map[string]interface{})["hostIP"].(string)
	podName, _ := podInformationJsonMap["metadata"].(map[string]interface{})["name"].(string)
	namespace, _ := podInformationJsonMap["metadata"].(map[string]interface{})["namespace"].(string)
	uid, _ := podInformationJsonMap["metadata"].(map[string]interface{})["uid"].(string)

	containerInformationSlice := podInformationJsonMap["spec"].(map[string]interface{})["containers"]
	for _, container := range containerInformationSlice.([]interface{}) {
		containerName, ok := container.(map[string]interface{})["name"].(string)
		if ok {
			containerNameSlice = append(containerNameSlice, containerName)
			urlSlice = append(urlSlice, "https://"+kubeletHost+":10250/stats/"+namespace+"/"+podName+"/"+uid+"/"+containerName)
		}
	}

	return urlSlice, containerNameSlice, kubeletHost
}

func getContainerMonitorData(urlSlice []string) ([]map[string]interface{}, []error) {
	dataMapSlice := make([]map[string]interface{}, 0)
	errorSlice := make([]error, 0)
	for _, url := range urlSlice {
		result, err := restclient.RequestGet(url, nil, true)
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
