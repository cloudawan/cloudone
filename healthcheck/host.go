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
	"github.com/cloudawan/cloudone/control"
	"github.com/cloudawan/cloudone/filesystem/glusterfs"
	"github.com/cloudawan/cloudone/slb"
	"github.com/cloudawan/cloudone/utility/configuration"
	"github.com/cloudawan/cloudone/utility/database/etcd"
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
	"strings"
)

func CreateHostControl() (*HostControl, error) {
	hostControl := &HostControl{}

	dataJsonMap, err := hostControl.getDataFromEtcd()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	hostControl.dataJsonMap = dataJsonMap

	return hostControl, nil
}

type HostControl struct {
	dataJsonMap map[string]interface{}
}

func (hostControl *HostControl) getDataFromEtcd() (map[string]interface{}, error) {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return nil, err
	}

	response, err := keysAPI.Get(context.Background(), etcd.EtcdClient.EtcdBasePath+"/health", nil)
	etcdError, _ := err.(client.Error)
	if etcdError.Code == client.ErrorCodeKeyNotFound {
		return nil, etcdError
	}
	if err != nil {
		log.Error("Load status with error: %s", err)
		log.Error(response)
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
		} else {
			splitSlice := strings.Split(node.Key, "/")
			ip := splitSlice[len(splitSlice)-1]

			jsonMap[ip] = valueJsonMap
		}
	}

	if hasError {
		return jsonMap, errors.New(errorBuffer.String())
	} else {
		return jsonMap, nil
	}

	return jsonMap, nil
}

const (
	HostTypeUnknown    = "unknown"
	HostTypeKubernetes = "kubernetes"
	HostTypeGlusterfs  = "glusterfs"
	HostTypeSLB        = "slb"
)

func (hostControl *HostControl) isHostType(hostJsonMap map[string]interface{}, targetHostType string) bool {
	hostTypeSlice, ok := hostJsonMap["host_type_list"].([]interface{})
	if ok {
		for _, hostType := range hostTypeSlice {
			value, ok := hostType.(string)
			if ok {
				if value == targetHostType {
					return true
				}
			}
		}
		return false
	} else {
		if targetHostType == HostTypeUnknown {
			return true
		} else {
			return false
		}
	}
}

func (hostControl *HostControl) GetKubernetesHostStatus() (map[string]interface{}, error) {
	ipSlice, err := hostControl.GetKubernetesAllNodeIP()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	kubernetesJsonMap := make(map[string]interface{})
	for key, value := range hostControl.dataJsonMap {
		hostJsonMap, ok := value.(map[string]interface{})
		if ok {
			if hostControl.isHostType(hostJsonMap, HostTypeKubernetes) || hostControl.isHostType(hostJsonMap, HostTypeUnknown) {
				hostJsonMap["active"] = false
				kubernetesJsonMap[key] = hostJsonMap
			}
		}
	}

	for _, ip := range ipSlice {
		if kubernetesJsonMap[ip] == nil {
			kubernetesJsonMap[ip] = make(map[string]interface{})
			kubernetesJsonMap[ip].(map[string]interface{})["active"] = false
		} else {
			kubernetesJsonMap[ip].(map[string]interface{})["active"] = true
		}
	}

	return kubernetesJsonMap, nil
}

func (hostControl *HostControl) GetKubernetesAllNodeIP() ([]string, error) {
	kubeApiServerEndPoint, kubeApiServerToken, err := configuration.GetAvailablekubeApiServerEndPoint()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return control.GetAllNodeIP(kubeApiServerEndPoint, kubeApiServerToken)
}

func (hostControl *HostControl) GetGlusterfsHostStatus() (map[string]interface{}, error) {
	glusterfsClusterSlice, err := glusterfs.GetStorage().LoadAllGlusterfsCluster()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	glusterfsJsonMap := make(map[string]interface{})
	for key, value := range hostControl.dataJsonMap {
		hostJsonMap, ok := value.(map[string]interface{})
		if ok {
			if hostControl.isHostType(hostJsonMap, HostTypeGlusterfs) {
				hostJsonMap["active"] = true
				glusterfsJsonMap[key] = hostJsonMap
			}
		}
	}

	allGlusterfsClusterJsonMap := make(map[string]interface{})
	for _, glusterfsCluster := range glusterfsClusterSlice {
		glusterfsClusterJsonMap := make(map[string]interface{})
		for _, host := range glusterfsCluster.HostSlice {
			if glusterfsJsonMap[host] == nil {
				glusterfsClusterJsonMap[host] = make(map[string]interface{})
				glusterfsClusterJsonMap[host].(map[string]interface{})["active"] = false
			} else {
				glusterfsClusterJsonMap[host] = glusterfsJsonMap[host]
			}
		}
		allGlusterfsClusterJsonMap[glusterfsCluster.Name] = glusterfsClusterJsonMap
	}

	return allGlusterfsClusterJsonMap, nil
}

func (hostControl *HostControl) GetSLBHostStatus() (map[string]interface{}, error) {
	slbDaemonSlice, err := slb.GetStorage().LoadAllSLBDaemon()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	slbDaemonJsonMap := make(map[string]interface{})
	for key, value := range hostControl.dataJsonMap {
		hostJsonMap, ok := value.(map[string]interface{})
		if ok {
			if hostControl.isHostType(hostJsonMap, HostTypeSLB) {
				hostJsonMap["active"] = true
				slbDaemonJsonMap[key] = hostJsonMap
			}
		}
	}

	allSLBDaemonSetJsonMap := make(map[string]interface{})
	for _, slbDaemon := range slbDaemonSlice {
		slbDaemonSetJsonMap := make(map[string]interface{})
		for _, endPoint := range slbDaemon.EndPointSlice {
			host, err := parseHostFromEndpoint(endPoint, slbDaemon)
			if err != nil {
				log.Error(err)
				return nil, err
			}

			if slbDaemonJsonMap[host] == nil {
				slbDaemonSetJsonMap[host] = make(map[string]interface{})
				slbDaemonSetJsonMap[host].(map[string]interface{})["active"] = false
			} else {
				slbDaemonSetJsonMap[host] = slbDaemonJsonMap[host]
			}
		}
		allSLBDaemonSetJsonMap[slbDaemon.Name] = slbDaemonSetJsonMap
	}

	return allSLBDaemonSetJsonMap, nil
}

func parseHostFromEndpoint(endPoint string, slbDaemon slb.SLBDaemon) (string, error) {
	splitSlice := strings.Split(endPoint, "//")
	if len(splitSlice) != 2 {
		errMessage := fmt.Sprintf("Parse endpoint %s error slbDaemon %v", endPoint, slbDaemon)
		log.Error(errMessage)
		return "", errors.New(errMessage)
	}
	splitSlice = strings.Split(splitSlice[1], ":")
	if len(splitSlice) != 2 {
		errMessage := fmt.Sprintf("Parse endpoint %s error slbDaemon %v", endPoint, slbDaemon)
		log.Error(errMessage)
		return "", errors.New(errMessage)
	}
	return splitSlice[0], nil
}
