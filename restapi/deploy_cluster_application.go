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
	"github.com/cloudawan/cloudone/deploy"
	"github.com/emicklei/go-restful"
	"net/http"
	"strconv"
)

func registerWebServiceDeployClusterApplication() {
	ws := new(restful.WebService)
	ws.Path("/api/v1/deployclusterapplications")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	restful.Add(ws)

	ws.Route(ws.GET("/").Filter(authorize).Filter(auditLog).To(getAllDeployClusterApplication).
		Doc("Get all of the cluster application deployment").
		Do(returns200AllDeployCluster, returns404, returns500))

	ws.Route(ws.GET("/{namespace}/").Filter(authorize).Filter(auditLog).To(getAllDeployClusterApplicationInNamespace).
		Doc("Get all of the cluster application deployment in namespace").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Do(returns200AllDeployCluster, returns404, returns500))

	ws.Route(ws.GET("/{namespace}/{clusterapplication}").Filter(authorize).Filter(auditLog).To(getDeployClusterApplication).
		Doc("Get all of the cluster application deployment").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.PathParameter("clusterapplication", "Cluster Application name for this deployment").DataType("string")).
		Do(returns200DeployCluster, returns404, returns500))

	ws.Route(ws.PUT("/size/{namespace}/{clusterapplication}").Filter(authorize).Filter(auditLog).To(putDeployClusterApplicationSize).
		Doc("Resize the cluster application deployment").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.PathParameter("clusterapplication", "Cluster Application name for this deployment").DataType("string")).
		Param(ws.QueryParameter("kubeapihost", "Kubernetes host").DataType("string")).
		Param(ws.QueryParameter("kubeapiport", "Kubernetes port").DataType("int")).
		Param(ws.QueryParameter("size", "Instance amount to change").DataType("int")).
		Do(returns200, returns400, returns422, returns500).
		Reads(SizeInput{}))

	ws.Route(ws.DELETE("/{namespace}/{clusterapplication}").Filter(authorize).Filter(auditLog).To(deleteDeployClusterApplication).
		Doc("Delete the cluster application deployment").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.PathParameter("clusterapplication", "Cluster Application name for this deployment").DataType("string")).
		Param(ws.QueryParameter("kubeapihost", "Kubernetes host").DataType("string")).
		Param(ws.QueryParameter("kubeapiport", "Kubernetes port").DataType("int")).
		Do(returns200, returns422, returns500))
}

func getAllDeployClusterApplication(request *restful.Request, response *restful.Response) {
	deployClusterApplicationSlice, err := deploy.GetAllDeployClusterApplication()
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get all cluster application deployment"
		jsonMap["ErrorMessage"] = err.Error()
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(deployClusterApplicationSlice, "[]DeployClusterApplication")
}

func getAllDeployClusterApplicationInNamespace(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")

	deployClusterApplicationSlice, err := deploy.GetAllDeployClusterApplicationInNamespace(namespace)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get all cluster application deployment in the namespace failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["namespace"] = namespace
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(deployClusterApplicationSlice, "[]DeployClusterApplication")
}

func getDeployClusterApplication(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	clusterApplication := request.PathParameter("clusterapplication")

	deployClusterApplication, err := deploy.GetDeployClusterApplication(namespace, clusterApplication)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get cluster application deployment failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["namespace"] = namespace
		jsonMap["clusterApplication"] = clusterApplication
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(deployClusterApplication, "DeployClusterApplication")
}

func putDeployClusterApplicationSize(request *restful.Request, response *restful.Response) {
	kubeapiHost := request.QueryParameter("kubeapihost")
	kubeapiPortText := request.QueryParameter("kubeapiport")
	namespace := request.PathParameter("namespace")
	clusterApplication := request.PathParameter("clusterapplication")
	sizeText := request.QueryParameter("size")
	if kubeapiHost == "" || kubeapiPortText == "" || sizeText == "" {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Input is incorrect. The fields kubeapihost, kubeapiport, and sizeText are required."
		jsonMap["kubeapiHost"] = kubeapiHost
		jsonMap["kubeapiPortText"] = kubeapiPortText
		jsonMap["sizeText"] = sizeText
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
	size, err := strconv.Atoi(sizeText)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Could not parse sizeText"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["sizeText"] = sizeText
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	environmentSlice := make([]interface{}, 0)
	err = request.ReadEntity(&environmentSlice)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Read body failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeapiHost"] = kubeapiHost
		jsonMap["kubeapiPort"] = kubeapiPort
		jsonMap["namespace"] = namespace
		jsonMap["clusterApplication"] = namespace
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	err = deploy.ResizeDeployClusterApplication(kubeapiHost, kubeapiPort, namespace, clusterApplication, environmentSlice, size)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Resize cluster application deployment failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeapiHost"] = kubeapiHost
		jsonMap["kubeapiPort"] = kubeapiPort
		jsonMap["namespace"] = namespace
		jsonMap["clusterApplication"] = clusterApplication
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func deleteDeployClusterApplication(request *restful.Request, response *restful.Response) {
	kubeapiHost := request.QueryParameter("kubeapihost")
	kubeapiPortText := request.QueryParameter("kubeapiport")
	namespace := request.PathParameter("namespace")
	clusterApplication := request.PathParameter("clusterapplication")
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

	err = deploy.DeleteDeployClusterApplication(kubeapiHost, kubeapiPort, namespace, clusterApplication)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Delete cluster application deployment failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeapiHost"] = kubeapiHost
		jsonMap["kubeapiPort"] = kubeapiPort
		jsonMap["namespace"] = namespace
		jsonMap["clusterApplication"] = clusterApplication
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func returns200AllDeployCluster(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", []deploy.DeployClusterApplication{})
}

func returns200DeployCluster(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", deploy.DeployClusterApplication{})
}
