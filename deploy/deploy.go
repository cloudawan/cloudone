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
	"github.com/cloudawan/kubernetes_management/control"
	"github.com/cloudawan/kubernetes_management/image"
	"time"
)

var waitingDuration = 5 * time.Second

type DeployInformation struct {
	Namespace            string
	ImageInformationName string
	CurrentVersion       string
	Description          string
}

func DeployCreate(
	kubeapiHost string, kubeapiPort int, namespace string, imageInformationName string,
	version string, description string, replicaAmount int,
	replicationControllerContainerPortSlice []control.ReplicationControllerContainerPort,
	replicationControllerContainerEnvironmentSlice []control.ReplicationControllerContainerEnvironment) error {

	imageRecord, err := image.LoadImageRecord(imageInformationName, version)
	if err != nil {
		log.Error("Load image record error: %s imageInformationName %s version %s", err, imageInformationName, version)
		return err
	}

	selectorName := imageInformationName
	replicationControllerName := selectorName + version
	image := imageRecord.Path

	replicationControllerContainerSlice := make([]control.ReplicationControllerContainer, 0)
	replicationControllerContainerSlice = append(
		replicationControllerContainerSlice,
		control.ReplicationControllerContainer{
			replicationControllerName,
			image,
			replicationControllerContainerPortSlice,
			replicationControllerContainerEnvironmentSlice,
		})

	replicationController := control.ReplicationController{
		replicationControllerName,
		replicaAmount,
		control.ReplicationControllerSelector{
			selectorName,
			version,
		},
		control.ReplicationControllerLabel{
			replicationControllerName,
		},
		replicationControllerContainerSlice,
	}

	err = control.CreateReplicationController(kubeapiHost, kubeapiPort,
		namespace, replicationController)
	if err != nil {
		log.Error("Create replication controller error: %s", err)
		return err
	}

	deployInformation := &DeployInformation{
		namespace,
		imageInformationName,
		version,
		description,
	}

	err = saveDeployInformation(deployInformation)
	if err != nil {
		log.Error("Save deploy information error: %s", err)
		return err
	}

	return nil
}

func DeployUpdate(kubeapiHost string, kubeapiPort int, namespace string,
	imageInformationName string, version string, description string,
	environmentSlice []control.ReplicationControllerContainerEnvironment) error {

	imageRecord, err := image.LoadImageRecord(imageInformationName, version)
	if err != nil {
		log.Error("Load image record error: %s imageInformationName %s version %s", err, imageInformationName, version)
		return err
	}

	deployInformation, err := LoadDeployInformation(namespace, imageInformationName)
	if err != nil {
		log.Error("Load deploy information error: %s imageInformationName %s version %s", err, imageInformationName, version)
		return err
	}

	oldVersion := deployInformation.CurrentVersion
	deployInformation.CurrentVersion = version
	deployInformation.Description = description

	oldReplicationControllerName := deployInformation.ImageInformationName + oldVersion
	newReplicationControllerName := deployInformation.ImageInformationName + version

	err = control.RollingUpdateReplicationControllerWithSingleContainer(
		kubeapiHost, kubeapiPort, namespace,
		oldReplicationControllerName, newReplicationControllerName,
		imageRecord.Path, imageRecord.Version, waitingDuration, environmentSlice)
	if err != nil {
		log.Error("Rollingupdate replication controller error: %s", err)
		return err
	}

	err = saveDeployInformation(deployInformation)
	if err != nil {
		log.Error("Save deploy information error: %s", err)
		return err
	}

	return nil
}
