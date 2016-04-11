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

package autoscaler

import (
	"bytes"
	"errors"
	"github.com/cloudawan/cloudone/control"
	"github.com/cloudawan/cloudone/monitor"
	"time"
)

type ReplicationControllerAutoScaler struct {
	Check             bool
	CoolDownDuration  time.Duration
	RemainingCoolDown time.Duration
	KubeapiHost       string
	KubeapiPort       int
	Namespace         string
	Kind              string
	Name              string
	MaximumReplica    int
	MinimumReplica    int
	IndicatorSlice    []Indicator
}

type Indicator struct {
	Type                  string
	AboveAllOrOne         bool
	AbovePercentageOfData float64
	AboveThreshold        int64
	BelowAllOrOne         bool
	BelowPercentageOfData float64
	BelowThreshold        int64
}

func CheckAndExecuteAutoScaler(replicationControllerAutoScaler *ReplicationControllerAutoScaler) (bool, int, error) {
	switch replicationControllerAutoScaler.Kind {
	case "selector":
		nameSlice, err := monitor.GetReplicationControllerNameFromSelector(
			replicationControllerAutoScaler.KubeapiHost,
			replicationControllerAutoScaler.KubeapiPort,
			replicationControllerAutoScaler.Namespace,
			replicationControllerAutoScaler.Name)
		if err != nil {
			return false, -1, errors.New("Could not find replication controller name with selector " + replicationControllerAutoScaler.Name + " error " + err.Error())
		} else {
			resized := false
			hasError := false
			errorMessage := bytes.Buffer{}
			totalSize := 0
			for _, name := range nameSlice {
				result, size, err := CheckAndExecuteAutoScalerOnReplicationController(replicationControllerAutoScaler, name)
				resized = resized || result
				totalSize += size
				if err != nil {
					hasError = true
					errorMessage.WriteString(err.Error())
				}
			}

			if hasError {
				return false, -1, errors.New(errorMessage.String())
			} else {
				return resized, totalSize, nil
			}
		}
	case "replicationController":
		return CheckAndExecuteAutoScalerOnReplicationController(replicationControllerAutoScaler, replicationControllerAutoScaler.Name)
	default:
		return false, -1, errors.New("No such kind " + replicationControllerAutoScaler.Kind)
	}
}

func CheckAndExecuteAutoScalerOnReplicationController(replicationControllerAutoScaler *ReplicationControllerAutoScaler, replicationControllerName string) (bool, int, error) {
	replicationControllerMetric, err := monitor.MonitorReplicationController(replicationControllerAutoScaler.KubeapiHost, replicationControllerAutoScaler.KubeapiPort, replicationControllerAutoScaler.Namespace, replicationControllerName)
	if err != nil {
		log.Error("Get ReplicationController data failure: %s where replicationControllerAutoScaler %v", err.Error(), replicationControllerAutoScaler)
		return false, -1, err
	}
	toIncrease, toDecrease := false, false
	for _, indicator := range replicationControllerAutoScaler.IndicatorSlice {
		toIncrease = monitor.CheckThresholdReplicationController(indicator.Type, true, indicator.AboveAllOrOne, replicationControllerMetric, indicator.AbovePercentageOfData, indicator.AboveThreshold)
		if toIncrease {
			break
		}
		toDecrease = monitor.CheckThresholdReplicationController(indicator.Type, false, indicator.BelowAllOrOne, replicationControllerMetric, indicator.BelowPercentageOfData, indicator.BelowThreshold)
		if toDecrease {
			break
		}
	}

	if toIncrease {
		resized, size, err := control.ResizeReplicationController(replicationControllerAutoScaler.KubeapiHost, replicationControllerAutoScaler.KubeapiPort, replicationControllerAutoScaler.Namespace, replicationControllerName, 1, replicationControllerAutoScaler.MaximumReplica, replicationControllerAutoScaler.MinimumReplica)
		if err != nil {
			log.Error("ResizeReplicationController failure: %s where ReplicationControllerAutoScaler %v", err.Error(), replicationControllerAutoScaler)
		}
		return resized, size, err
	} else if toDecrease {
		resized, size, err := control.ResizeReplicationController(replicationControllerAutoScaler.KubeapiHost, replicationControllerAutoScaler.KubeapiPort, replicationControllerAutoScaler.Namespace, replicationControllerName, -1, replicationControllerAutoScaler.MaximumReplica, replicationControllerAutoScaler.MinimumReplica)
		if err != nil {
			log.Error("ResizeReplicationController failure: %s where ReplicationControllerAutoScaler %v", err.Error(), replicationControllerAutoScaler)
		}
		return resized, size, err
	} else {
		return false, replicationControllerMetric.Size, nil
	}
}
