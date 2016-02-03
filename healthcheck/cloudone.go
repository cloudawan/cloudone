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

package healthcheck

import (
	"errors"
	"github.com/cloudawan/cloudone/utility/configuration"
	"github.com/cloudawan/cloudone_utility/restclient"
	"os/exec"
	"strconv"
	"time"
)

func CreateCloudoneControl() (*CloudoneControl, error) {
	restapiPort, ok := configuration.LocalConfiguration.GetInt("restapiPort")
	if ok == false {
		log.Error("Can't find restapiPort")
		return nil, errors.New("Can't find restapiPort")
	}
	cloudoneControl := &CloudoneControl{
		restapiPort,
	}
	return cloudoneControl, nil
}

type CloudoneControl struct {
	RestapiPort int
}

func (cloudoneControl *CloudoneControl) testRestAPI() bool {
	result, _ := restclient.HealthCheck(
		"https://127.0.0.1:"+strconv.Itoa(cloudoneControl.RestapiPort)+"/apidocs.json",
		time.Millisecond*300)
	return result
}

func (cloudoneControl *CloudoneControl) testStorage() bool {
	if err := GetStorage().saveTest("test", time.Now()); err != nil {
		log.Error(err)
		return false
	} else {
		return true
	}
}

func (cloudoneControl *CloudoneControl) testDocker() bool {
	command := exec.Command("docker", "ps")
	_, err := command.CombinedOutput()
	if err != nil {
		log.Error(err)
		return false
	} else {
		return true
	}
}

func (cloudoneControl *CloudoneControl) GetStatus() map[string]interface{} {
	jsonMap := make(map[string]interface{})
	jsonMap["restapi"] = cloudoneControl.testRestAPI()
	jsonMap["storage"] = cloudoneControl.testStorage()
	jsonMap["docker"] = cloudoneControl.testDocker()
	return jsonMap
}
