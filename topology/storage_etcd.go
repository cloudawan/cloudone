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

package topology

import (
	"encoding/json"
	"github.com/cloudawan/cloudone/utility/database/etcd"
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

type StorageEtcd struct {
}

func (storageEtcd *StorageEtcd) initialize() error {
	if err := etcd.EtcdClient.CreateDirectoryIfNotExist(etcd.EtcdClient.EtcdBasePath + "/topology"); err != nil {
		log.Error("Create if not existing topology directory error: %s", err)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) DeleteTopology(name string) error {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return err
	}

	response, err := keysAPI.Delete(context.Background(), etcd.EtcdClient.EtcdBasePath+"/topology/"+name, nil)
	if err != nil {
		log.Error("Delete topology with name %s error: %s", name, err)
		log.Error(response)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) SaveTopology(topology *Topology) error {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return err
	}

	byteSlice, err := json.Marshal(topology)
	if err != nil {
		log.Error("Marshal topology %v error %s", topology, err)
		return err
	}

	response, err := keysAPI.Set(context.Background(), etcd.EtcdClient.EtcdBasePath+"/topology/"+topology.Name, string(byteSlice), nil)
	if err != nil {
		log.Error("Save topology %v error: %s", topology, err)
		log.Error(response)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) LoadTopology(name string) (*Topology, error) {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return nil, err
	}

	response, err := keysAPI.Get(context.Background(), etcd.EtcdClient.EtcdBasePath+"/topology/"+name, nil)
	etcdError, _ := err.(client.Error)
	if etcdError.Code == client.ErrorCodeKeyNotFound {
		return nil, etcdError
	}
	if err != nil {
		log.Error("Load topology with name %s error: %s", name, err)
		log.Error(response)
		return nil, err
	}

	topology := new(Topology)
	err = json.Unmarshal([]byte(response.Node.Value), &topology)
	if err != nil {
		log.Error("Unmarshal topology %v error %s", response.Node.Value, err)
		return nil, err
	}

	return topology, nil
}

func (storageEtcd *StorageEtcd) LoadAllTopology() ([]Topology, error) {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return nil, err
	}

	response, err := keysAPI.Get(context.Background(), etcd.EtcdClient.EtcdBasePath+"/topology", nil)
	if err != nil {
		log.Error("Load all topology error: %s", err)
		log.Error(response)
		return nil, err
	}

	topologySlice := make([]Topology, 0)
	for _, node := range response.Node.Nodes {
		topology := Topology{}
		err := json.Unmarshal([]byte(node.Value), &topology)
		if err != nil {
			log.Error("Unmarshal topology %v error %s", node.Value, err)
			return nil, err
		}
		topologySlice = append(topologySlice, topology)
	}

	return topologySlice, nil
}
