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
	"encoding/json"
	"github.com/cloudawan/cloudone/utility/database/etcd"
	"golang.org/x/net/context"
)

type StorageEtcd struct {
}

func (storageEtcd *StorageEtcd) initialize() error {
	if err := etcd.EtcdClient.CreateDirectoryIfNotExist(etcd.EtcdClient.EtcdBasePath + "/notifier"); err != nil {
		log.Error("Create if not existing notifier directory error: %s", err)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) getKeyAutoScaler(namespace string, kind string, name string) string {
	return namespace + "." + kind + "." + name
}

func (storageEtcd *StorageEtcd) DeleteReplicationControllerNotifierSerializable(namespace string, kind string, name string) error {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return err
	}

	key := storageEtcd.getKeyAutoScaler(namespace, kind, name)
	response, err := keysAPI.Delete(context.Background(), etcd.EtcdClient.EtcdBasePath+"/notifier/"+key, nil)
	if err != nil {
		log.Error("Delete notifier with namespace %s kind %s name %s error: %s", namespace, kind, name, err)
		log.Error(response)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) SaveReplicationControllerNotifierSerializable(replicationControllerNotifierSerializable *ReplicationControllerNotifierSerializable) error {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return err
	}

	byteSlice, err := json.Marshal(replicationControllerNotifierSerializable)
	if err != nil {
		log.Error("Marshal notifier %v error %s", replicationControllerNotifierSerializable, err)
		return err
	}

	key := storageEtcd.getKeyAutoScaler(replicationControllerNotifierSerializable.Namespace, replicationControllerNotifierSerializable.Kind, replicationControllerNotifierSerializable.Name)
	response, err := keysAPI.Set(context.Background(), etcd.EtcdClient.EtcdBasePath+"/notifier/"+key, string(byteSlice), nil)
	if err != nil {
		log.Error("Save notifier %v error: %s", replicationControllerNotifierSerializable, err)
		log.Error(response)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) LoadReplicationControllerNotifierSerializable(namespace string, kind string, name string) (*ReplicationControllerNotifierSerializable, error) {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return nil, err
	}

	key := storageEtcd.getKeyAutoScaler(namespace, kind, name)
	response, err := keysAPI.Get(context.Background(), etcd.EtcdClient.EtcdBasePath+"/notifier/"+key, nil)
	if err != nil {
		log.Error("Load notifier with name %s error: %s", name, err)
		log.Error(response)
		return nil, err
	}

	replicationControllerNotifierSerializable := new(ReplicationControllerNotifierSerializable)
	err = json.Unmarshal([]byte(response.Node.Value), &replicationControllerNotifierSerializable)
	if err != nil {
		log.Error("Unmarshal notifier %v error %s", response.Node.Value, err)
		return nil, err
	}

	return replicationControllerNotifierSerializable, nil
}

func (storageEtcd *StorageEtcd) LoadAllReplicationControllerNotifierSerializable() ([]ReplicationControllerNotifierSerializable, error) {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return nil, err
	}

	response, err := keysAPI.Get(context.Background(), etcd.EtcdClient.EtcdBasePath+"/notifier", nil)
	if err != nil {
		log.Error("Load all notifier error: %s", err)
		log.Error(response)
		return nil, err
	}

	replicationControllerNotifierSerializableSlice := make([]ReplicationControllerNotifierSerializable, 0)
	for _, node := range response.Node.Nodes {
		replicationControllerNotifierSerializable := ReplicationControllerNotifierSerializable{}
		err := json.Unmarshal([]byte(node.Value), &replicationControllerNotifierSerializable)
		if err != nil {
			log.Error("Unmarshal notifier %v error %s", node.Value, err)
			return nil, err
		}
		replicationControllerNotifierSerializableSlice = append(replicationControllerNotifierSerializableSlice, replicationControllerNotifierSerializable)
	}

	return replicationControllerNotifierSerializableSlice, nil
}
