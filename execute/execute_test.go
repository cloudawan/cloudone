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

/*
import (
	"github.com/cloudawan/kubernetes_management/autoscaler"
	"github.com/cloudawan/kubernetes_management/monitor"
	"github.com/cloudawan/kubernetes_management/notification"
	"testing"
	"time"
)

func TestAutoScaler(t *testing.T) {
	indicatorSlice := make([]autoscaler.Indicator, 1)
	indicatorSlice[0] = autoscaler.Indicator{monitor.CPU, false, 0.3, 800000000, true, 0.3, 5000000}
	replicationControllerAutoScaler := &autoscaler.ReplicationControllerAutoScaler{true, 10 * time.Second, 0 * time.Second, "192.168.0.33", 8080, "default", "flask", 3, 1, indicatorSlice}
	AddReplicationControllerAutoScaler(replicationControllerAutoScaler)
	time.Sleep(100 * time.Second)
	replicationControllerAutoScaler.Check = false
	AddReplicationControllerAutoScaler(replicationControllerAutoScaler)
	time.Sleep(10 * time.Second)
	Close()
}

func TestAutoScaler(t *testing.T) {
	indicatorSlice := make([]notification.Indicator, 1)
	indicatorSlice[0] = notification.Indicator{monitor.CPU, false, 0.3, 800000000, true, 0.3, 5000000}
	notifierSlice := make([]notification.Notifier, 1)
	notifierSlice[0] = notification.NotifierEmail{[]string{"cloudawanemailtest@gmail.com"}}
	replicationControllerNotifier := &notification.ReplicationControllerNotifier{true, 10 * time.Second, 0 * time.Second, "192.168.0.33", 8080, "default", "flask", notifierSlice, indicatorSlice}
	AddReplicationControllerNotifier(replicationControllerNotifier)
	time.Sleep(100 * time.Second)
	replicationControllerNotifier.Check = false
	AddReplicationControllerNotifier(replicationControllerNotifier)
	time.Sleep(10 * time.Second)
	Close()
}
*/
