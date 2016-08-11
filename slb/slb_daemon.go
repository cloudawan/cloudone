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

package slb

import (
	"bytes"
	"errors"
	"github.com/cloudawan/cloudone_utility/restclient"
	"github.com/cloudawan/cloudone_utility/slb"
)

type SLBDaemon struct {
	Name          string
	EndPointSlice []string
	NodeHostSlice []string
	Description   string
}

func (slbDaemon *SLBDaemon) SendCommand(command *slb.Command) error {
	command.NodeHostSlice = slbDaemon.NodeHostSlice

	buffer := bytes.Buffer{}
	for _, endPoint := range slbDaemon.EndPointSlice {
		url := endPoint + "/api/v1/slb"
		_, err := restclient.RequestPut(url, command, nil, false)
		if err != nil {
			log.Error(err)
			buffer.WriteString("Fail to configure " + endPoint + " with error " + err.Error() + "\n")
		}
	}

	if buffer.Len() > 0 {
		return errors.New(buffer.String())
	} else {
		return nil
	}
}

func SendCommandToAllSLBDaemon() error {
	slbDaemonSlice, err := GetStorage().LoadAllSLBDaemon()
	if err != nil {
		log.Error(err)
		return nil
	}

	command, err := CreateCommand()
	if err != nil {
		log.Error(err)
		return nil
	}

	buffer := bytes.Buffer{}
	for _, slbDaemon := range slbDaemonSlice {
		err := slbDaemon.SendCommand(command)
		if err != nil {
			log.Error(err)
			buffer.WriteString(err.Error())
		}
	}

	if buffer.Len() > 0 {
		return errors.New(buffer.String())
	} else {
		return nil
	}
}

// Used for failed slb host to reconfigure
func SendCommandToSLBDaemon(targetEndPoint string) error {
	slbDaemonSlice, err := GetStorage().LoadAllSLBDaemon()
	if err != nil {
		log.Error(err)
		return nil
	}

	command, err := CreateCommand()
	if err != nil {
		log.Error(err)
		return nil
	}

	for _, slbDaemon := range slbDaemonSlice {
		for _, endPoint := range slbDaemon.EndPointSlice {
			if endPoint == targetEndPoint {
				command.NodeHostSlice = slbDaemon.NodeHostSlice
				url := endPoint + "/api/v1/slb"
				_, err := restclient.RequestPut(url, command, nil, false)
				if err != nil {
					log.Error(err)
					return errors.New("Fail to configure " + endPoint + "with error " + err.Error() + "\n")
				} else {
					return nil
				}
			}
		}
	}

	return errors.New("Can't find endpoint " + targetEndPoint)
}
