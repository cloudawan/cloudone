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

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/cloudawan/cloudone/monitor"
	"strconv"
	"time"
)

type NotifierSerializable struct {
	Kind string
	Data string
}

type ReplicationControllerNotifierSerializable struct {
	Check                 bool
	CoolDownDuration      int64
	RemainingCoolDown     int64
	KubeApiServerEndPoint string
	KubeApiServerToken    string
	Namespace             string
	Kind                  string
	Name                  string
	NotifierSlice         []NotifierSerializable
	IndicatorSlice        []Indicator
}

type Notifier interface {
	notify(message string) error
}

type ReplicationControllerNotifier struct {
	Check                 bool
	CoolDownDuration      time.Duration
	RemainingCoolDown     time.Duration
	KubeApiServerEndPoint string
	KubeApiServerToken    string
	Namespace             string
	Kind                  string
	Name                  string
	NotifierSlice         []Notifier
	IndicatorSlice        []Indicator
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

func CheckAndExecuteNotifier(replicationControllerNotifier *ReplicationControllerNotifier) (bool, error) {
	switch replicationControllerNotifier.Kind {
	case "selector":
		nameSlice, err := monitor.GetReplicationControllerNameFromSelector(
			replicationControllerNotifier.KubeApiServerEndPoint,
			replicationControllerNotifier.KubeApiServerToken,
			replicationControllerNotifier.Namespace,
			replicationControllerNotifier.Name)
		if err != nil {
			return false, errors.New("Could not find replication controller name with selector " + replicationControllerNotifier.Name + " error " + err.Error())
		} else {
			atLeastOneNotify := false
			errorMessage := bytes.Buffer{}
			hasError := false
			for _, name := range nameSlice {
				result, err := CheckAndExecuteNotifierOnReplicationController(replicationControllerNotifier, name)
				atLeastOneNotify = atLeastOneNotify || result
				if err != nil {
					errorMessage.WriteString(err.Error())
					hasError = true
				}
			}

			if hasError == false {
				return atLeastOneNotify, nil
			} else {
				return atLeastOneNotify, errors.New(errorMessage.String())
			}
		}
	case "replicationController":
		return CheckAndExecuteNotifierOnReplicationController(replicationControllerNotifier, replicationControllerNotifier.Name)
	default:
		return false, errors.New("No such kind " + replicationControllerNotifier.Kind)
	}
}

func CheckAndExecuteNotifierOnReplicationController(replicationControllerNotifier *ReplicationControllerNotifier, replicationControllerName string) (bool, error) {
	replicationControllerMetric, err := monitor.MonitorReplicationController(replicationControllerNotifier.KubeApiServerEndPoint, replicationControllerNotifier.KubeApiServerToken, replicationControllerNotifier.Namespace, replicationControllerName)
	if err != nil {
		log.Error("Get ReplicationController data failure: %s where replicationControllerNotifier %s", err.Error(), replicationControllerNotifier)
		return false, err
	}

	message := bytes.Buffer{}
	message.WriteString("Replication Controller: " + replicationControllerName + "\n")
	toNotify := false
	for _, indicator := range replicationControllerNotifier.IndicatorSlice {
		toNotifyAbove := monitor.CheckThresholdReplicationController(indicator.Type, true, indicator.AboveAllOrOne, replicationControllerMetric, indicator.AbovePercentageOfData, indicator.AboveThreshold)
		if toNotifyAbove {
			message.WriteString(generateMessage(indicator.Type, true, indicator.AboveAllOrOne, indicator.AbovePercentageOfData, indicator.AboveThreshold))
		}
		toNotifyBelow := monitor.CheckThresholdReplicationController(indicator.Type, false, indicator.BelowAllOrOne, replicationControllerMetric, indicator.BelowPercentageOfData, indicator.BelowThreshold)
		if toNotifyBelow {
			message.WriteString(generateMessage(indicator.Type, false, indicator.BelowAllOrOne, indicator.BelowPercentageOfData, indicator.BelowThreshold))
		}
		toNotify = toNotify || toNotifyAbove || toNotifyBelow
	}

	errorBuffer := bytes.Buffer{}
	if toNotify {
		for _, notifier := range replicationControllerNotifier.NotifierSlice {
			err := notifier.notify(message.String())
			if err != nil {
				errorBuffer.WriteString(err.Error())
			}
		}
	}

	if errorBuffer.Len() > 0 {
		return toNotify, errors.New(errorBuffer.String())
	} else {
		return toNotify, nil
	}
}

func generateMessage(indicator string, aboveOrBelow bool, allOrOne bool, percentageOfData float64, threshold int64) string {
	message := bytes.Buffer{}
	message.WriteString("For the indicator " + indicator)
	if allOrOne {
		message.WriteString(", all containers were ")
	} else {
		message.WriteString(", at least one container was ")
	}
	if aboveOrBelow {
		message.WriteString("above the threshold " + strconv.Itoa(int(threshold)))
	} else {
		message.WriteString("below the threshold " + strconv.Itoa(int(threshold)))
	}
	message.WriteString(" for more than " + strconv.Itoa(int(percentageOfData*100)) + "% of the obeserverd duration (1 minute).\n")
	return message.String()
}

func ConvertToSerializable(replicationControllerNotifier ReplicationControllerNotifier) (ReplicationControllerNotifierSerializable, error) {
	var err error = nil
	returnedNotifierSlice := make([]NotifierSerializable, 0)
	for _, notifier := range replicationControllerNotifier.NotifierSlice {
		switch notifier.(type) {
		case NotifierEmail:
			byteSlice, err := json.Marshal(notifier.(NotifierEmail))
			if err == nil {
				returnedNotifierSlice = append(returnedNotifierSlice, NotifierSerializable{"email", string(byteSlice)})
			}
		case NotifierSMSNexmo:
			byteSlice, err := json.Marshal(notifier.(NotifierSMSNexmo))
			if err == nil {
				returnedNotifierSlice = append(returnedNotifierSlice, NotifierSerializable{"smsNexmo", string(byteSlice)})
			}
		default:
			err = errors.New("No such kind")
		}
	}

	returnedReplicationControllerNotifier := ReplicationControllerNotifierSerializable{
		replicationControllerNotifier.Check,
		int64(replicationControllerNotifier.CoolDownDuration),
		int64(replicationControllerNotifier.RemainingCoolDown),
		replicationControllerNotifier.KubeApiServerEndPoint,
		replicationControllerNotifier.KubeApiServerToken,
		replicationControllerNotifier.Namespace,
		replicationControllerNotifier.Kind,
		replicationControllerNotifier.Name,
		returnedNotifierSlice,
		replicationControllerNotifier.IndicatorSlice,
	}

	return returnedReplicationControllerNotifier, err
}

func ConvertFromSerializable(replicationControllerNotifierSerializable ReplicationControllerNotifierSerializable) (ReplicationControllerNotifier, error) {
	var err error = nil
	returnedNotifierSlice := make([]Notifier, 0)
	for _, notifier := range replicationControllerNotifierSerializable.NotifierSlice {
		switch notifier.Kind {
		case "email":
			notifierEmail := NotifierEmail{}
			err := json.Unmarshal([]byte(notifier.Data), &notifierEmail)
			if err == nil {
				returnedNotifierSlice = append(returnedNotifierSlice, notifierEmail)
			}
		case "smsNexmo":
			notifierSMSNexmo := NotifierSMSNexmo{}
			err := json.Unmarshal([]byte(notifier.Data), &notifierSMSNexmo)
			if err == nil {
				returnedNotifierSlice = append(returnedNotifierSlice, notifierSMSNexmo)
			}
		default:
			err = errors.New("No such kind " + notifier.Kind)
		}
	}

	replicationControllerNotifier := ReplicationControllerNotifier{
		replicationControllerNotifierSerializable.Check,
		time.Duration(replicationControllerNotifierSerializable.CoolDownDuration),
		time.Duration(replicationControllerNotifierSerializable.RemainingCoolDown),
		replicationControllerNotifierSerializable.KubeApiServerEndPoint,
		replicationControllerNotifierSerializable.KubeApiServerToken,
		replicationControllerNotifierSerializable.Namespace,
		replicationControllerNotifierSerializable.Kind,
		replicationControllerNotifierSerializable.Name,
		returnedNotifierSlice,
		replicationControllerNotifierSerializable.IndicatorSlice,
	}

	return replicationControllerNotifier, err
}
