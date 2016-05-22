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

func registerWebServicePod() {
	ws := new(restful.WebService)
	ws.Path("/api/v1/pods")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	restful.Add(ws)

	ws.Route(ws.DELETE("/{namespace}/{pod}/").Filter(authorize).Filter(auditLog).To(deletePod).
		Doc("Delete the pod in the namespace").
		Param(ws.PathParameter("namespace", "Namespace name").DataType("string")).
		Param(ws.PathParameter("pod", "Pod name").DataType("string")).
		Do(returns200, returns400, returns404, returns422, returns500))

	ws.Route(ws.GET("/{namespace}/{pod}/logs").Filter(authorize).Filter(auditLog).To(getPodLog).
		Doc("Get log for pod").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.PathParameter("pod", "Kubernetes pod").DataType("string")).
		Do(returns200PodLog, returns400, returns404, returns422, returns500))
}

func deletePod(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	pod := request.PathParameter("pod")

	kubeApiServerEndPoint, kubeApiServerToken, err := configuration.GetAvailablekubeApiServerEndPoint()
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get kube apiserver endpoint and token failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["namespace"] = namespace
		jsonMap["pod"] = pod
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	err = control.DeletePod(kubeApiServerEndPoint, kubeApiServerToken, namespace, pod)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Delete pod failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeApiServerEndPoint"] = kubeApiServerEndPoint
		jsonMap["namespace"] = namespace
		jsonMap["pod"] = pod
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func getPodLog(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	pod := request.PathParameter("pod")

	kubeApiServerEndPoint, kubeApiServerToken, err := configuration.GetAvailablekubeApiServerEndPoint()
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get kube apiserver endpoint and token failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["namespace"] = namespace
		jsonMap["pod"] = pod
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	logJsonMap, err := control.GetPodLog(kubeApiServerEndPoint, kubeApiServerToken, namespace, pod)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get pod log failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeApiServerEndPoint"] = kubeApiServerEndPoint
		jsonMap["namespace"] = namespace
		jsonMap["pod"] = pod
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(logJsonMap, "{}")
}

func returns200PodLog(b *restful.RouteBuilder) {
	jsonMap := make(map[string]interface{})
	b.Returns(http.StatusOK, "OK", jsonMap)
}
