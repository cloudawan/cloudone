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
	"github.com/cloudawan/cloudone/application"
	"github.com/cloudawan/cloudone/deploy"
	"github.com/emicklei/go-restful"
	"net/http"
	"strconv"
)

type ClusterDescription struct {
	Name                      string
	Description               string
	ReplicationControllerJson string
	ServiceJson               string
	Environment               interface{}
	ScriptType                string
	ScriptContent             string
}

type ClusterLaunch struct {
	Size                              int
	EnvironmentSlice                  []interface{}
	ReplicationControllerExtraJsonMap map[string]interface{}
}

func registerWebServiceClusterApplication() {
	ws := new(restful.WebService)
	ws.Path("/api/v1/clusterapplications")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	restful.Add(ws)

	ws.Route(ws.GET("/").Filter(authorize).Filter(auditLog).To(getAllClusterApplication).
		Doc("Get all of the configuration of cluster application").
		Do(returns200AllClusterApplication, returns404, returns500))

	ws.Route(ws.GET("/{clusterapplication}").Filter(authorize).Filter(auditLog).To(getClusterApplication).
		Doc("Get the configuration of cluster application").
		Param(ws.PathParameter("clusterapplication", "Cluster application name").DataType("string")).
		Do(returns200ClusterApplication, returns404, returns500))

	ws.Route(ws.POST("/").Filter(authorize).Filter(auditLog).To(postClusterApplication).
		Doc("Add a cluster application").
		Do(returns200, returns400, returns422, returns500).
		Reads(ClusterDescription{}))

	ws.Route(ws.DELETE("/{clusterapplication}").Filter(authorize).Filter(auditLog).To(deleteClusterApplication).
		Doc("Delete an cluster application").
		Param(ws.PathParameter("clusterapplication", "Cluster application name").DataType("string")).
		Do(returns200, returns404, returns500))

	ws.Route(ws.POST("/launch/{namespace}/{clusterapplication}").Filter(authorize).Filter(auditLog).To(postLaunchClusterApplication).
		Doc("Launch a cluster application").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.PathParameter("clusterapplication", "cluster application").DataType("string")).
		Param(ws.QueryParameter("kubeapihost", "Kubernetes host").DataType("string")).
		Param(ws.QueryParameter("kubeapiport", "Kubernetes port").DataType("int")).
		Param(ws.QueryParameter("size", "How many instances to launch").DataType("int")).
		Do(returns200, returns400, returns409, returns422, returns500).
		Reads(ClusterLaunch{}))
}

func getAllClusterApplication(request *restful.Request, response *restful.Response) {
	clusterSlice, err := application.GetStorage().LoadAllClusterApplication()
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get all cluster application failure"
		jsonMap["ErrorMessage"] = err.Error()
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(clusterSlice, "[]Cluster")
}

func getClusterApplication(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("clusterapplication")

	cluster, err := application.GetStorage().LoadClusterApplication(name)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get cluster application failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["cluster"] = cluster
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(cluster, "Cluster")
}

func postClusterApplication(request *restful.Request, response *restful.Response) {
	cluster := &application.Cluster{}
	err := request.ReadEntity(&cluster)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Read body failure"
		jsonMap["ErrorMessage"] = err.Error()
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	err = application.GetStorage().SaveClusterApplication(cluster)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Save cluster application failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["cluster"] = cluster
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func deleteClusterApplication(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("clusterapplication")
	err := application.GetStorage().DeleteClusterApplication(name)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Delete cluster application failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["name"] = name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}
}

func postLaunchClusterApplication(request *restful.Request, response *restful.Response) {
	kubeapiHost := request.QueryParameter("kubeapihost")
	kubeapiPortText := request.QueryParameter("kubeapiport")
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("clusterapplication")
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
		jsonMap["name"] = name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	clusterLaunch := ClusterLaunch{}
	err = request.ReadEntity(&clusterLaunch)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Read body failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeapiHost"] = kubeapiHost
		jsonMap["kubeapiPort"] = kubeapiPort
		jsonMap["namespace"] = namespace
		jsonMap["name"] = name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	oldDeployClusterApplication, _ := deploy.GetDeployClusterApplication(namespace, name)
	if oldDeployClusterApplication != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "The cluster application already exists"
		jsonMap["kubeapiHost"] = kubeapiHost
		jsonMap["kubeapiPort"] = kubeapiPort
		jsonMap["namespace"] = namespace
		jsonMap["name"] = name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(409, string(errorMessageByteSlice))
		return
	}

	err = application.LaunchClusterApplication(kubeapiHost, kubeapiPort, namespace, name, clusterLaunch.EnvironmentSlice, clusterLaunch.Size, clusterLaunch.ReplicationControllerExtraJsonMap)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Launch cluster application deployment failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeapiHost"] = kubeapiHost
		jsonMap["kubeapiPort"] = kubeapiPort
		jsonMap["namespace"] = namespace
		jsonMap["name"] = name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}

	err = deploy.InitializeDeployClusterApplication(kubeapiHost, kubeapiPort, namespace, name, clusterLaunch.EnvironmentSlice, clusterLaunch.Size, clusterLaunch.ReplicationControllerExtraJsonMap)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Create cluster application deployment failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeapiHost"] = kubeapiHost
		jsonMap["kubeapiPort"] = kubeapiPort
		jsonMap["namespace"] = namespace
		jsonMap["name"] = name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func returns200AllClusterApplication(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", []ClusterDescription{})
}

func returns200ClusterApplication(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", ClusterDescription{})
}
