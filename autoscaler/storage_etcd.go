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

package autoscaler

import (
	"encoding/json"
	"github.com/cloudawan/cloudone/utility/database/etcd"
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

type StorageEtcd struct {
}

func (storageEtcd *StorageEtcd) initialize() error {
	if err := etcd.EtcdClient.CreateDirectoryIfNotExist(etcd.EtcdClient.EtcdBasePath + "/auto_scaler"); err != nil {
		log.Error("Create if not existing auto scaler directory error: %s", err)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) getKeyAutoScaler(namespace string, kind string, name string) string {
	return namespace + "." + kind + "." + name
}

func (storageEtcd *StorageEtcd) DeleteReplicationControllerAutoScaler(namespace string, kind string, name string) error {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return err
	}

	key := storageEtcd.getKeyAutoScaler(namespace, kind, name)
	response, err := keysAPI.Delete(context.Background(), etcd.EtcdClient.EtcdBasePath+"/auto_scaler/"+key, nil)
	if err != nil {
		log.Error("Delete auto scaler with namespace %s kind %s name %s error: %s", namespace, kind, name, err)
		log.Error(response)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) SaveReplicationControllerAutoScaler(replicationControllerAutoScaler *ReplicationControllerAutoScaler) error {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return err
	}

	byteSlice, err := json.Marshal(replicationControllerAutoScaler)
	if err != nil {
		log.Error("Marshal auto scaler %v error %s", replicationControllerAutoScaler, err)
		return err
	}

	key := storageEtcd.getKeyAutoScaler(replicationControllerAutoScaler.Namespace, replicationControllerAutoScaler.Kind, replicationControllerAutoScaler.Name)
	response, err := keysAPI.Set(context.Background(), etcd.EtcdClient.EtcdBasePath+"/auto_scaler/"+key, string(byteSlice), nil)
	if err != nil {
		log.Error("Save auto scaler %v error: %s", replicationControllerAutoScaler, err)
		log.Error(response)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) LoadReplicationControllerAutoScaler(namespace string, kind string, name string) (*ReplicationControllerAutoScaler, error) {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return nil, err
	}

	key := storageEtcd.getKeyAutoScaler(namespace, kind, name)
	response, err := keysAPI.Get(context.Background(), etcd.EtcdClient.EtcdBasePath+"/auto_scaler/"+key, nil)
	etcdError, _ := err.(client.Error)
	if etcdError.Code == client.ErrorCodeKeyNotFound {
		return nil, etcdError
	}
	if err != nil {
		log.Error("Load auto scaler with namespace %s kind %s name %s error: %s", namespace, kind, name, err)
		log.Error(response)
		return nil, err
	}

	replicationControllerAutoScaler := new(ReplicationControllerAutoScaler)
	err = json.Unmarshal([]byte(response.Node.Value), &replicationControllerAutoScaler)
	if err != nil {
		log.Error("Unmarshal auto scaler %v error %s", response.Node.Value, err)
		return nil, err
	}

	return replicationControllerAutoScaler, nil
}

func (storageEtcd *StorageEtcd) LoadAllReplicationControllerAutoScaler() ([]ReplicationControllerAutoScaler, error) {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return nil, err
	}

	response, err := keysAPI.Get(context.Background(), etcd.EtcdClient.EtcdBasePath+"/auto_scaler", nil)
	if err != nil {
		log.Error("Load all auto scaler error: %s", err)
		log.Error(response)
		return nil, err
	}

	replicationControllerAutoScalerSlice := make([]ReplicationControllerAutoScaler, 0)
	for _, node := range response.Node.Nodes {
		replicationControllerAutoScaler := ReplicationControllerAutoScaler{}
		err := json.Unmarshal([]byte(node.Value), &replicationControllerAutoScaler)
		if err != nil {
			log.Error("Unmarshal replicationControllerAutoScaler %v error %s", node.Value, err)
			return nil, err
		}
		replicationControllerAutoScalerSlice = append(replicationControllerAutoScalerSlice, replicationControllerAutoScaler)
	}

	return replicationControllerAutoScalerSlice, nil
}
