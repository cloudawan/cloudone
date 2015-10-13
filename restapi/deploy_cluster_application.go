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
	"github.com/cloudawan/kubernetes_management/deploy"
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

	ws.Route(ws.GET("/{namespace}/").To(getAllDeployClusterApplication).
		Doc("Get all of the cluster application deployment").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.QueryParameter("kubeapihost", "Kubernetes host").DataType("string")).
		Param(ws.QueryParameter("kubeapiport", "Kubernetes port").DataType("int")).
		Do(returns200AllDeployCluster, returns404, returns500))

	ws.Route(ws.PUT("/size/{namespace}/{clusterapplication}").To(putDeployClusterApplicationSize).
		Doc("Resize the cluster application deployment").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.PathParameter("clusterapplication", "Cluster Application name for this deployment").DataType("string")).
		Param(ws.QueryParameter("kubeapihost", "Kubernetes host").DataType("string")).
		Param(ws.QueryParameter("kubeapiport", "Kubernetes port").DataType("int")).
		Param(ws.QueryParameter("size", "Instance amount to change").DataType("int")).
		Do(returns200, returns400, returns404, returns500).
		Reads(SizeInput{}))

	ws.Route(ws.DELETE("/{namespace}/{clusterapplication}").To(deleteDeployClusterApplication).
		Doc("Delete the cluster application deployment").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.PathParameter("clusterapplication", "Cluster Application name for this deployment").DataType("string")).
		Param(ws.QueryParameter("kubeapihost", "Kubernetes host").DataType("string")).
		Param(ws.QueryParameter("kubeapiport", "Kubernetes port").DataType("int")).
		Do(returns200, returns404, returns500))
}

func getAllDeployClusterApplication(request *restful.Request, response *restful.Response) {
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

	deployClusterApplicationSlice, err := deploy.GetAllDeployClusterApplication(kubeapiHost, kubeapiPort, namespace)
	if err != nil {
		errorText := fmt.Sprintf("Get all cluster application deployment in namespace %s error %s", namespace, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}

	response.WriteJson(deployClusterApplicationSlice, "[]DeployClusterApplication")
}

func putDeployClusterApplicationSize(request *restful.Request, response *restful.Response) {
	kubeapiHost := request.QueryParameter("kubeapihost")
	kubeapiPortText := request.QueryParameter("kubeapiport")
	namespace := request.PathParameter("namespace")
	clusterapplication := request.PathParameter("clusterapplication")
	sizeText := request.QueryParameter("size")

	if kubeapiHost == "" || kubeapiPortText == "" || namespace == "" || sizeText == "" {
		errorText := fmt.Sprintf("Input text is incorrect kubeapiHost %s kubeapiPort %s namespace %s size %s", kubeapiHost, kubeapiPortText, namespace, sizeText)
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

	if err != nil {
		errorText := fmt.Sprintf("PUT namespace %s cluster application %s kubeapiHost %s kubeapiPort %s failure with error %s", namespace, clusterapplication, kubeapiHost, kubeapiPort, err)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	err = deploy.ResizeDeployClusterApplication(kubeapiHost, kubeapiPort, namespace, clusterapplication, environmentSlice, size)

	if err != nil {
		errorText := fmt.Sprintf("Resize cluster application deployment %s in namespace %s error %s", clusterapplication, namespace, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
}

func deleteDeployClusterApplication(request *restful.Request, response *restful.Response) {
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

	clusterapplication := request.PathParameter("clusterapplication")

	err = deploy.DeleteDeployClusterApplication(kubeapiHost, kubeapiPort, namespace, clusterapplication)

	if err != nil {
		errorText := fmt.Sprintf("Delete cluster application deployment %s in namespace %s error %s", clusterapplication, namespace, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
}

func returns200AllDeployCluster(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", []deploy.DeployClusterApplication{})
}
