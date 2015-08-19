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
	"github.com/cloudawan/kubernetes_management/notification"
	"time"
)

var registerNotifierChannel = make(chan notification.ReplicationControllerNotifier)

var replicationControllerNotifierMap = make(map[string]*notification.ReplicationControllerNotifier)

func init() {
	// Load from database
	replicationControllerNotifierSerializableSlice, err := notification.LoadAllReplicationControllerNotifierSerializable()
	if err != nil {
		log.Error(err)
	} else {
		for _, replicationControllerNotifierSerializable := range replicationControllerNotifierSerializableSlice {
			replicationControllerNotifier, err := notification.ConvertFromSerializable(replicationControllerNotifierSerializable)
			if err != nil {
				log.Error(err)
			} else {
				AddReplicationControllerNotifier(&replicationControllerNotifier)
			}
		}
	}
}

func loopNotifier(ticker *time.Ticker, checkingInterval time.Duration) {
	for {
		select {
		// Notifier
		case replicationControllerNotifier := <-registerNotifierChannel:
			receiveFromNotifierChannel(replicationControllerNotifier)
		case <-ticker.C:
			// Notifier
			periodicalCheckNotifier(checkingInterval)
		case <-quitChannel:
			ticker.Stop()
			close(registerNotifierChannel)
			log.Info("Loop notifier quit")
			return
		}
	}
}

func GetReplicationControllerNotifierMap() map[string]notification.ReplicationControllerNotifier {
	// Return a copy rather than a original one to prevent from concurrency issue
	// This read is not guaranteed for consistency between multiple goroutines. The data may be old but not corrupted.
	returnedReplicationControllerNotifierMap := make(map[string]notification.ReplicationControllerNotifier)
	for key, value := range replicationControllerNotifierMap {
		returnedReplicationControllerNotifierMap[key] = *value
	}
	return returnedReplicationControllerNotifierMap
}

func GetReplicationControllerNotifier(namespace string, kind string, name string) (bool, notification.ReplicationControllerNotifier) {
	// Return a copy rather than a original one to prevent from concurrency issue.
	// This read is not guaranteed for consistency between multiple goroutines. The data may be old but not corrupted.
	returnedReplicationControllerNotifier := replicationControllerNotifierMap[getKeyForReplicationControllerNotifierMap(namespace, kind, name)]
	if returnedReplicationControllerNotifier != nil {
		return true, *returnedReplicationControllerNotifier
	} else {
		return false, notification.ReplicationControllerNotifier{}
	}
}

func AddReplicationControllerNotifier(replicationControllerNotifier *notification.ReplicationControllerNotifier) {
	registerNotifierChannel <- *replicationControllerNotifier
}

func getKeyForReplicationControllerNotifierMap(namespace string, kind string, name string) string {
	return namespace + "/" + kind + "/" + name
}

func receiveFromNotifierChannel(replicationControllerNotifier notification.ReplicationControllerNotifier) {
	id := getKeyForReplicationControllerNotifierMap(replicationControllerNotifier.Namespace, replicationControllerNotifier.Kind, replicationControllerNotifier.Name)
	if replicationControllerNotifier.Check {
		replicationControllerNotifierMap[id] = &replicationControllerNotifier
	} else {
		delete(replicationControllerNotifierMap, id)
	}
}

func periodicalCheckNotifier(checkingInterval time.Duration) {
	for _, replicationControllerNotifier := range replicationControllerNotifierMap {
		if replicationControllerNotifier.RemainingCoolDown > 0 {
			replicationControllerNotifier.RemainingCoolDown -= checkingInterval
		}
		if replicationControllerNotifier.RemainingCoolDown <= 0*time.Second {
			toNotify, err := notification.CheckAndExecuteNotifier(replicationControllerNotifier)
			if err != nil {
				log.Error("CheckAndExecuteNotifier error: %s where ReplicationControllerNotifier %s", err.Error(), replicationControllerNotifier)
			}
			if toNotify {
				replicationControllerNotifier.RemainingCoolDown = replicationControllerNotifier.CoolDownDuration
				log.Info("CheckAndExecuteNotifier notified where ReplicationControllerNotifier %s", replicationControllerNotifier)
			}
		}
	}
}
