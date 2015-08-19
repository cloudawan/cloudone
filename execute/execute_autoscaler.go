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

package execute

import (
	"github.com/cloudawan/kubernetes_management/autoscaler"
	"time"
)

var registerAutoScalerChannel = make(chan autoscaler.ReplicationControllerAutoScaler)

var replicationControllerAutoScalerMap = make(map[string]*autoscaler.ReplicationControllerAutoScaler)

func init() {
	// Load from database
	replicationControllerAutoScalerSlice, err := autoscaler.LoadAllReplicationControllerAutoScaler()
	if err != nil {
		log.Error(err)
	} else {
		for _, replicationControllerAutoScaler := range replicationControllerAutoScalerSlice {
			AddReplicationControllerAutoScaler(&replicationControllerAutoScaler)
		}
	}
}

func loopAutoScaler(ticker *time.Ticker, checkingInterval time.Duration) {
	for {
		select {
		// Auto scaler
		case replicationControllerAutoScaler := <-registerAutoScalerChannel:
			receiveFromAutoScalerChannel(replicationControllerAutoScaler)
		case <-ticker.C:
			// Auto scaler
			periodicalCheckAutoScaler(checkingInterval)
		case <-quitChannel:
			ticker.Stop()
			close(registerAutoScalerChannel)
			log.Info("Loop auto scaler quit")
			return
		}
	}
}

func GetReplicationControllerAutoScalerMap() map[string]autoscaler.ReplicationControllerAutoScaler {
	// Return a copy rather than a original one to prevent from concurrency issue
	// This read is not guaranteed for consistency between multiple goroutines. The data may be old but not corrupted.
	returnedReplicationControllerAutoScalerMap := make(map[string]autoscaler.ReplicationControllerAutoScaler)
	for key, value := range replicationControllerAutoScalerMap {
		returnedReplicationControllerAutoScalerMap[key] = *value
	}
	return returnedReplicationControllerAutoScalerMap
}

func GetReplicationControllerAutoScaler(namespace string, kind string, name string) (bool, autoscaler.ReplicationControllerAutoScaler) {
	// Return a copy rather than a original one to prevent from concurrency issue.
	// This read is not guaranteed for consistency between multiple goroutines. The data may be old but not corrupted.
	returnedReplicationControllerAutoScaler := replicationControllerAutoScalerMap[getKeyForReplicationControllerAutoScalerMap(namespace, kind, name)]
	if returnedReplicationControllerAutoScaler != nil {
		return true, *returnedReplicationControllerAutoScaler
	} else {
		return false, autoscaler.ReplicationControllerAutoScaler{}
	}
}

func AddReplicationControllerAutoScaler(replicationControllerAutoScaler *autoscaler.ReplicationControllerAutoScaler) {
	registerAutoScalerChannel <- *replicationControllerAutoScaler
}

func getKeyForReplicationControllerAutoScalerMap(namespace string, kind string, name string) string {
	return namespace + "/" + kind + "/" + name
}

func receiveFromAutoScalerChannel(replicationControllerAutoScaler autoscaler.ReplicationControllerAutoScaler) {
	id := getKeyForReplicationControllerAutoScalerMap(replicationControllerAutoScaler.Namespace, replicationControllerAutoScaler.Kind, replicationControllerAutoScaler.Name)
	if replicationControllerAutoScaler.Check {
		replicationControllerAutoScalerMap[id] = &replicationControllerAutoScaler
	} else {
		delete(replicationControllerAutoScalerMap, id)
	}
}

func periodicalCheckAutoScaler(checkingInterval time.Duration) {
	for _, replicationControllerAutoScaler := range replicationControllerAutoScalerMap {
		if replicationControllerAutoScaler.RemainingCoolDown > 0 {
			replicationControllerAutoScaler.RemainingCoolDown -= checkingInterval
		}
		if replicationControllerAutoScaler.RemainingCoolDown <= 0*time.Second {
			resized, size, err := autoscaler.CheckAndExecuteAutoScaler(replicationControllerAutoScaler)
			if err != nil {
				log.Error("CheckAndExecuteAutoSclae error: %s where ReplicationControllerAutoScaler %s", err.Error(), replicationControllerAutoScaler)
			}
			if resized {
				replicationControllerAutoScaler.RemainingCoolDown = replicationControllerAutoScaler.CoolDownDuration
				log.Info("CheckAndExecuteAutoSclae resized to %d where ReplicationControllerAutoScaler %s", size, replicationControllerAutoScaler)
			}
		}
	}
}
