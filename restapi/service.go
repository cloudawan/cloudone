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

package restapi

import (
	"encoding/json"
	"github.com/cloudawan/cloudone/control"
	"github.com/emicklei/go-restful"
	"net/http"
	"strconv"
)

type ServiceInputDescription struct {
	Name            string
	Namespace       string
	PortSlice       []control.ServicePort
	Selector        interface{}
	ClusterIP       string
	LabelMap        interface{}
	SessionAffinity string
}

func registerWebServiceReplicationController() {
	ws := new(restful.WebService)
	ws.Path("/api/v1/services")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	restful.Add(ws)

	ws.Route(ws.GET("/{namespace}").Filter(authorize).Filter(auditLog).To(getAllKubernetesService).
		Doc("Get all of the kubernetes service in the namespace").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.QueryParameter("kubeapihost", "Kubernetes host").DataType("string")).
		Param(ws.QueryParameter("kubeapiport", "Kubernetes port").DataType("int")).
		Do(returns200AllKubernetesService, returns400, returns422, returns500))

	ws.Route(ws.POST("/{namespace}").Filter(authorize).Filter(auditLog).To(postKubernetesService).
		Doc("Add a kubernetes service in the namespace").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.QueryParameter("kubeapihost", "Kubernetes host").DataType("string")).
		Param(ws.QueryParameter("kubeapiport", "Kubernetes port").DataType("int")).
		Do(returns200, returns400, returns422, returns500).
		Reads(ServiceInputDescription{}))

	ws.Route(ws.DELETE("/{namespace}/{service}").Filter(authorize).Filter(auditLog).To(deleteKubernetesService).
		Doc("Delete the kubernetes service in the namespace").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.PathParameter("service", "Kubernetes service name").DataType("string")).
		Param(ws.QueryParameter("kubeapihost", "Kubernetes host").DataType("string")).
		Param(ws.QueryParameter("kubeapiport", "Kubernetes port").DataType("int")).
		Do(returns200, returns400, returns422, returns500))

	ws.Route(ws.POST("/json/{namespace}").Filter(authorize).Filter(auditLog).To(postKubernetesServiceFromJson).
		Doc("Add an kubernetes service in the namespace from json source").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.QueryParameter("kubeapihost", "Kubernetes host").DataType("string")).
		Param(ws.QueryParameter("kubeapiport", "Kubernetes port").DataType("int")).
		Do(returns200, returns400, returns422, returns500).
		Reads(new(struct{})))

	ws.Route(ws.PUT("/json/{namespace}/{service}").Filter(authorize).Filter(auditLog).To(putKubernetesServiceFromJson).
		Doc("Add an kubernetes service in the namespace from json source").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.PathParameter("service", "Kubernetes service name").DataType("string")).
		Param(ws.QueryParameter("kubeapihost", "Kubernetes host").DataType("string")).
		Param(ws.QueryParameter("kubeapiport", "Kubernetes port").DataType("int")).
		Do(returns200, returns400, returns422, returns500).
		Reads(new(struct{})))
}

func getAllKubernetesService(request *restful.Request, response *restful.Response) {
	kubeapiHost := request.QueryParameter("kubeapihost")
	kubeapiPortText := request.QueryParameter("kubeapiport")
	namespace := request.PathParameter("namespace")
	if kubeapiHost == "" || kubeapiPortText == "" {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Input is incorrect. The fields kubeapihost and kubeapiport are required."
		jsonMap["kubeapiHost"] = kubeapiHost
		jsonMap["kubeapiPortText"] = kubeapiPortText
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}
	kubeapiPort, err := strconv.Atoi(kubeapiPortText)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Could not parse kubeapiPortText"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeapiPortText"] = kubeapiPortText
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	serviceSlice, err := control.GetAllService(kubeapiHost, kubeapiPort, namespace)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get all service failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeapiHost"] = kubeapiHost
		jsonMap["kubeapiPort"] = kubeapiPort
		jsonMap["namespace"] = namespace
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(serviceSlice, "[]Service")
}

func postKubernetesService(request *restful.Request, response *restful.Response) {
	kubeapiHost := request.QueryParameter("kubeapihost")
	kubeapiPortText := request.QueryParameter("kubeapiport")
	namespace := request.PathParameter("namespace")
	if kubeapiHost == "" || kubeapiPortText == "" {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Input is incorrect. The fields kubeapihost and kubeapiport are required."
		jsonMap["kubeapiHost"] = kubeapiHost
		jsonMap["kubeapiPortText"] = kubeapiPortText
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}
	kubeapiPort, err := strconv.Atoi(kubeapiPortText)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Could not parse kubeapiPortText"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeapiPortText"] = kubeapiPortText
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	service := new(control.Service)
	err = request.ReadEntity(&service)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Read body failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeapiHost"] = kubeapiHost
		jsonMap["kubeapiPort"] = kubeapiPort
		jsonMap["namespace"] = namespace
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	if service.Namespace != namespace {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Path parameter namespace is different from namespace in the body"
		jsonMap["path"] = namespace
		jsonMap["body"] = service.Namespace
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	err = control.CreateService(kubeapiHost, kubeapiPort, namespace, *service)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Create service failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeapiHost"] = kubeapiHost
		jsonMap["kubeapiPort"] = kubeapiPort
		jsonMap["namespace"] = namespace
		jsonMap["service"] = service
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func deleteKubernetesService(request *restful.Request, response *restful.Response) {
	kubeapiHost := request.QueryParameter("kubeapihost")
	kubeapiPortText := request.QueryParameter("kubeapiport")
	namespace := request.PathParameter("namespace")
	service := request.PathParameter("service")
	if kubeapiHost == "" || kubeapiPortText == "" {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Input is incorrect. The fields kubeapihost and kubeapiport are required."
		jsonMap["kubeapiHost"] = kubeapiHost
		jsonMap["kubeapiPortText"] = kubeapiPortText
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}
	kubeapiPort, err := strconv.Atoi(kubeapiPortText)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Could not parse kubeapiPortText"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeapiPortText"] = kubeapiPortText
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	err = control.DeleteService(kubeapiHost, kubeapiPort, namespace, service)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Delete service failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeapiHost"] = kubeapiHost
		jsonMap["kubeapiPort"] = kubeapiPort
		jsonMap["namespace"] = namespace
		jsonMap["service"] = service
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func postKubernetesServiceFromJson(request *restful.Request, response *restful.Response) {
	kubeapiHost := request.QueryParameter("kubeapihost")
	kubeapiPortText := request.QueryParameter("kubeapiport")
	namespace := request.PathParameter("namespace")
	if kubeapiHost == "" || kubeapiPortText == "" {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Input is incorrect. The fields kubeapihost and kubeapiport are required."
		jsonMap["kubeapiHost"] = kubeapiHost
		jsonMap["kubeapiPortText"] = kubeapiPortText
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}
	kubeapiPort, err := strconv.Atoi(kubeapiPortText)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Could not parse kubeapiPortText"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeapiPortText"] = kubeapiPortText
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	service := make(map[string]interface{})
	err = request.ReadEntity(&service)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Read body failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeapiHost"] = kubeapiHost
		jsonMap["kubeapiPort"] = kubeapiPort
		jsonMap["namespace"] = namespace
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	err = control.CreateServiceWithJson(kubeapiHost, kubeapiPort, namespace, service)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Create service failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeapiHost"] = kubeapiHost
		jsonMap["kubeapiPort"] = kubeapiPort
		jsonMap["namespace"] = namespace
		jsonMap["service"] = service
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func putKubernetesServiceFromJson(request *restful.Request, response *restful.Response) {
	kubeapiHost := request.QueryParameter("kubeapihost")
	kubeapiPortText := request.QueryParameter("kubeapiport")
	namespace := request.PathParameter("namespace")
	serviceName := request.PathParameter("service")
	if kubeapiHost == "" || kubeapiPortText == "" {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Input is incorrect. The fields kubeapihost and kubeapiport are required."
		jsonMap["kubeapiHost"] = kubeapiHost
		jsonMap["kubeapiPortText"] = kubeapiPortText
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}
	kubeapiPort, err := strconv.Atoi(kubeapiPortText)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Could not parse kubeapiPortText"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeapiPortText"] = kubeapiPortText
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	service := make(map[string]interface{})
	err = request.ReadEntity(&service)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Read body failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeapiHost"] = kubeapiHost
		jsonMap["kubeapiPort"] = kubeapiPort
		jsonMap["namespace"] = namespace
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	err = control.UpdateServiceWithJson(kubeapiHost, kubeapiPort, namespace, serviceName, service)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Update service failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeapiHost"] = kubeapiHost
		jsonMap["kubeapiPort"] = kubeapiPort
		jsonMap["namespace"] = namespace
		jsonMap["service"] = service
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func returns200AllKubernetesService(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", []control.Service{})
}
