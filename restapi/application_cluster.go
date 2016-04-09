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
	"github.com/cloudawan/cloudone/application"
	"github.com/cloudawan/cloudone/monitor"
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
		Do(returns200, returns400, returns404, returns500).
		Reads(ClusterDescription{}))

	ws.Route(ws.DELETE("/{clusterapplication}").Filter(authorize).Filter(auditLog).To(deleteClusterApplication).
		Doc("Delete an cluster application").
		Param(ws.PathParameter("clusterapplication", "Cluster application name").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.POST("/launch/{namespace}/{clusterapplication}").Filter(authorize).Filter(auditLog).To(postLaunchClusterApplication).
		Doc("Launch a cluster application").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.PathParameter("clusterapplication", "cluster application").DataType("string")).
		Param(ws.QueryParameter("kubeapihost", "Kubernetes host").DataType("string")).
		Param(ws.QueryParameter("kubeapiport", "Kubernetes port").DataType("int")).
		Param(ws.QueryParameter("size", "How many instances to launch").DataType("int")).
		Do(returns200, returns400, returns404, returns500).
		Reads(new(struct{})))
}

func getAllClusterApplication(request *restful.Request, response *restful.Response) {
	clusterSlice, err := application.GetStorage().LoadAllClusterApplication()
	if err != nil {
		errorText := fmt.Sprintf("Get read database fail with error %s", err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
	response.WriteJson(clusterSlice, "[]Cluster")
}

func getClusterApplication(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("clusterapplication")

	cluster, err := application.GetStorage().LoadClusterApplication(name)
	if err != nil {
		errorText := fmt.Sprintf("Get read database fail with error %s", err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}

	response.WriteJson(cluster, "Cluster")
}

func postClusterApplication(request *restful.Request, response *restful.Response) {
	cluster := new(application.Cluster)
	err := request.ReadEntity(&cluster)

	if err != nil {
		errorText := fmt.Sprintf("POST read body failure with error %s", err)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	err = application.GetStorage().SaveClusterApplication(cluster)
	if err != nil {
		errorText := fmt.Sprintf("POST fail to save %s to database with error %s", cluster, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
}

func deleteClusterApplication(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("clusterapplication")
	err := application.GetStorage().DeleteClusterApplication(name)
	if err != nil {
		errorText := fmt.Sprintf("Delete cluster application %s fail with error %s", name, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
}

func postLaunchClusterApplication(request *restful.Request, response *restful.Response) {
	kubeapiHost := request.QueryParameter("kubeapihost")
	kubeapiPortText := request.QueryParameter("kubeapiport")
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("clusterapplication")
	sizeText := request.QueryParameter("size")

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
	size, err := strconv.Atoi(sizeText)
	if err != nil {
		errorText := fmt.Sprintf("Could not parse sizeText %s with error %s", sizeText, err)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	environmentSlice := make([]interface{}, 0)
	err = request.ReadEntity(&environmentSlice)

	exist, err := monitor.ExistReplicationController(kubeapiHost, kubeapiPort, namespace, name)
	if exist {
		errorText := fmt.Sprintf("Replication controller already exists kubeapiHost %s, kubeapiPort %d, namespace %s, name %s", kubeapiHost, kubeapiPort, namespace, name)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	err = application.LaunchClusterApplication(kubeapiHost, kubeapiPort, namespace, name, environmentSlice, size)
	if err != nil {
		errorText := fmt.Sprintf("Could not launch cluster application %s with kubeapiHost %s, kubeapiPort %d, namespace %s, error %s", name, kubeapiHost, kubeapiPort, namespace, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
}

func returns200AllClusterApplication(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", []ClusterDescription{})
}

func returns200ClusterApplication(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", ClusterDescription{})
}
