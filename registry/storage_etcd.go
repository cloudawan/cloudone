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

package registry

import (
	"encoding/json"
	"github.com/cloudawan/cloudone/utility/database/etcd"
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

type StorageEtcd struct {
}

func (storageEtcd *StorageEtcd) initialize() error {
	if err := etcd.EtcdClient.CreateDirectoryIfNotExist(etcd.EtcdClient.EtcdBasePath + "/private_registry"); err != nil {
		log.Error("Create if not existing private registry directory error: %s", err)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) DeletePrivateRegistry(name string) error {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return err
	}

	response, err := keysAPI.Delete(context.Background(), etcd.EtcdClient.EtcdBasePath+"/private_registry/"+name, nil)
	etcdError, _ := err.(client.Error)
	if etcdError.Code == client.ErrorCodeKeyNotFound {
		log.Debug(err)
		log.Debug(response)
		return nil
	}
	if err != nil {
		log.Error("Delete private registry with name %s error: %s", name, err)
		log.Error(response)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) SavePrivateRegistry(privateRegistry *PrivateRegistry) error {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return err
	}

	byteSlice, err := json.Marshal(privateRegistry)
	if err != nil {
		log.Error("Marshal private registry %v error %s", privateRegistry, err)
		return err
	}

	response, err := keysAPI.Set(context.Background(), etcd.EtcdClient.EtcdBasePath+"/private_registry/"+privateRegistry.Name, string(byteSlice), nil)
	if err != nil {
		log.Error("Save private registry %v error: %s", privateRegistry, err)
		log.Error(response)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) LoadPrivateRegistry(name string) (*PrivateRegistry, error) {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return nil, err
	}

	response, err := keysAPI.Get(context.Background(), etcd.EtcdClient.EtcdBasePath+"/private_registry/"+name, nil)
	etcdError, _ := err.(client.Error)
	if etcdError.Code == client.ErrorCodeKeyNotFound {
		return nil, etcdError
	}
	if err != nil {
		log.Error("Load private registry with name %s error: %s", name, err)
		log.Error(response)
		return nil, err
	}

	privateRegistry := new(PrivateRegistry)
	err = json.Unmarshal([]byte(response.Node.Value), &privateRegistry)
	if err != nil {
		log.Error("Unmarshal private registry %v error %s", response.Node.Value, err)
		return nil, err
	}

	return privateRegistry, nil
}

func (storageEtcd *StorageEtcd) LoadAllPrivateRegistry() ([]PrivateRegistry, error) {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return nil, err
	}

	response, err := keysAPI.Get(context.Background(), etcd.EtcdClient.EtcdBasePath+"/private_registry", nil)
	if err != nil {
		log.Error("Load all private registry error: %s", err)
		log.Error(response)
		return nil, err
	}

	privateRegistrySlice := make([]PrivateRegistry, 0)
	for _, node := range response.Node.Nodes {
		privateRegistry := PrivateRegistry{}
		err := json.Unmarshal([]byte(node.Value), &privateRegistry)
		if err != nil {
			log.Error("Unmarshal private registry %v error %s", node.Value, err)
			return nil, err
		}
		privateRegistrySlice = append(privateRegistrySlice, privateRegistry)
	}

	return privateRegistrySlice, nil
}
