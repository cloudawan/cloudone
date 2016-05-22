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
	"github.com/cloudawan/cloudone/utility/configuration"
	"github.com/emicklei/go-restful"
	"net/http"
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
		Do(returns200AllKubernetesNamesapceName, returns400, returns404, returns422, returns500))

	ws.Route(ws.POST("/").Filter(authorize).Filter(auditLog).To(postKubernetesNamespace).
		Doc("Add a kubernetes namespace").
		Do(returns200, returns400, returns404, returns422, returns500).
		Reads(Namesapce{}))

	ws.Route(ws.DELETE("/{namespace}").Filter(authorize).Filter(auditLog).To(deleteKubernetesNamespace).
		Doc("Delete the kubernetes namespace").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Do(returns200, returns400, returns404, returns422, returns500))
}

func getAllKubernetesNamespaceName(request *restful.Request, response *restful.Response) {
	kubeApiServerEndPoint, kubeApiServerToken, err := configuration.GetAvailablekubeApiServerEndPoint()
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get kube apiserver endpoint and token failure"
		jsonMap["ErrorMessage"] = err.Error()
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	nameSlice, err := control.GetAllNamespaceName(kubeApiServerEndPoint, kubeApiServerToken)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get all namespace failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeApiServerEndPoint"] = kubeApiServerEndPoint
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(nameSlice, "[]string")
}

func postKubernetesNamespace(request *restful.Request, response *restful.Response) {
	kubeApiServerEndPoint, kubeApiServerToken, err := configuration.GetAvailablekubeApiServerEndPoint()
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get kube apiserver endpoint and token failure"
		jsonMap["ErrorMessage"] = err.Error()
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	namespace := new(Namesapce)
	err = request.ReadEntity(&namespace)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Read body failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeApiServerEndPoint"] = kubeApiServerEndPoint
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	err = control.CreateNamespace(kubeApiServerEndPoint, kubeApiServerToken, namespace.Name)

	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Create namespace failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeApiServerEndPoint"] = kubeApiServerEndPoint
		jsonMap["namespace"] = namespace
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func deleteKubernetesNamespace(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	kubeApiServerEndPoint, kubeApiServerToken, err := configuration.GetAvailablekubeApiServerEndPoint()
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get kube apiserver endpoint and token failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["namespace"] = namespace
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	err = control.DeleteNamespace(kubeApiServerEndPoint, kubeApiServerToken, namespace)

	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Delete namespace failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeApiServerEndPoint"] = kubeApiServerEndPoint
		jsonMap["namespace"] = namespace
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func returns200AllKubernetesNamesapceName(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", []string{})
}
