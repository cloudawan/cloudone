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

type Namesapce struct {
	Name string
}

func registerWebServiceNamespace() {
	ws := new(restful.WebService)
	ws.Path("/api/v1/namespaces")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	restful.Add(ws)

	ws.Route(ws.GET("/").Filter(authorize).Filter(auditLog).To(getAllKubernetesNamespaceName).
		Doc("Get all of the kubernetes namespace name").
		Param(ws.QueryParameter("kubeapihost", "Kubernetes host").DataType("string")).
		Param(ws.QueryParameter("kubeapiport", "Kubernetes port").DataType("int")).
		Do(returns200AllKubernetesNamesapceName, returns400, returns404, returns500))

	ws.Route(ws.POST("/").Filter(authorize).Filter(auditLog).To(postKubernetesNamespace).
		Doc("Add a kubernetes namespace").
		Param(ws.QueryParameter("kubeapihost", "Kubernetes host").DataType("string")).
		Param(ws.QueryParameter("kubeapiport", "Kubernetes port").DataType("int")).
		Do(returns200, returns400, returns404, returns500).
		Reads(Namesapce{}))

	ws.Route(ws.DELETE("/{namespace}").Filter(authorize).Filter(auditLog).To(deleteKubernetesNamespace).
		Doc("Delete the kubernetes namespace").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.QueryParameter("kubeapihost", "Kubernetes host").DataType("string")).
		Param(ws.QueryParameter("kubeapiport", "Kubernetes port").DataType("int")).
		Do(returns200, returns400, returns404, returns500))
}

func getAllKubernetesNamespaceName(request *restful.Request, response *restful.Response) {
	kubeapiHost := request.QueryParameter("kubeapihost")
	kubeapiPortText := request.QueryParameter("kubeapiport")
	if kubeapiHost == "" || kubeapiPortText == "" {
		errorText := fmt.Sprintf("Input text is incorrect kubeapiHost %s kubeapiPort %s", kubeapiHost, kubeapiPortText)
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

	nameSlice, err := control.GetAllNamespaceName(kubeapiHost, kubeapiPort)
	if err != nil {
		errorText := fmt.Sprintf("Get all namespace name failure kubeapiHost %s kubeapiPort %s", kubeapiHost, kubeapiPortText)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}

	response.WriteJson(nameSlice, "[]string")
}

func postKubernetesNamespace(request *restful.Request, response *restful.Response) {
	kubeapiHost := request.QueryParameter("kubeapihost")
	kubeapiPortText := request.QueryParameter("kubeapiport")
	if kubeapiHost == "" || kubeapiPortText == "" {
		errorText := fmt.Sprintf("Input text is incorrect kubeapiHost %s kubeapiPort %s", kubeapiHost, kubeapiPortText)
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

	namespace := new(Namesapce)
	err = request.ReadEntity(&namespace)

	if err != nil {
		errorText := fmt.Sprintf("POST namespace %s kubeapiHost %s kubeapiPort %s failure with error %s", namespace, kubeapiHost, kubeapiPort, err)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	err = control.CreateNamespace(kubeapiHost, kubeapiPort, namespace.Name)

	if err != nil {
		errorText := fmt.Sprintf("Create namespace failure kubeapiHost %s kubeapiPort %s namespace %s", kubeapiHost, kubeapiPort, namespace)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
}

func deleteKubernetesNamespace(request *restful.Request, response *restful.Response) {
	kubeapiHost := request.QueryParameter("kubeapihost")
	kubeapiPortText := request.QueryParameter("kubeapiport")
	namespace := request.PathParameter("namespace")
	if kubeapiHost == "" || kubeapiPortText == "" || namespace == "" {
		errorText := fmt.Sprintf("Input text is incorrect kubeapiHost %s kubeapiHost %s namespace %s", kubeapiHost, kubeapiPortText, namespace)
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

	err = control.DeleteNamespace(kubeapiHost, kubeapiPort, namespace)

	if err != nil {
		errorText := fmt.Sprintf("Delete namespace failure kubeapiHost %s kubeapiPort %s namespace %s error %s", kubeapiHost, kubeapiPort, namespace, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
}

func returns200AllKubernetesNamesapceName(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", []string{})
}
