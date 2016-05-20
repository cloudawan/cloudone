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
	"errors"
	"github.com/cloudawan/cloudone/control"
	"github.com/cloudawan/cloudone/image"
	"github.com/cloudawan/cloudone/utility/lock"
	"strconv"
	"time"
)

const (
	LockKind        = "deploy"
	waitingDuration = 5 * time.Second
)

type DeployContainerPort struct {
	Name          string
	ContainerPort int
	NodePort      int
}

type DeployInformation struct {
	Namespace                 string
	ImageInformationName      string
	CurrentVersion            string
	CurrentVersionDescription string
	Description               string
	ReplicaAmount             int
	ContainerPortSlice        []DeployContainerPort
	EnvironmentSlice          []control.ReplicationControllerContainerEnvironment
	ResourceMap               map[string]interface{}
	ExtraJsonMap              map[string]interface{}
	CreatedTime               time.Time
	AutoUpdateForNewBuild     bool
}

func GetDeployInformationInNamespace(namespace string) ([]DeployInformation, error) {
	deployInformationSlice, err := GetStorage().LoadAllDeployInformation()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	filteredDeployInformationSlice := make([]DeployInformation, 0)
	for _, deployInformation := range deployInformationSlice {
		if deployInformation.Namespace == namespace {
			filteredDeployInformationSlice = append(filteredDeployInformationSlice, deployInformation)
		}
	}
	return filteredDeployInformationSlice, nil
}

func getLockName(namespace string, imageInformationName string) string {
	return namespace + "." + imageInformationName
}

func DeployCreate(
	kubeapiHost string, kubeapiPort int, namespace string, imageInformationName string,
	version string, description string, replicaAmount int,
	deployContainerPortSlice []DeployContainerPort,
	replicationControllerContainerEnvironmentSlice []control.ReplicationControllerContainerEnvironment,
	resourceMap map[string]interface{},
	extraJsonMap map[string]interface{},
	autoUpdateForNewBuild bool) error {
	if lock.AcquireLock(LockKind, getLockName(namespace, imageInformationName), 0) == false {
		return errors.New("Application is under deployment")
	}

	defer lock.ReleaseLock(LockKind, getLockName(namespace, imageInformationName))

	imageRecord, err := image.GetStorage().LoadImageRecord(imageInformationName, version)
	if err != nil {
		log.Error("Load image record error: %s imageInformationName %s version %s", err, imageInformationName, version)
		return err
	}

	selectorName := imageInformationName
	replicationControllerName := selectorName + version
	image := imageRecord.Path

	// Automatically generate the basic default service. For advanced configuration, it should be modified in the service
	servicePortSlice := make([]control.ServicePort, 0)
	for _, deployContainerPort := range deployContainerPortSlice {
		containerPort := strconv.Itoa(deployContainerPort.ContainerPort)
		servicePort := control.ServicePort{
			deployContainerPort.Name,
			"TCP",
			deployContainerPort.ContainerPort,
			containerPort,
			deployContainerPort.NodePort, // -1 means not to use. 0 means auto-generated. > 0 means the port number to use
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

	// Replication controller
	replicationControllerContainerPortSlice := make([]control.ReplicationControllerContainerPort, 0)
	for _, deployContainerPort := range deployContainerPortSlice {
		replicationControllerContainerPortSlice = append(replicationControllerContainerPortSlice,
			control.ReplicationControllerContainerPort{deployContainerPort.Name, deployContainerPort.ContainerPort})
	}

	replicationControllerContainerSlice := make([]control.ReplicationControllerContainer, 0)
	replicationControllerContainerSlice = append(
		replicationControllerContainerSlice,
		control.ReplicationControllerContainer{
			replicationControllerName,
			image,
			replicationControllerContainerPortSlice,
			replicationControllerContainerEnvironmentSlice,
			resourceMap,
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
		extraJsonMap,
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
		imageRecord.Description,
		description,
		replicaAmount,
		deployContainerPortSlice,
		replicationControllerContainerEnvironmentSlice,
		resourceMap,
		extraJsonMap,
		time.Now(),
		autoUpdateForNewBuild,
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
	if lock.AcquireLock(LockKind, getLockName(namespace, imageInformationName), 0) == false {
		return errors.New("Application is under deployment")
	}

	defer lock.ReleaseLock(LockKind, getLockName(namespace, imageInformationName))

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

func DeployResize(kubeapiHost string, kubeapiPort int, namespace string, imageInformation string, size int) error {
	deployInformation, err := GetStorage().LoadDeployInformation(namespace, imageInformation)
	if err != nil {
		log.Error(err)
		return err
	}

	replicationControllerName := deployInformation.ImageInformationName + deployInformation.CurrentVersion

	delta := size - deployInformation.ReplicaAmount

	_, _, err = control.ResizeReplicationController(kubeapiHost, kubeapiPort, namespace, replicationControllerName, delta, deployInformation.ReplicaAmount+delta, 1)
	if err != nil {
		log.Error(err)
		return err
	}

	deployInformation.ReplicaAmount = size
	err = GetStorage().saveDeployInformation(deployInformation)
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func IsImageRecordUsed(imageInformationName string, imageRecordVersion string) (bool, error) {
	deployInformationSlice, err := GetStorage().LoadAllDeployInformation()
	if err != nil {
		log.Error(err)
		return false, err
	}

	for _, deployInformation := range deployInformationSlice {
		if deployInformation.ImageInformationName == imageInformationName && deployInformation.CurrentVersion == imageRecordVersion {
			return true, nil
		}
	}
	return false, nil
}

func IsImageInformationUsed(imageInformationName string) (bool, error) {
	deployInformationSlice, err := GetStorage().LoadAllDeployInformation()
	if err != nil {
		log.Error(err)
		return false, err
	}

	for _, deployInformation := range deployInformationSlice {
		if deployInformation.ImageInformationName == imageInformationName {
			return true, nil
		}
	}
	return false, nil
}

func GetDeployInformationWithAutoUpdateForNewBuild(imageInformationName string) ([]DeployInformation, error) {
	deployInformationSlice, err := GetStorage().LoadAllDeployInformation()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	fileteredDeployInformationSlice := make([]DeployInformation, 0)
	for _, deployInformation := range deployInformationSlice {
		if deployInformation.AutoUpdateForNewBuild && deployInformation.ImageInformationName == imageInformationName {
			fileteredDeployInformationSlice = append(fileteredDeployInformationSlice, deployInformation)
		}
	}
	return fileteredDeployInformationSlice, nil
}
