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

package image

import (
	"encoding/json"
	"github.com/cloudawan/cloudone/utility/database/etcd"
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

type StorageEtcd struct {
}

func (storageEtcd *StorageEtcd) initialize() error {
	if err := etcd.EtcdClient.CreateDirectoryIfNotExist(etcd.EtcdClient.EtcdBasePath + "/image_information"); err != nil {
		log.Error("Create if not existing image information directory error: %s", err)
		return err
	}

	if err := etcd.EtcdClient.CreateDirectoryIfNotExist(etcd.EtcdClient.EtcdBasePath + "/image_record"); err != nil {
		log.Error("Create if not existing image record directory error: %s", err)
		return err
	}

	if err := etcd.EtcdClient.CreateDirectoryIfNotExist(etcd.EtcdClient.EtcdBasePath + "/image_information_build_lock"); err != nil {
		log.Error("Create if not existing image information build lock directory error: %s", err)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) DeleteImageInformationAndRelatedRecord(name string) error {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return err
	}

	response, err := keysAPI.Delete(context.Background(), etcd.EtcdClient.EtcdBasePath+"/image_information/"+name, nil)
	if err != nil {
		log.Error("Delete image information with name %s error: %s", name, err)
		log.Error(response)
		return err
	}

	return storageEtcd.deleteImageRecordWithImageInformationName(name)
}

func (storageEtcd *StorageEtcd) SaveImageInformation(imageInformation *ImageInformation) error {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return err
	}

	byteSlice, err := json.Marshal(imageInformation)
	if err != nil {
		log.Error("Marshal image information %v error %s", imageInformation, err)
		return err
	}

	response, err := keysAPI.Set(context.Background(), etcd.EtcdClient.EtcdBasePath+"/image_information/"+imageInformation.Name, string(byteSlice), nil)
	if err != nil {
		log.Error("Save image information %v error: %s", imageInformation, err)
		log.Error(response)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) LoadImageInformation(name string) (*ImageInformation, error) {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return nil, err
	}

	response, err := keysAPI.Get(context.Background(), etcd.EtcdClient.EtcdBasePath+"/image_information/"+name, nil)
	etcdError, _ := err.(client.Error)
	if etcdError.Code == client.ErrorCodeKeyNotFound {
		return nil, etcdError
	}
	if err != nil {
		log.Error("Load image information with name %s error: %s", name, err)
		log.Error(response)
		return nil, err
	}

	imageInformation := new(ImageInformation)
	err = json.Unmarshal([]byte(response.Node.Value), &imageInformation)
	if err != nil {
		log.Error("Unmarshal image information %v error %s", response.Node.Value, err)
		return nil, err
	}

	return imageInformation, nil
}

func (storageEtcd *StorageEtcd) LoadAllImageInformation() ([]ImageInformation, error) {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return nil, err
	}

	response, err := keysAPI.Get(context.Background(), etcd.EtcdClient.EtcdBasePath+"/image_information", nil)
	if err != nil {
		log.Error("Load all image information error: %s", err)
		log.Error(response)
		return nil, err
	}

	imageInformationSlice := make([]ImageInformation, 0)
	for _, node := range response.Node.Nodes {
		imageInformation := ImageInformation{}
		err := json.Unmarshal([]byte(node.Value), &imageInformation)
		if err != nil {
			log.Error("Unmarshal image information %v error %s", node.Value, err)
			return nil, err
		}
		imageInformationSlice = append(imageInformationSlice, imageInformation)
	}

	return imageInformationSlice, nil
}

func (storageEtcd *StorageEtcd) DeleteImageRecord(imageInformationName string, version string) error {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return err
	}

	response, err := keysAPI.Delete(context.Background(), etcd.EtcdClient.EtcdBasePath+"/image_record/"+imageInformationName+"/"+version, nil)
	if err != nil {
		log.Error("Delete image record with image information %s version %s error: %s", imageInformationName, version, err)
		log.Error(response)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) deleteImageRecordWithImageInformationName(imageInformationName string) error {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return err
	}

	response, err := keysAPI.Delete(context.Background(), etcd.EtcdClient.EtcdBasePath+"/image_record/"+imageInformationName, &client.DeleteOptions{Recursive: true})
	if err != nil {
		log.Error("Delete image record with image information %s error: %s", imageInformationName, err)
		log.Error(response)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) saveImageRecord(imageRecord *ImageRecord) error {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return err
	}

	byteSlice, err := json.Marshal(imageRecord)
	if err != nil {
		log.Error("Marshal image record %v error %s", imageRecord, err)
		return err
	}

	response, err := keysAPI.Set(context.Background(), etcd.EtcdClient.EtcdBasePath+"/image_record/"+imageRecord.ImageInformation+"/"+imageRecord.Version, string(byteSlice), nil)
	if err != nil {
		log.Error("Save image record %v error: %s", imageRecord, err)
		log.Error(response)
		return err
	}

	return nil
}

func (storageEtcd *StorageEtcd) LoadImageRecord(imageInformationName string, version string) (*ImageRecord, error) {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return nil, err
	}

	response, err := keysAPI.Get(context.Background(), etcd.EtcdClient.EtcdBasePath+"/image_record/"+imageInformationName+"/"+version, nil)
	etcdError, _ := err.(client.Error)
	if etcdError.Code == client.ErrorCodeKeyNotFound {
		return nil, etcdError
	}
	if err != nil {
		log.Error("Load image record with image information %s version %s error: %s", imageInformationName, version, err)
		log.Error(response)
		return nil, err
	}

	imageRecord := new(ImageRecord)
	err = json.Unmarshal([]byte(response.Node.Value), &imageRecord)
	if err != nil {
		log.Error("Unmarshal image record %v error %s", response.Node.Value, err)
		return nil, err
	}

	return imageRecord, nil
}

func (storageEtcd *StorageEtcd) LoadImageRecordWithImageInformationName(imageInformationName string) ([]ImageRecord, error) {
	keysAPI, err := etcd.EtcdClient.GetKeysAPI()
	if err != nil {
		log.Error("Get keysAPI error %s", err)
		return nil, err
	}

	response, err := keysAPI.Get(context.Background(), etcd.EtcdClient.EtcdBasePath+"/image_record/"+imageInformationName, nil)
	etcdError, _ := err.(client.Error)
	if etcdError.Code == client.ErrorCodeKeyNotFound {
		return nil, etcdError
	}
	if err != nil {
		log.Error("Load all image record belonging to image information %s error: %s", imageInformationName, err)
		log.Error(response)
		return nil, err
	}

	imageRecordSlice := make([]ImageRecord, 0)
	for _, node := range response.Node.Nodes {
		imageRecord := ImageRecord{}
		err := json.Unmarshal([]byte(node.Value), &imageRecord)
		if err != nil {
			log.Error("Unmarshal image record %v error %s", node.Value, err)
			return nil, err
		}
		imageRecordSlice = append(imageRecordSlice, imageRecord)
	}

	return imageRecordSlice, nil
}

// If Load all image record is used, the second level directory may have empty issue and need to be solved
