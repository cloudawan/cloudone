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
	"encoding/json"
	"github.com/cloudawan/cloudone/utility/database/etcd"
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

type StorageEtcd struct {
}

func (storageEtcd *StorageEtcd) initialize() error {
	if err := etcd.EtcdClient.CreateDirectoryIfNotExist(etcd.EtcdClient.EtcdBasePath + "/slb_daemon"); err != nil {
		log.Error("Create if not existing slb daemon directory error: %s", err)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) DeleteSLBDaemon(name string) error {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return err
	}

	response, err := keysAPI.Delete(context.Background(), etcd.EtcdClient.EtcdBasePath+"/slb_daemon/"+name, nil)
	etcdError, _ := err.(client.Error)
	if etcdError.Code == client.ErrorCodeKeyNotFound {
		log.Debug(err)
		log.Debug(response)
		return nil
	}
	if err != nil {
		log.Error("Delete slb daemon with name %s error: %s", name, err)
		log.Error(response)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) SaveSLBDaemon(slbDaemon *SLBDaemon) error {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return err
	}

	byteSlice, err := json.Marshal(slbDaemon)
	if err != nil {
		log.Error("Marshal slb daemon %v error %s", slbDaemon, err)
		return err
	}

	response, err := keysAPI.Set(context.Background(), etcd.EtcdClient.EtcdBasePath+"/slb_daemon/"+slbDaemon.Name, string(byteSlice), nil)
	if err != nil {
		log.Error("Save slb daemon %v error: %s", slbDaemon, err)
		log.Error(response)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) LoadSLBDaemon(name string) (*SLBDaemon, error) {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return nil, err
	}

	response, err := keysAPI.Get(context.Background(), etcd.EtcdClient.EtcdBasePath+"/slb_daemon/"+name, nil)
	etcdError, _ := err.(client.Error)
	if etcdError.Code == client.ErrorCodeKeyNotFound {
		return nil, etcdError
	}
	if err != nil {
		log.Error("Load slb daemon with name %s error: %s", name, err)
		log.Error(response)
		return nil, err
	}

	slbDaemon := new(SLBDaemon)
	err = json.Unmarshal([]byte(response.Node.Value), &slbDaemon)
	if err != nil {
		log.Error("Unmarshal slb daemon %v error %s", response.Node.Value, err)
		return nil, err
	}

	return slbDaemon, nil
}

func (storageEtcd *StorageEtcd) LoadAllSLBDaemon() ([]SLBDaemon, error) {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return nil, err
	}

	response, err := keysAPI.Get(context.Background(), etcd.EtcdClient.EtcdBasePath+"/slb_daemon", nil)
	if err != nil {
		log.Error("Load all slb daemon error: %s", err)
		log.Error(response)
		return nil, err
	}

	slbDaemonSlice := make([]SLBDaemon, 0)
	for _, node := range response.Node.Nodes {
		slbDaemon := SLBDaemon{}
		err := json.Unmarshal([]byte(node.Value), &slbDaemon)
		if err != nil {
			log.Error("Unmarshal slb daemon %v error %s", node.Value, err)
			return nil, err
		}
		slbDaemonSlice = append(slbDaemonSlice, slbDaemon)
	}

	return slbDaemonSlice, nil
}
