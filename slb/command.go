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

package slb

import (
	"github.com/cloudawan/cloudone/control"
	"github.com/cloudawan/cloudone/deploy"
	"github.com/cloudawan/cloudone/utility/configuration"
	"github.com/cloudawan/cloudone_utility/slb"
	"strconv"
	"time"
)

const (
	BlueGreenDeploymentPrefix = "bg."
)

func CreateCommand() (*slb.Command, error) {
	kubeApiServerEndPoint, kubeApiServerToken, err := configuration.GetAvailablekubeApiServerEndPoint()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	command := &slb.Command{
		time.Now(),
		nil,
		nil,
	}

	err = addCommandFromAllDeployInformation(command, kubeApiServerEndPoint, kubeApiServerToken)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	err = addCommandFromAllBlueGreenDeployment(command, kubeApiServerEndPoint, kubeApiServerToken)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return command, nil
}

func addCommandFromAllDeployInformation(command *slb.Command, kubeApiServerEndPoint string, kubeApiServerToken string) error {
	deployInformationSlice, err := deploy.GetStorage().LoadAllDeployInformation()
	if err != nil {
		log.Error(err)
		return err
	}

	kubernetesServiceHTTPSlice := make([]slb.KubernetesServiceHTTP, 0)

	for _, deployInformation := range deployInformationSlice {
		service, err := control.GetService(kubeApiServerEndPoint, kubeApiServerToken, deployInformation.Namespace, deployInformation.ImageInformationName)
		if err != nil {
			log.Error(err)
			return err
		}

		for _, servicePort := range service.PortSlice {
			// Get protocol
			protocol := ""
			for _, containerPort := range deployInformation.ContainerPortSlice {
				if servicePort.TargetPort == strconv.Itoa(containerPort.ContainerPort) {
					protocol = containerPort.Protocol
				}
			}
			// HTTP
			if protocol == deploy.ProtocolTypeHTTP {
				kubernetesServiceHTTP := slb.KubernetesServiceHTTP{
					deployInformation.Namespace,
					deployInformation.ImageInformationName,
					servicePort.Port,
					servicePort.NodePort,
				}

				kubernetesServiceHTTPSlice = append(kubernetesServiceHTTPSlice, kubernetesServiceHTTP)
			}
		}
	}

	if command.KubernetesServiceHTTPSlice == nil {
		command.KubernetesServiceHTTPSlice = kubernetesServiceHTTPSlice
	} else {
		command.KubernetesServiceHTTPSlice = append(command.KubernetesServiceHTTPSlice, kubernetesServiceHTTPSlice...)
	}

	return nil
}

func addCommandFromAllBlueGreenDeployment(command *slb.Command, kubeApiServerEndPoint string, kubeApiServerToken string) error {
	deployBlueGreenSlice, err := deploy.GetStorage().LoadAllDeployBlueGreen()
	if err != nil {
		log.Error(err)
		return err
	}

	kubernetesServiceHTTPSlice := make([]slb.KubernetesServiceHTTP, 0)

	for _, deployBlueGreen := range deployBlueGreenSlice {
		deployInformation, err := deploy.GetStorage().LoadDeployInformation(deployBlueGreen.Namespace, deployBlueGreen.ImageInformation)
		if err != nil {
			log.Error(err)
			return err
		}

		serviceName := deploy.GetBlueGreenServiceName(deployBlueGreen.ImageInformation)
		service, err := control.GetService(kubeApiServerEndPoint, kubeApiServerToken, deployBlueGreen.Namespace, serviceName)
		if err != nil {
			log.Error(err)
			return err
		}

		for _, servicePort := range service.PortSlice {
			// Get protocol
			protocol := ""
			for _, containerPort := range deployInformation.ContainerPortSlice {
				if servicePort.TargetPort == strconv.Itoa(containerPort.ContainerPort) {
					protocol = containerPort.Protocol
				}
			}
			// HTTP
			if protocol == deploy.ProtocolTypeHTTP {
				kubernetesServiceHTTP := slb.KubernetesServiceHTTP{
					deployBlueGreen.Namespace,
					BlueGreenDeploymentPrefix + deployBlueGreen.ImageInformation,
					servicePort.Port,
					servicePort.NodePort,
				}

				kubernetesServiceHTTPSlice = append(kubernetesServiceHTTPSlice, kubernetesServiceHTTP)
			}
		}
	}

	if command.KubernetesServiceHTTPSlice == nil {
		command.KubernetesServiceHTTPSlice = kubernetesServiceHTTPSlice
	} else {
		command.KubernetesServiceHTTPSlice = append(command.KubernetesServiceHTTPSlice, kubernetesServiceHTTPSlice...)
	}

	return nil
}
