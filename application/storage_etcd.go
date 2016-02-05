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
	"github.com/cloudawan/cloudone/utility/database/etcd"
	"golang.org/x/net/context"
)

type StorageEtcd struct {
}

func (storageEtcd *StorageEtcd) initialize() error {
	if err := etcd.EtcdClient.CreateDirectoryIfNotExist(etcd.EtcdClient.EtcdBasePath + "/stateless_application"); err != nil {
		log.Error("Create if not existing stateless application directory error: %s", err)
		return err
	}

	if err := etcd.EtcdClient.CreateDirectoryIfNotExist(etcd.EtcdClient.EtcdBasePath + "/cluster_application"); err != nil {
		log.Error("Create if not existing cluster application directory error: %s", err)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) DeleteStatelessApplication(name string) error {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return err
	}

	response, err := keysAPI.Delete(context.Background(), etcd.EtcdClient.EtcdBasePath+"/stateless_application/"+name, nil)
	if err != nil {
		log.Error("Delete stateless application with name %s error: %s", name, err)
		log.Error(response)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) saveStatelessApplication(stateless *Stateless) error {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return err
	}

	byteSlice, err := json.Marshal(stateless)
	if err != nil {
		log.Error("Marshal stateless application %v error %s", stateless, err)
		return err
	}

	response, err := keysAPI.Set(context.Background(), etcd.EtcdClient.EtcdBasePath+"/stateless_application/"+stateless.Name, string(byteSlice), nil)
	if err != nil {
		log.Error("Save stateless application %v error: %s", stateless, err)
		log.Error(response)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) LoadStatelessApplication(name string) (*Stateless, error) {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return nil, err
	}

	response, err := keysAPI.Get(context.Background(), etcd.EtcdClient.EtcdBasePath+"/stateless_application/"+name, nil)
	if err != nil {
		log.Error("Load stateless application with name %s error: %s", name, err)
		log.Error(response)
		return nil, err
	}

	stateless := new(Stateless)
	err = json.Unmarshal([]byte(response.Node.Value), &stateless)
	if err != nil {
		log.Error("Unmarshal stateless application %v error %s", response.Node.Value, err)
		return nil, err
	}

	return stateless, nil
}

func (storageEtcd *StorageEtcd) LoadAllStatelessApplication() ([]Stateless, error) {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return nil, err
	}

	response, err := keysAPI.Get(context.Background(), etcd.EtcdClient.EtcdBasePath+"/stateless_application", nil)
	if err != nil {
		log.Error("Load all stateless application error: %s", err)
		log.Error(response)
		return nil, err
	}

	statelessSlice := make([]Stateless, 0)
	for _, node := range response.Node.Nodes {
		stateless := Stateless{}
		err := json.Unmarshal([]byte(node.Value), &stateless)
		if err != nil {
			log.Error("Unmarshal stateless application %v error %s", node.Value, err)
			return nil, err
		}
		statelessSlice = append(statelessSlice, stateless)
	}

	return statelessSlice, nil
}

func (storageEtcd *StorageEtcd) DeleteClusterApplication(name string) error {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return err
	}

	response, err := keysAPI.Delete(context.Background(), etcd.EtcdClient.EtcdBasePath+"/cluster_application/"+name, nil)
	if err != nil {
		log.Error("Delete cluster application with name %s error: %s", name, err)
		log.Error(response)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) SaveClusterApplication(cluster *Cluster) error {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return err
	}

	byteSlice, err := json.Marshal(cluster)
	if err != nil {
		log.Error("Marshal cluster %v error %s", cluster, err)
		return err
	}

	response, err := keysAPI.Set(context.Background(), etcd.EtcdClient.EtcdBasePath+"/cluster_application/"+cluster.Name, string(byteSlice), nil)
	if err != nil {
		log.Error("Save cluster application %v error: %s", cluster, err)
		log.Error(response)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) LoadClusterApplication(name string) (*Cluster, error) {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return nil, err
	}

	response, err := keysAPI.Get(context.Background(), etcd.EtcdClient.EtcdBasePath+"/cluster_application/"+name, nil)
	if err != nil {
		log.Error("Load cluster application with name %s error: %s", name, err)
		log.Error(response)
		return nil, err
	}

	cluster := new(Cluster)
	err = json.Unmarshal([]byte(response.Node.Value), &cluster)
	if err != nil {
		log.Error("Unmarshal cluster application %v error %s", response.Node.Value, err)
		return nil, err
	}

	return cluster, nil
}

func (storageEtcd *StorageEtcd) LoadAllClusterApplication() ([]Cluster, error) {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return nil, err
	}

	response, err := keysAPI.Get(context.Background(), etcd.EtcdClient.EtcdBasePath+"/cluster_application", nil)
	if err != nil {
		log.Error("Load all cluster application error: %s", err)
		log.Error(response)
		return nil, err
	}

	clusterSlice := make([]Cluster, 0)
	for _, node := range response.Node.Nodes {
		cluster := Cluster{}
		err := json.Unmarshal([]byte(node.Value), &cluster)
		if err != nil {
			log.Error("Unmarshal cluster application %v error %s", node.Value, err)
			return nil, err
		}
		clusterSlice = append(clusterSlice, cluster)
	}

	return clusterSlice, nil
}
