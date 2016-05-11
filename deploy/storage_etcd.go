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

package deploy

import (
	"encoding/json"
	"github.com/cloudawan/cloudone/utility/database/etcd"
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

type StorageEtcd struct {
}

func (storageEtcd *StorageEtcd) initialize() error {
	if err := etcd.EtcdClient.CreateDirectoryIfNotExist(etcd.EtcdClient.EtcdBasePath + "/deploy_information"); err != nil {
		log.Error("Create if not existing deploy information directory error: %s", err)
		return err
	}

	if err := etcd.EtcdClient.CreateDirectoryIfNotExist(etcd.EtcdClient.EtcdBasePath + "/deploy_blue_green"); err != nil {
		log.Error("Create if not existing deploy blue green directory error: %s", err)
		return err
	}

	if err := etcd.EtcdClient.CreateDirectoryIfNotExist(etcd.EtcdClient.EtcdBasePath + "/deploy_cluster_application"); err != nil {
		log.Error("Create if not existing deploy cluster application directory error: %s", err)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) getKeyDeployInformation(namespace string, imageInformation string) string {
	return namespace + "." + imageInformation
}

func (storageEtcd *StorageEtcd) DeleteDeployInformation(namespace string, imageInformation string) error {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return err
	}

	key := storageEtcd.getKeyDeployInformation(namespace, imageInformation)
	response, err := keysAPI.Delete(context.Background(), etcd.EtcdClient.EtcdBasePath+"/deploy_information/"+key, nil)
	if err != nil {
		log.Error("Delete deploy information with namespace %s imageInformation %s error: %s", namespace, imageInformation, err)
		log.Error(response)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) saveDeployInformation(deployInformation *DeployInformation) error {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return err
	}

	byteSlice, err := json.Marshal(deployInformation)
	if err != nil {
		log.Error("Marshal deploy information %v error %s", deployInformation, err)
		return err
	}

	key := storageEtcd.getKeyDeployInformation(deployInformation.Namespace, deployInformation.ImageInformationName)
	response, err := keysAPI.Set(context.Background(), etcd.EtcdClient.EtcdBasePath+"/deploy_information/"+key, string(byteSlice), nil)
	if err != nil {
		log.Error("Save deploy information %v error: %s", deployInformation, err)
		log.Error(response)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) LoadDeployInformation(namespace string, imageInformation string) (*DeployInformation, error) {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return nil, err
	}

	key := storageEtcd.getKeyDeployInformation(namespace, imageInformation)
	response, err := keysAPI.Get(context.Background(), etcd.EtcdClient.EtcdBasePath+"/deploy_information/"+key, nil)
	etcdError, _ := err.(client.Error)
	if etcdError.Code == client.ErrorCodeKeyNotFound {
		return nil, etcdError
	}
	if err != nil {
		log.Error("Load deploy information with namespace %s imageInformation %s error: %s", namespace, imageInformation, err)
		log.Error(response)
		return nil, err
	}

	deployInformation := new(DeployInformation)
	err = json.Unmarshal([]byte(response.Node.Value), &deployInformation)
	if err != nil {
		log.Error("Unmarshal deploy information %v error %s", response.Node.Value, err)
		return nil, err
	}

	return deployInformation, nil
}

func (storageEtcd *StorageEtcd) LoadAllDeployInformation() ([]DeployInformation, error) {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return nil, err
	}

	response, err := keysAPI.Get(context.Background(), etcd.EtcdClient.EtcdBasePath+"/deploy_information", nil)
	if err != nil {
		log.Error("Load all deploy information error: %s", err)
		log.Error(response)
		return nil, err
	}

	deployInformationSlice := make([]DeployInformation, 0)
	for _, node := range response.Node.Nodes {
		deployInformation := DeployInformation{}
		err := json.Unmarshal([]byte(node.Value), &deployInformation)
		if err != nil {
			log.Error("Unmarshal deploy information %v error %s", node.Value, err)
			return nil, err
		}
		deployInformationSlice = append(deployInformationSlice, deployInformation)
	}

	return deployInformationSlice, nil
}

func (storageEtcd *StorageEtcd) DeleteDeployBlueGreen(imageInformation string) error {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return err
	}

	response, err := keysAPI.Delete(context.Background(), etcd.EtcdClient.EtcdBasePath+"/deploy_blue_green/"+imageInformation, nil)
	if err != nil {
		log.Error("Delete deploy blue green with image information %s error: %s", imageInformation, err)
		log.Error(response)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) saveDeployBlueGreen(deployBlueGreen *DeployBlueGreen) error {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return err
	}

	byteSlice, err := json.Marshal(deployBlueGreen)
	if err != nil {
		log.Error("Marshal deploy blue green %v error %s", deployBlueGreen, err)
		return err
	}

	response, err := keysAPI.Set(context.Background(), etcd.EtcdClient.EtcdBasePath+"/deploy_blue_green/"+deployBlueGreen.ImageInformation, string(byteSlice), nil)
	if err != nil {
		log.Error("Save deploy blue green %v error: %s", deployBlueGreen, err)
		log.Error(response)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) LoadDeployBlueGreen(imageInformation string) (*DeployBlueGreen, error) {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return nil, err
	}

	response, err := keysAPI.Get(context.Background(), etcd.EtcdClient.EtcdBasePath+"/deploy_blue_green/"+imageInformation, nil)
	etcdError, _ := err.(client.Error)
	if etcdError.Code == client.ErrorCodeKeyNotFound {
		return nil, etcdError
	}
	if err != nil {
		log.Error("Load deploy blue green with imageInformation %s error: %s", imageInformation, err)
		log.Error(response)
		return nil, err
	}

	deployBlueGreen := new(DeployBlueGreen)
	err = json.Unmarshal([]byte(response.Node.Value), &deployBlueGreen)
	if err != nil {
		log.Error("Unmarshal deploy blue green %v error %s", response.Node.Value, err)
		return nil, err
	}

	return deployBlueGreen, nil
}

func (storageEtcd *StorageEtcd) LoadAllDeployBlueGreen() ([]DeployBlueGreen, error) {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return nil, err
	}

	response, err := keysAPI.Get(context.Background(), etcd.EtcdClient.EtcdBasePath+"/deploy_blue_green", nil)
	if err != nil {
		log.Error("Load all deploy blue green error: %s", err)
		log.Error(response)
		return nil, err
	}

	deployBlueGreenSlice := make([]DeployBlueGreen, 0)
	for _, node := range response.Node.Nodes {
		deployBlueGreen := DeployBlueGreen{}
		err := json.Unmarshal([]byte(node.Value), &deployBlueGreen)
		if err != nil {
			log.Error("Unmarshal deploy blue green %v error %s", node.Value, err)
			return nil, err
		}
		deployBlueGreenSlice = append(deployBlueGreenSlice, deployBlueGreen)
	}

	return deployBlueGreenSlice, nil
}

func (storageEtcd *StorageEtcd) getKeyDeployClusterApplication(namespace string, name string) string {
	return namespace + "." + name
}

func (storageEtcd *StorageEtcd) DeleteDeployClusterApplication(namespace string, name string) error {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return err
	}

	key := storageEtcd.getKeyDeployClusterApplication(namespace, name)
	response, err := keysAPI.Delete(context.Background(), etcd.EtcdClient.EtcdBasePath+"/deploy_cluster_application/"+key, nil)
	if err != nil {
		log.Error("Delete deploy cluster application with namespace %s name %s error: %s", namespace, name, err)
		log.Error(response)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) SaveDeployClusterApplication(deployClusterApplication *DeployClusterApplication) error {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return err
	}

	byteSlice, err := json.Marshal(deployClusterApplication)
	if err != nil {
		log.Error("Marshal deploy cluster application %v error %s", deployClusterApplication, err)
		return err
	}

	key := storageEtcd.getKeyDeployClusterApplication(deployClusterApplication.Namespace, deployClusterApplication.Name)
	response, err := keysAPI.Set(context.Background(), etcd.EtcdClient.EtcdBasePath+"/deploy_cluster_application/"+key, string(byteSlice), nil)
	if err != nil {
		log.Error("Save deploy cluster application %v error: %s", deployClusterApplication, err)
		log.Error(response)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) LoadDeployClusterApplication(namespace string, name string) (*DeployClusterApplication, error) {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return nil, err
	}

	key := storageEtcd.getKeyDeployClusterApplication(namespace, name)
	response, err := keysAPI.Get(context.Background(), etcd.EtcdClient.EtcdBasePath+"/deploy_cluster_application/"+key, nil)
	etcdError, _ := err.(client.Error)
	if etcdError.Code == client.ErrorCodeKeyNotFound {
		return nil, etcdError
	}
	if err != nil {
		log.Error("Load deploy cluster application with namespace %s name %s error: %s", namespace, name, err)
		log.Error(response)
		return nil, err
	}

	deployClusterApplication := new(DeployClusterApplication)
	err = json.Unmarshal([]byte(response.Node.Value), &deployClusterApplication)
	if err != nil {
		log.Error("Unmarshal deploy cluster application %v error %s", response.Node.Value, err)
		return nil, err
	}

	return deployClusterApplication, nil
}

func (storageEtcd *StorageEtcd) LoadAllDeployClusterApplication() ([]DeployClusterApplication, error) {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return nil, err
	}

	response, err := keysAPI.Get(context.Background(), etcd.EtcdClient.EtcdBasePath+"/deploy_cluster_application", nil)
	if err != nil {
		log.Error("Load all deploy cluster application error: %s", err)
		log.Error(response)
		return nil, err
	}

	deployClusterApplicationSlice := make([]DeployClusterApplication, 0)
	for _, node := range response.Node.Nodes {
		deployClusterApplication := DeployClusterApplication{}
		err := json.Unmarshal([]byte(node.Value), &deployClusterApplication)
		if err != nil {
			log.Error("Unmarshal deploy cluster application %v error %s", node.Value, err)
			return nil, err
		}
		deployClusterApplicationSlice = append(deployClusterApplicationSlice, deployClusterApplication)
	}

	return deployClusterApplicationSlice, nil
}
