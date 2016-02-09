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

package glusterfs

import (
	"encoding/json"
	"github.com/cloudawan/cloudone/utility/database/etcd"
	"golang.org/x/net/context"
)

type StorageEtcd struct {
}

func (storageEtcd *StorageEtcd) initialize() error {
	if err := etcd.EtcdClient.CreateDirectoryIfNotExist(etcd.EtcdClient.EtcdBasePath + "/glusterfs_cluster"); err != nil {
		log.Error("Create if not existing auto scaler directory error: %s", err)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) DeleteGlusterfsCluster(name string) error {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return err
	}

	response, err := keysAPI.Delete(context.Background(), etcd.EtcdClient.EtcdBasePath+"/glusterfs_cluster/"+name, nil)
	if err != nil {
		log.Error("Delete glusterfs cluster with name %s error: %s", name, err)
		log.Error(response)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) SaveGlusterfsCluster(glusterfsCluster *GlusterfsCluster) error {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return err
	}

	byteSlice, err := json.Marshal(glusterfsCluster)
	if err != nil {
		log.Error("Marshal glusterfs cluster %v error %s", glusterfsCluster, err)
		return err
	}

	response, err := keysAPI.Set(context.Background(), etcd.EtcdClient.EtcdBasePath+"/glusterfs_cluster/"+glusterfsCluster.Name, string(byteSlice), nil)
	if err != nil {
		log.Error("Save glusterfs cluster %v error: %s", glusterfsCluster, err)
		log.Error(response)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) LoadGlusterfsCluster(name string) (*GlusterfsCluster, error) {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return nil, err
	}

	response, err := keysAPI.Get(context.Background(), etcd.EtcdClient.EtcdBasePath+"/glusterfs_cluster/"+name, nil)
	if err != nil {
		log.Error("Load glusterfs cluster with name %s error: %s", name, err)
		log.Error(response)
		return nil, err
	}

	glusterfsCluster := new(GlusterfsCluster)
	err = json.Unmarshal([]byte(response.Node.Value), &glusterfsCluster)
	if err != nil {
		log.Error("Unmarshal glusterfs cluster %v error %s", response.Node.Value, err)
		return nil, err
	}

	return glusterfsCluster, nil
}

func (storageEtcd *StorageEtcd) LoadAllGlusterfsCluster() ([]GlusterfsCluster, error) {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return nil, err
	}

	response, err := keysAPI.Get(context.Background(), etcd.EtcdClient.EtcdBasePath+"/glusterfs_cluster", nil)
	if err != nil {
		log.Error("Load all glusterfs cluster error: %s", err)
		log.Error(response)
		return nil, err
	}

	glusterfsClusterSlice := make([]GlusterfsCluster, 0)
	for _, node := range response.Node.Nodes {
		glusterfsCluster := GlusterfsCluster{}
		err := json.Unmarshal([]byte(node.Value), &glusterfsCluster)
		if err != nil {
			log.Error("Unmarshal glusterfs cluster %v error %s", node.Value, err)
			return nil, err
		}
		glusterfsClusterSlice = append(glusterfsClusterSlice, glusterfsCluster)
	}

	return glusterfsClusterSlice, nil
}
