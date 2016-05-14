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
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

type StorageEtcd struct {
}

func (storageEtcd *StorageEtcd) initialize() error {
	if err := etcd.EtcdClient.CreateDirectoryIfNotExist(etcd.EtcdClient.EtcdBasePath + "/glusterfs_cluster"); err != nil {
		log.Error("Create if not existing glusterfs cluster directory error: %s", err)
		return err
	}

	if err := etcd.EtcdClient.CreateDirectoryIfNotExist(etcd.EtcdClient.EtcdBasePath + "/glusterfs_volume_create_parameter"); err != nil {
		log.Error("Create if not existing glusterfs volume create parameter directory error: %s", err)
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
	etcdError, _ := err.(client.Error)
	if etcdError.Code == client.ErrorCodeKeyNotFound {
		return nil, etcdError
	}
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

func (storageEtcd *StorageEtcd) DeleteGlusterfsVolumeCreateParameter(clusterName string, volumeName string) error {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return err
	}

	response, err := keysAPI.Delete(context.Background(), etcd.EtcdClient.EtcdBasePath+"/glusterfs_volume_create_parameter/"+clusterName+"/"+volumeName, nil)
	if err != nil {
		log.Error("Delete glusterfs volume create parameter with clusterName %s volumeName %s error: %s", clusterName, volumeName, err)
		log.Error(response)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) SaveGlusterfsVolumeCreateParameter(glusterfsVolumeCreateParameter *GlusterfsVolumeCreateParameter) error {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return err
	}

	byteSlice, err := json.Marshal(glusterfsVolumeCreateParameter)
	if err != nil {
		log.Error("Marshal glusterfs volume create parameter %v error %s", glusterfsVolumeCreateParameter, err)
		return err
	}

	response, err := keysAPI.Set(context.Background(), etcd.EtcdClient.EtcdBasePath+"/glusterfs_volume_create_parameter/"+glusterfsVolumeCreateParameter.ClusterName+"/"+glusterfsVolumeCreateParameter.VolumeName, string(byteSlice), nil)
	if err != nil {
		log.Error("Save glusterfs volume create parameter %v error: %s", glusterfsVolumeCreateParameter, err)
		log.Error(response)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) LoadGlusterfsVolumeCreateParameter(clusterName string, volumeName string) (*GlusterfsVolumeCreateParameter, error) {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return nil, err
	}

	response, err := keysAPI.Get(context.Background(), etcd.EtcdClient.EtcdBasePath+"/glusterfs_volume_create_parameter/"+clusterName+"/"+volumeName, nil)
	etcdError, _ := err.(client.Error)
	if etcdError.Code == client.ErrorCodeKeyNotFound {
		return nil, etcdError
	}
	if err != nil {
		log.Error("Load glusterfs volume create parameter with clusterName %s volumeName %s error: %s", clusterName, volumeName, err)
		log.Error(response)
		return nil, err
	}

	glusterfsVolumeCreateParameter := new(GlusterfsVolumeCreateParameter)
	err = json.Unmarshal([]byte(response.Node.Value), &glusterfsVolumeCreateParameter)
	if err != nil {
		log.Error("Unmarshal glusterfs volume create parameter %v error %s", response.Node.Value, err)
		return nil, err
	}

	return glusterfsVolumeCreateParameter, nil
}

func (storageEtcd *StorageEtcd) LoadAllGlusterfsVolumeCreateParameter(clusterName string) ([]GlusterfsVolumeCreateParameter, error) {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return nil, err
	}

	response, err := keysAPI.Get(context.Background(), etcd.EtcdClient.EtcdBasePath+"/glusterfs_volume_create_parameter/"+clusterName, nil)
	etcdError, _ := err.(client.Error)
	if etcdError.Code == client.ErrorCodeKeyNotFound {
		return nil, etcdError
	}
	if err != nil {
		log.Error("Load all glusterfs volume create parameter with clusterName %s error: %s", clusterName, err)
		log.Error(response)
		return nil, err
	}

	glusterfsVolumeCreateParameterSlice := make([]GlusterfsVolumeCreateParameter, 0)
	for _, node := range response.Node.Nodes {
		glusterfsVolumeCreateParameter := GlusterfsVolumeCreateParameter{}
		err := json.Unmarshal([]byte(node.Value), &glusterfsVolumeCreateParameter)
		if err != nil {
			log.Error("Unmarshal glusterfs volume create parameter %v error %s", node.Value, err)
			return nil, err
		}
		glusterfsVolumeCreateParameterSlice = append(glusterfsVolumeCreateParameterSlice, glusterfsVolumeCreateParameter)
	}

	return glusterfsVolumeCreateParameterSlice, nil
}
