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
	"github.com/cloudawan/cloudone/control"
	"github.com/cloudawan/cloudone/image"
	"strconv"
	"time"
)

var waitingDuration = 5 * time.Second

type DeployInformation struct {
	Namespace                 string
	ImageInformationName      string
	CurrentVersion            string
	CurrentVersionDescription string
	Description               string
}

func DeployCreate(
	kubeapiHost string, kubeapiPort int, namespace string, imageInformationName string,
	version string, description string, replicaAmount int,
	replicationControllerContainerPortSlice []control.ReplicationControllerContainerPort,
	replicationControllerContainerEnvironmentSlice []control.ReplicationControllerContainerEnvironment) error {

	imageRecord, err := image.GetStorage().LoadImageRecord(imageInformationName, version)
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

	// Automatically generate the basic default service. For advanced configuration, it should be modified in the service
	servicePortSlice := make([]control.ServicePort, 0)
	for _, replicationControllerContainerPort := range replicationControllerContainerPortSlice {
		containerPort := strconv.Itoa(replicationControllerContainerPort.ContainerPort)
		servicePort := control.ServicePort{
			replicationControllerContainerPort.Name,
			"TCP",
			containerPort,
			containerPort,
			"0",
		}
		servicePortSlice = append(servicePortSlice, servicePort)
	}
	selectorLabelMap := make(map[string]interface{})
	selectorLabelMap["name"] = selectorName
	serviceLabelMap := make(map[string]interface{})
	serviceLabelMap["name"] = imageInformationName
	service := control.Service{
		imageInformationName,
		namespace,
		servicePortSlice,
		selectorLabelMap,
		"",
		serviceLabelMap,
		"",
	}
	err = control.CreateService(kubeapiHost, kubeapiPort, namespace, service)
	if err != nil {
		log.Error("Create service error: %s", err)
		return err
	}

	deployInformation := &DeployInformation{
		namespace,
		imageInformationName,
		version,
		imageRecord.Description,
		description,
	}

	err = GetStorage().saveDeployInformation(deployInformation)
	if err != nil {
		log.Error("Save deploy information error: %s", err)
		return err
	}

	return nil
}

func DeployUpdate(kubeapiHost string, kubeapiPort int, namespace string,
	imageInformationName string, version string, description string,
	environmentSlice []control.ReplicationControllerContainerEnvironment) error {

	imageRecord, err := image.GetStorage().LoadImageRecord(imageInformationName, version)
	if err != nil {
		log.Error("Load image record error: %s imageInformationName %s version %s", err, imageInformationName, version)
		return err
	}

	deployInformation, err := GetStorage().LoadDeployInformation(namespace, imageInformationName)
	if err != nil {
		log.Error("Load deploy information error: %s imageInformationName %s version %s", err, imageInformationName, version)
		return err
	}

	oldVersion := deployInformation.CurrentVersion
	deployInformation.CurrentVersion = version
	deployInformation.Description = description

	deployInformation.CurrentVersionDescription = imageRecord.Description

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

	err = GetStorage().saveDeployInformation(deployInformation)
	if err != nil {
		log.Error("Save deploy information error: %s", err)
		return err
	}

	return nil
}

func DeployDelete(kubeapiHost string, kubeapiPort int, namespace string, imageInformation string) error {
	deployInformation, err := GetStorage().LoadDeployInformation(namespace, imageInformation)
	if err != nil {
		log.Error(err)
		return err
	}

	err = GetStorage().DeleteDeployInformation(namespace, imageInformation)
	if err != nil {
		log.Error(err)
		return err
	}

	replicationControllerName := deployInformation.ImageInformationName + deployInformation.CurrentVersion

	err = control.DeleteReplicationControllerAndRelatedPod(kubeapiHost, kubeapiPort, namespace, replicationControllerName)
	if err != nil {
		log.Error(err)
		return err
	}

	err = control.DeleteService(kubeapiHost, kubeapiPort, namespace, deployInformation.ImageInformationName)
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}
