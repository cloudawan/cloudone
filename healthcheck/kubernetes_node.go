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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cloudawan/cloudone/utility/database/etcd"
	"golang.org/x/net/context"
	"strings"
)

func CreateKubernetesNodeControl() (*KubernetesNodeControl, error) {
	kubernetesNodeControl := &KubernetesNodeControl{}

	return kubernetesNodeControl, nil
}

type KubernetesNodeControl struct {
}

func (kubernetesNodeControl *KubernetesNodeControl) GetStatus() (map[string]interface{}, error) {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()

	response, err := keysAPI.Get(context.Background(), "/cloudawan/cloudone/health", nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	if response.Node == nil {
		text := fmt.Sprintf("Response node is nil. Response: %v", response)
		log.Error(text)
		return nil, errors.New(text)
	}

	hasError := false
	errorBuffer := bytes.Buffer{}
	jsonMap := make(map[string]interface{})
	for _, node := range response.Node.Nodes {
		valueJsonMap := make(map[string]interface{})
		err := json.Unmarshal([]byte(node.Value), &valueJsonMap)
		if err != nil {
			hasError = true
			errorBuffer.WriteString(err.Error())
			log.Error(err)
		}

		splitSlice := strings.Split(node.Key, "/")
		ip := splitSlice[len(splitSlice)-1]

		jsonMap[ip] = valueJsonMap
	}

	if hasError {
		return jsonMap, errors.New(errorBuffer.String())
	} else {
		return jsonMap, nil
	}
}

func (kubernetesNodeControl *KubernetesNodeControl) GetHostWithinFlannelNetwork() ([]string, error) {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()

	response, err := keysAPI.Get(context.Background(), "/coreos.com/network/subnets", nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	if response.Node == nil {
		text := fmt.Sprintf("Response node is nil. Response: %v", response)
		log.Error(text)
		return nil, errors.New(text)
	}

	hasError := false
	errorBuffer := bytes.Buffer{}
	ipSlice := make([]string, 0)
	for _, node := range response.Node.Nodes {
		valueJsonMap := make(map[string]interface{})
		err := json.Unmarshal([]byte(node.Value), &valueJsonMap)
		if err != nil {
			hasError = true
			errorBuffer.WriteString(err.Error())
			log.Error(err)
		}

		ip, ok := valueJsonMap["PublicIP"].(string)
		if ok {
			ipSlice = append(ipSlice, ip)
		} else {
			hasError = true
			text := fmt.Sprintf("Fail to convert valueJsonMap[PublicIP]: %v valueJsonMap: %v", valueJsonMap["PublicIP"], valueJsonMap)
			errorBuffer.WriteString(text)
			log.Error(text)
		}
	}

	if hasError {
		return ipSlice, errors.New(errorBuffer.String())
	} else {
		return ipSlice, nil
	}
}
