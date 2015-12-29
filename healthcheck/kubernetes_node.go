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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cloudawan/cloudone/utility/configuration"
	"github.com/cloudawan/cloudone_utility/restclient"
	"time"
)

func CreateKubernetesNodeControl() (*KubernetesNodeControl, error) {
	etcdHostAndPortSlice, ok := configuration.LocalConfiguration.GetStringSlice("etcdHostAndPort")
	if ok == false {
		log.Error("Can't load etcdHostAndPortSlice")
		return nil, errors.New("Can't load etcdHostAndPortSlice")
	}

	etcdCheckTimeoutInMilliSecond, ok := configuration.LocalConfiguration.GetInt("etcdCheckTimeoutInMilliSecond")
	if ok == false {
		log.Error("Can't load etcdCheckTimeoutInMilliSecond")
		return nil, errors.New("Can't load etcdCheckTimeoutInMilliSecond")
	}

	kubernetesNodeControl := &KubernetesNodeControl{
		etcdHostAndPortSlice,
		etcdCheckTimeoutInMilliSecond,
	}

	return kubernetesNodeControl, nil
}

type KubernetesNodeControl struct {
	EtcdHostAndPortSlice          []string
	EtcdCheckTimeoutInMilliSecond int
}

func (kubernetesNodeControl *KubernetesNodeControl) getAvailableHostAndPort() (*string, error) {
	for _, hostAndPort := range kubernetesNodeControl.EtcdHostAndPortSlice {
		result, err := restclient.HealthCheck("http://"+hostAndPort+"/v2/keys",
			time.Duration(kubernetesNodeControl.EtcdCheckTimeoutInMilliSecond)*time.Millisecond)
		if result {
			return &hostAndPort, nil
		} else if err != nil {
			log.Error(err)
		}
	}

	return nil, errors.New("No available host and port")
}

func (kubernetesNodeControl *KubernetesNodeControl) GetStatus() (map[string]interface{}, error) {
	hostAndPort, err := kubernetesNodeControl.getAvailableHostAndPort()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	result, err := restclient.RequestGet("http://"+*hostAndPort+"/v2/keys/cloudawan/cloudone/health", false)
	jsonMap, ok := result.(map[string]interface{})
	if ok {
		return jsonMap, nil
	} else {
		text := fmt.Sprintf("Fail to convert result: %v", result)
		log.Error(text)
		return nil, errors.New(text)
	}
}

func (kubernetesNodeControl *KubernetesNodeControl) GetHostWithinFlannelNetwork() ([]string, error) {
	hostAndPort, err := kubernetesNodeControl.getAvailableHostAndPort()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	result, err := restclient.RequestGet("http://"+*hostAndPort+"/v2/keys/coreos.com/network/subnets", false)
	jsonMap, ok := result.(map[string]interface{})
	if ok {
		nodeJsonMap, ok := jsonMap["node"].(map[string]interface{})
		if ok == false {
			text := fmt.Sprintf("Fail to convert jsonMap[node]: %v, jsonMap: %v", jsonMap["node"], jsonMap)
			log.Error(text)
			return nil, errors.New(text)
		}
		nodeSlice, _ := nodeJsonMap["nodes"].([]interface{})
		if ok == false {
			text := fmt.Sprintf("Fail to convert nodeJsonMap[nodes]: %v, jsonMap: %v", nodeJsonMap["nodes"], jsonMap)
			log.Error(text)
			return nil, errors.New(text)
		}

		ipSlice := make([]string, 0)
		for _, node := range nodeSlice {
			data, ok := node.(map[string]interface{})
			if ok == false {
				text := fmt.Sprintf("Fail to convert node: %v, jsonMap: %v", node, jsonMap)
				log.Error(text)
				return nil, errors.New(text)
			}
			value, ok := data["value"].(string)
			if ok == false {
				text := fmt.Sprintf("Fail to convert data[value]: %v jsonMap: %v", data["value"], jsonMap)
				log.Error(text)
				return nil, errors.New(text)
			}
			valueJsonMap := make(map[string]interface{})
			err := json.Unmarshal([]byte(value), &valueJsonMap)
			if err != nil {
				log.Error(err)
				return nil, err
			}
			ip, ok := valueJsonMap["PublicIP"].(string)
			if ok == false {
				text := fmt.Sprintf("Fail to convert valueJsonMap[PublicIP]: %v jsonMap: %v", valueJsonMap["PublicIP"], jsonMap)
				log.Error(text)
				return nil, errors.New(text)
			}
			ipSlice = append(ipSlice, ip)
		}
		return ipSlice, nil
	} else {
		text := fmt.Sprintf("Fail to convert result: %v", result)
		log.Error(text)
		return nil, errors.New(text)
	}
}
