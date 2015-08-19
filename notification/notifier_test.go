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

package notification

/*
import (
	"fmt"
	"testing"
	"time"
)

func TestCheckAndExecuteNotifier(t *testing.T) {
	indicatorSlice := make([]Indicator, 1)
	indicatorSlice[0] = Indicator{"cpu", false, 0.3, 800000000, false, 0.3, 1000000}
	notifierSlice := make([]Notifier, 1)
	notifierSlice[0] = NotifierEmail{[]string{"cloudawanemailtest@gmail.com"}}
	fmt.Println(CheckAndExecuteNotifier(&ReplicationControllerNotifier{
		true, 10 * time.Second, 0 * time.Second, "192.168.0.33", 8080,
		"default", "replicationController", "flask", notifierSlice, indicatorSlice}))
}

func TestConvertToSerializable(t *testing.T) {
	indicatorSlice := make([]Indicator, 1)
	indicatorSlice[0] = Indicator{"cpu", false, 0.3, 800000000, false, 0.3, 1000000}
	notifierSlice := make([]Notifier, 1)
	notifierSlice[0] = NotifierSMSNexmo{"cloudawan", []string{"1234567"}}
	serialized, _ := ConvertToSerializable(ReplicationControllerNotifier{
		true, 10 * time.Second, 0 * time.Second, "192.168.0.33", 8080,
		"default", "replicationController", "flask", notifierSlice, indicatorSlice})
	data, _ := ConvertFromSerializable(serialized)
	fmt.Println(data.NotifierSlice[0])
}
*/
