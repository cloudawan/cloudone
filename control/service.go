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

package control

import (
	"encoding/json"
	"github.com/cloudawan/cloudone_utility/logger"
	"github.com/cloudawan/cloudone_utility/restclient"
	"strconv"
)

type Service struct {
	Name            string
	Namespace       string
	PortSlice       []ServicePort
	Selector        map[string]interface{}
	ClusterIP       string
	LabelMap        map[string]interface{}
	SessionAffinity string
}

type ServicePort struct {
	Name       string
	Protocol   string
	Port       int
	TargetPort string
	NodePort   int
}

func CreateService(kubeapiHost string, kubeapiPort int, namespace string, service Service) (returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("CreateService Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedError = err.(error)
		}
	}()

	hasNodePort := false

	portJsonMapSlice := make([]map[string]interface{}, 0)
	for _, port := range service.PortSlice {
		portJsonMap := make(map[string]interface{})
		portJsonMap["name"] = port.Name
		portJsonMap["protocol"] = port.Protocol
		portJsonMap["port"] = port.Port
		targetPortNumber, err := strconv.Atoi(port.TargetPort)
		if err != nil {
			portJsonMap["targetPort"] = port.TargetPort
		} else {
			portJsonMap["targetPort"] = targetPortNumber
		}

		// < 0 not used, == 0 auto-generated, > 0 port number
		if port.NodePort >= 0 {
			hasNodePort = true
			if port.NodePort > 0 {
				portJsonMap["nodePort"] = port.NodePort
			}
		}

		portJsonMapSlice = append(portJsonMapSlice, portJsonMap)
	}

	bodyJsonMap := make(map[string]interface{})
	bodyJsonMap["kind"] = "Service"
	bodyJsonMap["apiVersion"] = "v1"
	bodyJsonMap["metadata"] = make(map[string]interface{})
	bodyJsonMap["metadata"].(map[string]interface{})["name"] = service.Name
	bodyJsonMap["metadata"].(map[string]interface{})["labels"] = service.LabelMap
	bodyJsonMap["spec"] = make(map[string]interface{})
	bodyJsonMap["spec"].(map[string]interface{})["ports"] = portJsonMapSlice
	bodyJsonMap["spec"].(map[string]interface{})["selector"] = service.Selector

	// Use sticky session so the same client will be forwarded to the same pod
	if service.SessionAffinity != "" {
		bodyJsonMap["spec"].(map[string]interface{})["sessionAffinity"] = service.SessionAffinity
	}

	if hasNodePort {
		bodyJsonMap["spec"].(map[string]interface{})["type"] = "NodePort"
	}

	url := "http://" + kubeapiHost + ":" + strconv.Itoa(kubeapiPort) + "/api/v1/namespaces/" + namespace + "/services/"
	_, err := restclient.RequestPost(url, bodyJsonMap, nil, true)

	if err != nil {
		log.Error(err)
	}

	return err
}

func DeleteService(kubeapiHost string, kubeapiPort int, namespace string, serviceName string) (returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("DeleteService Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedError = err.(error)
		}
	}()

	url := "http://" + kubeapiHost + ":" + strconv.Itoa(kubeapiPort) + "/api/v1/namespaces/" + namespace + "/services/" + serviceName
	_, err := restclient.RequestDelete(url, nil, nil, true)

	return err
}

func GetService(kubeapiHost string, kubeapiPort int, namespace string, serviceName string) (returnedService *Service, returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("GetService Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedService = nil
			returnedError = err.(error)
		}
	}()

	url := "http://" + kubeapiHost + ":" + strconv.Itoa(kubeapiPort) +
		"/api/v1/namespaces/" + namespace + "/services/" + serviceName
	result, err := restclient.RequestGet(url, nil, true)
	jsonMap, _ := result.(map[string]interface{})
	if err != nil {
		return nil, err
	} else {
		service := new(Service)

		service.Name, _ = jsonMap["metadata"].(map[string]interface{})["name"].(string)
		service.Namespace, _ = jsonMap["metadata"].(map[string]interface{})["namespace"].(string)
		service.LabelMap, _ = jsonMap["metadata"].(map[string]interface{})["labels"].(map[string]interface{})
		service.ClusterIP, _ = jsonMap["spec"].(map[string]interface{})["clusterIP"].(string)
		service.Selector, _ = jsonMap["spec"].(map[string]interface{})["selector"].(map[string]interface{})
		service.SessionAffinity, _ = jsonMap["spec"].(map[string]interface{})["sessionAffinity"].(string)
		portSlice := jsonMap["spec"].(map[string]interface{})["ports"].([]interface{})
		servicePortSlice := make([]ServicePort, 0)
		for _, port := range portSlice {
			servicePort := ServicePort{}
			servicePort.Name, _ = port.(map[string]interface{})["name"].(string)
			servicePort.Protocol, _ = port.(map[string]interface{})["protocol"].(string)
			unknownTypePort := port.(map[string]interface{})["port"]
			switch knowTypePort := unknownTypePort.(type) {
			case json.Number:
				portInt64, _ := knowTypePort.Int64()
				servicePort.Port = int(portInt64)
			case string:
				servicePort.Port, _ = strconv.Atoi(knowTypePort)
			case int64:
				servicePort.Port = int(knowTypePort)
			case float64:
				servicePort.Port = int(knowTypePort)
			default:
				servicePort.Port = -1
			}
			unknownTypeTargetPort := port.(map[string]interface{})["targetPort"]
			switch knowTypeTargetPort := unknownTypeTargetPort.(type) {
			case json.Number:
				servicePort.TargetPort = knowTypeTargetPort.String()
			case string:
				servicePort.TargetPort = knowTypeTargetPort
			case int64:
				servicePort.TargetPort = strconv.FormatInt(knowTypeTargetPort, 10)
			case float64:
				servicePort.TargetPort = strconv.FormatFloat(knowTypeTargetPort, 'f', -1, 64)
			}
			unknownTypeNodePort := port.(map[string]interface{})["nodePort"]
			switch knowTypeNodePort := unknownTypeNodePort.(type) {
			case json.Number:
				nodePortInt64, _ := knowTypeNodePort.Int64()
				servicePort.NodePort = int(nodePortInt64)
			case string:
				servicePort.NodePort, _ = strconv.Atoi(knowTypeNodePort)
			case int64:
				servicePort.NodePort = int(knowTypeNodePort)
			case float64:
				servicePort.NodePort = int(knowTypeNodePort)
			default:
				servicePort.NodePort = -1
			}
			servicePortSlice = append(servicePortSlice, servicePort)
		}
		service.PortSlice = servicePortSlice

		return service, nil
	}
}

func GetAllService(kubeapiHost string, kubeapiPort int, namespace string) (returnedServiceSlice []Service, returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("GetAllService Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedServiceSlice = nil
			returnedError = err.(error)
		}
	}()

	url := "http://" + kubeapiHost + ":" + strconv.Itoa(kubeapiPort) + "/api/v1/namespaces/" + namespace + "/services/"
	result, err := restclient.RequestGet(url, nil, true)
	jsonMap, _ := result.(map[string]interface{})
	if err != nil {
		return nil, err
	} else {
		serviceSlice := make([]Service, 0)
		for _, item := range jsonMap["items"].([]interface{}) {
			service := Service{}
			service.Name, _ = item.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
			service.Namespace, _ = item.(map[string]interface{})["metadata"].(map[string]interface{})["namespace"].(string)
			service.LabelMap, _ = item.(map[string]interface{})["metadata"].(map[string]interface{})["labels"].(map[string]interface{})
			service.ClusterIP, _ = item.(map[string]interface{})["spec"].(map[string]interface{})["clusterIP"].(string)
			service.Selector, _ = item.(map[string]interface{})["spec"].(map[string]interface{})["selector"].(map[string]interface{})
			service.SessionAffinity, _ = item.(map[string]interface{})["spec"].(map[string]interface{})["sessionAffinity"].(string)
			portSlice := item.(map[string]interface{})["spec"].(map[string]interface{})["ports"].([]interface{})
			servicePortSlice := make([]ServicePort, 0)
			for _, port := range portSlice {
				servicePort := ServicePort{}
				servicePort.Name, _ = port.(map[string]interface{})["name"].(string)
				servicePort.Protocol, _ = port.(map[string]interface{})["protocol"].(string)
				unknownTypePort := port.(map[string]interface{})["port"]
				switch knowTypePort := unknownTypePort.(type) {
				case json.Number:
					portInt64, _ := knowTypePort.Int64()
					servicePort.Port = int(portInt64)
				case string:
					servicePort.Port, _ = strconv.Atoi(knowTypePort)
				case int64:
					servicePort.Port = int(knowTypePort)
				case float64:
					servicePort.Port = int(knowTypePort)
				}
				unknownTypeTargetPort := port.(map[string]interface{})["targetPort"]
				switch knowTypeTargetPort := unknownTypeTargetPort.(type) {
				case json.Number:
					servicePort.TargetPort = knowTypeTargetPort.String()
				case string:
					servicePort.TargetPort = knowTypeTargetPort
				case int64:
					servicePort.TargetPort = strconv.FormatInt(knowTypeTargetPort, 10)
				case float64:
					servicePort.TargetPort = strconv.FormatFloat(knowTypeTargetPort, 'f', -1, 64)
				}
				unknownTypeNodePort := port.(map[string]interface{})["nodePort"]
				switch knowTypeNodePort := unknownTypeNodePort.(type) {
				case json.Number:
					nodePortInt64, _ := knowTypeNodePort.Int64()
					servicePort.NodePort = int(nodePortInt64)
				case string:
					servicePort.NodePort, _ = strconv.Atoi(knowTypeNodePort)
				case int64:
					servicePort.NodePort = int(knowTypeNodePort)
				case float64:
					servicePort.NodePort = int(knowTypeNodePort)
				default:
					servicePort.NodePort = -1
				}
				servicePortSlice = append(servicePortSlice, servicePort)
			}
			service.PortSlice = servicePortSlice
			serviceSlice = append(serviceSlice, service)
		}
		return serviceSlice, nil
	}
}

func CreateServiceWithJson(kubeapiHost string, kubeapiPort int, namespace string, bodyJsonMap map[string]interface{}) (returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("CreateServiceWithJson Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedError = err.(error)
		}
	}()

	url := "http://" + kubeapiHost + ":" + strconv.Itoa(kubeapiPort) + "/api/v1/namespaces/" + namespace + "/services/"
	_, err := restclient.RequestPost(url, bodyJsonMap, nil, true)

	if err != nil {
		log.Error(err)
		return err
	} else {
		return nil
	}
}

func UpdateServiceWithJson(kubeapiHost string, kubeapiPort int, namespace string, serviceName string, bodyJsonMap map[string]interface{}) (returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("CreateServiceWithJson Error: %s", err)
			log.Error(logger.GetStackTrace(4096, false))
			returnedError = err.(error)
		}
	}()

	url := "http://" + kubeapiHost + ":" + strconv.Itoa(kubeapiPort) + "/api/v1/namespaces/" + namespace + "/services/" + serviceName
	result, err := restclient.RequestGet(url, nil, true)
	jsonMap, _ := result.(map[string]interface{})

	metadataJsonMap, _ := jsonMap["metadata"].(map[string]interface{})
	resourceVersion, _ := metadataJsonMap["resourceVersion"].(string)

	specJsonMap, _ := jsonMap["spec"].(map[string]interface{})
	clusterIP, _ := specJsonMap["clusterIP"].(string)

	// Update requires the resoruce version and cluster ip
	bodyJsonMap["metadata"].(map[string]interface{})["resourceVersion"] = resourceVersion
	bodyJsonMap["spec"].(map[string]interface{})["clusterIP"] = clusterIP

	url = "http://" + kubeapiHost + ":" + strconv.Itoa(kubeapiPort) + "/api/v1/namespaces/" + namespace + "/services/" + serviceName
	_, err = restclient.RequestPut(url, bodyJsonMap, nil, true)

	if err != nil {
		log.Error(err)
		return err
	} else {
		return nil
	}
}
