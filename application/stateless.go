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

package application

import (
	"encoding/json"
	"github.com/cloudawan/cloudone/control"
)

type Stateless struct {
	Name                      string
	Description               string
	replicationControllerJson []byte
	serviceJson               []byte
	Environment               map[string]string
}

type StatelessSerializable struct {
	Name                      string
	Description               string
	ReplicationControllerJson map[string]interface{}
	ServiceJson               map[string]interface{}
	Environment               map[string]string
}

func StoreStatelessApplication(name string, description string,
	replicationController map[string]interface{}, service map[string]interface{}, environment map[string]string) error {
	replicationControllerByteSlice, err := json.Marshal(replicationController)
	if err != nil {
		log.Error("Marshal replication controller json for stateless application error %s", err)
		return err
	}
	serviceByteSlice, err := json.Marshal(service)
	if err != nil {
		log.Error("Marshal service json for stateless application error %s", err)
		return err
	}
	err = saveStatelessApplication(&Stateless{name, description, replicationControllerByteSlice, serviceByteSlice, environment})
	if err != nil {
		log.Error("Store statelss application error %s", err)
		return err
	}
	return nil
}

func RetrieveStatelessApplication(name string) (*StatelessSerializable, error) {
	stateless, err := LoadStatelessApplication(name)
	if err != nil {
		log.Error("Load stateless application error %s", err)
		return nil, err
	}

	replicationControllerJsonMap := make(map[string]interface{})
	err = json.Unmarshal(stateless.replicationControllerJson, &replicationControllerJsonMap)
	if err != nil {
		log.Error("Unmarshal replication controller json for stateless application error %s", err)
		return nil, err
	}
	serviceJsonMap := make(map[string]interface{})
	err = json.Unmarshal(stateless.serviceJson, &serviceJsonMap)
	if err != nil {
		log.Error("Unmarshal service json for stateless application error %s", err)
		return nil, err
	}

	statelessSerializable := &StatelessSerializable{
		stateless.Name,
		stateless.Description,
		replicationControllerJsonMap,
		serviceJsonMap,
		stateless.Environment,
	}

	return statelessSerializable, nil
}

func LaunchStatelessApplication(kubeapiHost string, kubeapiPort int, namespace string, name string, environmentSlice []interface{}) error {
	stateless, err := LoadStatelessApplication(name)
	if err != nil {
		log.Error("Load stateless application error %s", err)
		return err
	}

	replicationControllerJsonMap := make(map[string]interface{})
	err = json.Unmarshal(stateless.replicationControllerJson, &replicationControllerJsonMap)
	if err != nil {
		log.Error("Unmarshal replication controller json for stateless application error %s", err)
		return err
	}
	serviceJsonMap := make(map[string]interface{})
	err = json.Unmarshal(stateless.serviceJson, &serviceJsonMap)
	if err != nil {
		log.Error("Unmarshal service json for stateless application error %s", err)
		return err
	}

	// Add environment variable
	if environmentSlice != nil {
		containerSlice := replicationControllerJsonMap["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{})["containers"].([]interface{})
		for i := 0; i < len(containerSlice); i++ {
			_, ok := containerSlice[i].(map[string]interface{})["env"].([]interface{})
			if ok {
				for _, environment := range environmentSlice {
					containerSlice[i].(map[string]interface{})["env"] = append(containerSlice[i].(map[string]interface{})["env"].([]interface{}), environment)
				}
			} else {
				containerSlice[i].(map[string]interface{})["env"] = environmentSlice
			}
		}
	}

	err = control.CreateReplicationControllerWithJson(kubeapiHost, kubeapiPort, namespace, replicationControllerJsonMap)
	if err != nil {
		log.Error("CreateReplicationControllerWithJson error %s", err)
		return err
	}
	err = control.CreateServiceWithJson(kubeapiHost, kubeapiPort, namespace, serviceJsonMap)
	if err != nil {
		log.Error("CreateReplicationControllerWithJson error %s", err)
		return err
	}
	return nil
}
