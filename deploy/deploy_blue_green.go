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
	"strconv"
)

type DeployBlueGreen struct {
	ImageInformation string
	Namespace        string
	NodePort         int
	Description      string
	SessionAffinity  string
}

const (
	blueGreenServiceNamePrefix = "bg-"
)

func getBlueGreenReplicationControllerName(imageInformation string, currentVersion string) string {
	return imageInformation + currentVersion
}

func getBlueGreenServiceName(imageInformation string) string {
	return blueGreenServiceNamePrefix + imageInformation
}

func UpdateDeployBlueGreen(kubeapiHost string, kubeapiPort int, deployBlueGreen *DeployBlueGreen) error {
	deployInformation, err := GetStorage().LoadDeployInformation(
		deployBlueGreen.Namespace, deployBlueGreen.ImageInformation)
	if err != nil {
		log.Error("Fail to load deploy information %s in namespace %s with error %s",
			deployBlueGreen.ImageInformation, deployBlueGreen.Namespace, err)
		return err
	}

	replicationController, err := control.GetReplicationController(
		kubeapiHost, kubeapiPort, deployBlueGreen.Namespace,
		getBlueGreenReplicationControllerName(deployInformation.ImageInformationName, deployInformation.CurrentVersion))
	if err != nil {
		log.Error("Fail to load target replication controller information %s in namespace %s with error %s",
			deployInformation.ImageInformationName+deployInformation.CurrentVersion, deployBlueGreen.Namespace, err)
		return err
	}

	// TODO support multiple ports
	portName := replicationController.ContainerSlice[0].PortSlice[0].Name
	containerPort := replicationController.ContainerSlice[0].PortSlice[0].ContainerPort

	// Clean all the previous blue green deployment
	CleanAllServiceUnderBlueGreenDeployment(kubeapiHost, kubeapiPort, deployBlueGreen.ImageInformation)

	selector := make(map[string]interface{})
	selector["name"] = deployBlueGreen.ImageInformation
	labelMap := make(map[string]interface{})
	labelMap["name"] = deployBlueGreen.ImageInformation

	portSlice := make([]control.ServicePort, 0)
	// TODO read protocol rather than TCP
	portSlice = append(portSlice, control.ServicePort{
		portName, "TCP", strconv.Itoa(containerPort), strconv.Itoa(containerPort), strconv.Itoa(deployBlueGreen.NodePort)})

	service := control.Service{
		getBlueGreenServiceName(deployBlueGreen.ImageInformation),
		deployBlueGreen.Namespace,
		portSlice,
		selector,
		"",
		labelMap,
		deployBlueGreen.SessionAffinity,
	}
	err = control.CreateService(kubeapiHost, kubeapiPort, deployBlueGreen.Namespace, service)
	if err != nil {
		log.Error("Create target service failure service %s with error %s",
			service, err)
		return err
	}

	// Update DeployBlueGreen
	err = GetStorage().saveDeployBlueGreen(deployBlueGreen)
	if err != nil {
		log.Error("Save deploy blude grenn %s with error %s",
			deployBlueGreen, err)
		return err
	}

	return nil
}

func CleanAllServiceUnderBlueGreenDeployment(kubeapiHost string, kubeapiPort int, imageInformationName string) error {
	// Clean all service with this deployment name
	namespaceSlice, err := control.GetAllNamespaceName(kubeapiHost, kubeapiPort)
	if err != nil {
		log.Error("Fail to get all namesapce with error %s", err)
		return err
	}
	for _, namespace := range namespaceSlice {
		service, _ := control.GetService(kubeapiHost, kubeapiPort, namespace, getBlueGreenServiceName(imageInformationName))
		if service != nil {
			err := control.DeleteService(kubeapiHost, kubeapiPort, namespace, service.Name)
			if err != nil {
				log.Error("Fail to delete service %s in namesapce %s with error %s", imageInformationName, namespace, err)
				return err
			}
		}
	}
	return nil
}

func GetAllBlueGreenDeployableNamespace(imageInformationName string) ([]string, error) {
	deployInformationSlice, err := GetStorage().LoadAllDeployInformation()
	if err != nil {
		log.Error("Fail to get all deploy information with error %s", err)
		return nil, err
	}
	namespaceSlice := make([]string, 0)
	for _, deployInformation := range deployInformationSlice {
		if deployInformation.ImageInformationName == imageInformationName {
			namespaceSlice = append(namespaceSlice, deployInformation.Namespace)
		}
	}
	return namespaceSlice, nil
}
