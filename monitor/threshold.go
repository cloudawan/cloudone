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

import ()

const (
	CPU    = "cpu"
	Memory = "memory"
)

func CheckThresholdReplicationController(indicator string, aboveOrBelow bool, allOrOne bool, replicationControllerMetric *ReplicationControllerMetric, percentageOfData float64, threshold int64) bool {
	for index, valid := range replicationControllerMetric.ValidPodSlice {
		if valid {
			result := false
			switch indicator {
			case CPU:
				result = CheckThresholdPodCPU(aboveOrBelow, &replicationControllerMetric.PodMetricSlice[index], percentageOfData, threshold)
			case Memory:
				result = CheckThresholdPodMemory(aboveOrBelow, &replicationControllerMetric.PodMetricSlice[index], percentageOfData, threshold)
			default:
				log.Error("CheckThresholdReplicationController no such indicator %s", indicator)
				return false
			}

			if allOrOne {
				if result == false {
					return false
				}
			} else {
				if result {
					return true
				}
			}
		}
	}
	if allOrOne {
		return true
	} else {
		return false
	}
}

func CheckThresholdPodCPU(aboveOrBelow bool, podMetric *PodMetric, percentageOfData float64, threshold int64) bool {

	for index, valid := range podMetric.ValidContainerSlice {
		if valid {
			differenceSlice := make([]int64, len(podMetric.ContainerMetricSlice[index].CpuUsageTotalSlice)-1)
			for i := 0; i < len(differenceSlice); i++ {
				differenceSlice[i] = podMetric.ContainerMetricSlice[index].CpuUsageTotalSlice[i+1] - podMetric.ContainerMetricSlice[index].CpuUsageTotalSlice[i]
			}

			if passThreshold(aboveOrBelow, differenceSlice, percentageOfData, threshold) {
				return true
			}
		}
	}

	return false
}

func CheckThresholdPodMemory(aboveOrBelow bool, podMetric *PodMetric, percentageOfData float64, threshold int64) bool {
	for index, valid := range podMetric.ValidContainerSlice {
		if valid {
			if passThreshold(aboveOrBelow, podMetric.ContainerMetricSlice[index].MemoryUsageSlice, percentageOfData, threshold) {
				return true
			}
		}
	}

	return false
}

func passThreshold(aboveOrBelow bool, slice []int64, percentageOfData float64, threshold int64) bool {
	thresholdAmount := 0
	for _, value := range slice {
		if aboveOrBelow {
			if value > threshold {
				thresholdAmount++
			}
		} else {
			if value < threshold {
				thresholdAmount++
			}
		}
	}

	if float64(thresholdAmount)/float64(len(slice)) >= percentageOfData {
		return true
	} else {
		return false
	}
}
