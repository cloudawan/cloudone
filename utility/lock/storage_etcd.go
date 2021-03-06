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

package lock

import (
	"encoding/json"
	"github.com/cloudawan/cloudone/utility/database/etcd"
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

type StorageEtcd struct {
}

func (storageEtcd *StorageEtcd) initialize() error {
	if err := etcd.EtcdClient.CreateDirectoryIfNotExist(etcd.EtcdClient.EtcdBasePath + "/lock"); err != nil {
		log.Error("Create if not existing lock directory error: %s", err)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) deleteLock(name string) error {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return err
	}

	response, err := keysAPI.Delete(context.Background(), etcd.EtcdClient.EtcdBasePath+"/lock/"+name, nil)
	etcdError, _ := err.(client.Error)
	if etcdError.Code == client.ErrorCodeKeyNotFound {
		log.Debug(err)
		log.Debug(response)
		return nil
	}
	if err != nil {
		log.Error("Delete lock with name %s error: %s", name, err)
		log.Error(response)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) saveLock(lock *Lock) error {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return err
	}

	byteSlice, err := json.Marshal(lock)
	if err != nil {
		log.Error("Marshal lock %v error %s", lock, err)
		return err
	}

	response, err := keysAPI.Set(context.Background(), etcd.EtcdClient.EtcdBasePath+"/lock/"+lock.Name, string(byteSlice), &client.SetOptions{TTL: lock.Timeout})
	if err != nil {
		log.Error("Save lock %v error: %s", lock, err)
		log.Error(response)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) loadLock(name string) (*Lock, error) {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return nil, err
	}

	response, err := keysAPI.Get(context.Background(), etcd.EtcdClient.EtcdBasePath+"/lock/"+name, &client.GetOptions{Quorum: true})
	etcdError, _ := err.(client.Error)
	if etcdError.Code == client.ErrorCodeKeyNotFound {
		return nil, etcdError
	}
	if err != nil {
		log.Error("Load lock with name %s error: %s", name, err)
		log.Error(response)
		return nil, err
	}

	lock := new(Lock)
	err = json.Unmarshal([]byte(response.Node.Value), &lock)
	if err != nil {
		log.Error("Unmarshal lock %v error %s", response.Node.Value, err)
		return nil, err
	}

	return lock, nil
}

func (storageEtcd *StorageEtcd) LoadAllLock() ([]Lock, error) {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return nil, err
	}

	response, err := keysAPI.Get(context.Background(), etcd.EtcdClient.EtcdBasePath+"/lock", &client.GetOptions{Quorum: true})
	if err != nil {
		log.Error("Load all lock error: %s", err)
		log.Error(response)
		return nil, err
	}

	lockSlice := make([]Lock, 0)
	for _, node := range response.Node.Nodes {
		lock := Lock{}
		err := json.Unmarshal([]byte(node.Value), &lock)
		if err != nil {
			log.Error("Unmarshal lock %v error %s", node.Value, err)
			return nil, err
		}
		lockSlice = append(lockSlice, lock)
	}

	return lockSlice, nil
}
