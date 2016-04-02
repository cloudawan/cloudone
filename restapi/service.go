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
	"fmt"
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

	ws.Route(ws.GET("/{namespace}").Filter(authorize).To(getAllKubernetesService).
		Doc("Get all of the kubernetes service in the namespace").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.QueryParameter("kubeapihost", "Kubernetes host").DataType("string")).
		Param(ws.QueryParameter("kubeapiport", "Kubernetes port").DataType("int")).
		Do(returns200AllKubernetesService, returns400, returns404, returns500))

	ws.Route(ws.POST("/{namespace}").Filter(authorize).To(postKubernetesService).
		Doc("Add a kubernetes service in the namespace").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.QueryParameter("kubeapihost", "Kubernetes host").DataType("string")).
		Param(ws.QueryParameter("kubeapiport", "Kubernetes port").DataType("int")).
		Do(returns200, returns400, returns404, returns500).
		Reads(ServiceInputDescription{}))

	ws.Route(ws.DELETE("/{namespace}/{service}").Filter(authorize).To(deleteKubernetesService).
		Doc("Delete the kubernetes service in the namespace").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.PathParameter("service", "Kubernetes service name").DataType("string")).
		Param(ws.QueryParameter("kubeapihost", "Kubernetes host").DataType("string")).
		Param(ws.QueryParameter("kubeapiport", "Kubernetes port").DataType("int")).
		Do(returns200, returns400, returns404, returns500))

	ws.Route(ws.POST("/json/{namespace}").Filter(authorize).To(postKubernetesServiceFromJson).
		Doc("Add an kubernetes service in the namespace from json source").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.QueryParameter("kubeapihost", "Kubernetes host").DataType("string")).
		Param(ws.QueryParameter("kubeapiport", "Kubernetes port").DataType("int")).
		Do(returns200, returns400, returns404, returns500).
		Reads(new(struct{})))

	ws.Route(ws.PUT("/json/{namespace}/{service}").Filter(authorize).To(putKubernetesServiceFromJson).
		Doc("Add an kubernetes service in the namespace from json source").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.PathParameter("service", "Kubernetes service name").DataType("string")).
		Param(ws.QueryParameter("kubeapihost", "Kubernetes host").DataType("string")).
		Param(ws.QueryParameter("kubeapiport", "Kubernetes port").DataType("int")).
		Do(returns200, returns400, returns404, returns500).
		Reads(new(struct{})))
}

func getAllKubernetesService(request *restful.Request, response *restful.Response) {
	kubeapiHost := request.QueryParameter("kubeapihost")
	kubeapiPortText := request.QueryParameter("kubeapiport")
	namespace := request.PathParameter("namespace")
	if kubeapiHost == "" || kubeapiPortText == "" || namespace == "" {
		errorText := fmt.Sprintf("Input text is incorrect kubeapiHost %s kubeapiPort %s namespace %s", kubeapiHost, kubeapiPortText, namespace)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}
	kubeapiPort, err := strconv.Atoi(kubeapiPortText)
	if err != nil {
		errorText := fmt.Sprintf("Could not parse kubeapiPortText %s with error %s", kubeapiPortText, err)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	serviceSlice, err := control.GetAllService(kubeapiHost, kubeapiPort, namespace)
	if err != nil {
		errorText := fmt.Sprintf("Get all service failure kubeapiHost %s kubeapiPort %s namespace %s", kubeapiHost, kubeapiPortText, namespace)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}

	response.WriteJson(serviceSlice, "[]Service")
}

func postKubernetesService(request *restful.Request, response *restful.Response) {
	kubeapiHost := request.QueryParameter("kubeapihost")
	kubeapiPortText := request.QueryParameter("kubeapiport")
	namespace := request.PathParameter("namespace")
	if kubeapiHost == "" || kubeapiPortText == "" || namespace == "" {
		errorText := fmt.Sprintf("Input text is incorrect kubeapiHost %s kubeapiPort %s namespace %s", kubeapiHost, kubeapiPortText, namespace)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}
	kubeapiPort, err := strconv.Atoi(kubeapiPortText)
	if err != nil {
		errorText := fmt.Sprintf("Could not parse kubeapiPortText %s with error %s", kubeapiPortText, err)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	service := new(control.Service)
	err = request.ReadEntity(&service)

	if err != nil {
		errorText := fmt.Sprintf("POST namespace %s kubeapiHost %s kubeapiPort %s failure with error %s", namespace, kubeapiHost, kubeapiPort, err)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	if service.Namespace != namespace {
		errorText := fmt.Sprintf("POST path parameter %s is different from body %s", namespace, service.Namespace)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	err = control.CreateService(kubeapiHost, kubeapiPort, namespace, *service)

	if err != nil {
		errorText := fmt.Sprintf("Create service failure kubeapiHost %s kubeapiPort %d namespace %s service %s error %s", kubeapiHost, kubeapiPort, namespace, service, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
}

func deleteKubernetesService(request *restful.Request, response *restful.Response) {
	kubeapiHost := request.QueryParameter("kubeapihost")
	kubeapiPortText := request.QueryParameter("kubeapiport")
	namespace := request.PathParameter("namespace")
	service := request.PathParameter("service")
	if kubeapiHost == "" || kubeapiPortText == "" || namespace == "" || service == "" {
		errorText := fmt.Sprintf("Input text is incorrect kubeapiHost %s kubeapiHost %s namespace %s service %s", kubeapiHost, kubeapiPortText, namespace, service)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}
	kubeapiPort, err := strconv.Atoi(kubeapiPortText)
	if err != nil {
		errorText := fmt.Sprintf("Could not parse kubeapiPortText %s with error %s", kubeapiPortText, err)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	err = control.DeleteService(kubeapiHost, kubeapiPort, namespace, service)

	if err != nil {
		errorText := fmt.Sprintf("Delete service failure kubeapiHost %s kubeapiPort %s namespace %s service %s error %s", kubeapiHost, kubeapiPort, namespace, service, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
}

func postKubernetesServiceFromJson(request *restful.Request, response *restful.Response) {
	kubeapiHost := request.QueryParameter("kubeapihost")
	kubeapiPortText := request.QueryParameter("kubeapiport")
	namespace := request.PathParameter("namespace")
	if kubeapiHost == "" || kubeapiPortText == "" || namespace == "" {
		errorText := fmt.Sprintf("Input text is incorrect kubeapiHost %s kubeapiPort %s namespace %s", kubeapiHost, kubeapiPortText, namespace)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}
	kubeapiPort, err := strconv.Atoi(kubeapiPortText)
	if err != nil {
		errorText := fmt.Sprintf("Could not parse kubeapiPortText %s with error %s", kubeapiPortText, err)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	service := make(map[string]interface{})
	err = request.ReadEntity(&service)

	if err != nil {
		errorText := fmt.Sprintf("POST namespace %s kubeapiHost %s kubeapiPort %s failure with error %s", namespace, kubeapiHost, kubeapiPort, err)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	err = control.CreateServiceWithJson(kubeapiHost, kubeapiPort, namespace, service)

	if err != nil {
		errorText := fmt.Sprintf("Create service failure kubeapiHost %s kubeapiPort %d namespace %s service %s error %s", kubeapiHost, kubeapiPort, namespace, service, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
}

func putKubernetesServiceFromJson(request *restful.Request, response *restful.Response) {
	kubeapiHost := request.QueryParameter("kubeapihost")
	kubeapiPortText := request.QueryParameter("kubeapiport")
	namespace := request.PathParameter("namespace")
	serviceName := request.PathParameter("service")
	if kubeapiHost == "" || kubeapiPortText == "" || namespace == "" {
		errorText := fmt.Sprintf("Input text is incorrect kubeapiHost %s kubeapiPort %s namespace %s", kubeapiHost, kubeapiPortText, namespace)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}
	kubeapiPort, err := strconv.Atoi(kubeapiPortText)
	if err != nil {
		errorText := fmt.Sprintf("Could not parse kubeapiPortText %s with error %s", kubeapiPortText, err)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	service := make(map[string]interface{})
	err = request.ReadEntity(&service)

	if err != nil {
		errorText := fmt.Sprintf("PUT namespace %s kubeapiHost %s kubeapiPort %s failure with error %s", namespace, kubeapiHost, kubeapiPort, err)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	err = control.UpdateServiceWithJson(kubeapiHost, kubeapiPort, namespace, serviceName, service)

	if err != nil {
		errorText := fmt.Sprintf("Update service failure kubeapiHost %s kubeapiPort %d namespace %s service %s error %s", kubeapiHost, kubeapiPort, namespace, service, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
}

func returns200AllKubernetesService(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", []control.Service{})
}
