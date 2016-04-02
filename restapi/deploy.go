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
	"github.com/cloudawan/cloudone/deploy"
	"github.com/emicklei/go-restful"
	"net/http"
	"strconv"
)

type DeployCreateInput struct {
	ImageInformationName string
	Version              string
	Description          string
	ReplicaAmount        int
	PortSlice            []control.ReplicationControllerContainerPort
	EnvironmentSlice     []control.ReplicationControllerContainerEnvironment
	NodePort             int
}

type DeployUpdateInput struct {
	ImageInformationName string
	Version              string
	Description          string
	EnvironmentSlice     []control.ReplicationControllerContainerEnvironment
}

func registerWebServiceDeploy() {
	ws := new(restful.WebService)
	ws.Path("/api/v1/deploys")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	restful.Add(ws)

	ws.Route(ws.GET("/").Filter(authorize).To(getAllDeployInformation).
		Doc("Get all of the deplpoy information").
		Do(returns200AllDeployInformation, returns404, returns500))

	ws.Route(ws.DELETE("/{namespace}/{imageinformation}").Filter(authorize).To(deleteDeployInformation).
		Doc("Delete deploy information").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.PathParameter("imageinformation", "Image information").DataType("string")).
		Param(ws.QueryParameter("kubeapihost", "Kubernetes host").DataType("string")).
		Param(ws.QueryParameter("kubeapiport", "Kubernetes port").DataType("int")).
		Do(returns200, returns404, returns500))

	ws.Route(ws.POST("/create/{namespace}").Filter(authorize).To(postDeployCreate).
		Doc("Create dployment from selected image build and version").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.QueryParameter("kubeapihost", "Kubernetes host").DataType("string")).
		Param(ws.QueryParameter("kubeapiport", "Kubernetes port").DataType("int")).
		Do(returns200, returns400, returns404, returns500).
		Reads(DeployCreateInput{}))

	ws.Route(ws.PUT("/update/{namespace}").Filter(authorize).To(putDeployUpdate).
		Doc("Update dployment from selected image build and version").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.QueryParameter("kubeapihost", "Kubernetes host").DataType("string")).
		Param(ws.QueryParameter("kubeapiport", "Kubernetes port").DataType("int")).
		Do(returns200, returns400, returns404, returns500).
		Reads(DeployUpdateInput{}))
}

func getAllDeployInformation(request *restful.Request, response *restful.Response) {
	deployInformationSlice, err := deploy.GetStorage().LoadAllDeployInformation()
	if err != nil {
		errorText := fmt.Sprintf("Get all deploy information failure %s", err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}

	response.WriteJson(deployInformationSlice, "[]DeployInformation")
}

func deleteDeployInformation(request *restful.Request, response *restful.Response) {
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

	imageInformation := request.PathParameter("imageinformation")

	err = deploy.DeployDelete(kubeapiHost, kubeapiPort, namespace, imageInformation)
	if err != nil {
		errorText := fmt.Sprintf("Delete imageInformation %s namespace %s failure %s", imageInformation, namespace, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
}

func postDeployCreate(request *restful.Request, response *restful.Response) {
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

	deployCreateInput := new(DeployCreateInput)
	err = request.ReadEntity(&deployCreateInput)

	if err != nil {
		errorText := fmt.Sprintf("POST namespace %s kubeapiHost %s kubeapiPort %s deployCreateInput %s failure with error %s", namespace, kubeapiHost, kubeapiPort, deployCreateInput, err)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	err = deploy.DeployCreate(
		kubeapiHost,
		kubeapiPort,
		namespace,
		deployCreateInput.ImageInformationName,
		deployCreateInput.Version,
		deployCreateInput.Description,
		deployCreateInput.ReplicaAmount,
		deployCreateInput.PortSlice,
		deployCreateInput.EnvironmentSlice,
		deployCreateInput.NodePort,
	)

	if err != nil {
		errorText := fmt.Sprintf("Deploy create failure kubeapiHost %s kubeapiPort %s namespace %s deploy create input %s error %s", kubeapiHost, kubeapiPort, namespace, deployCreateInput, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
}

func putDeployUpdate(request *restful.Request, response *restful.Response) {
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

	deployUpdateInput := new(DeployUpdateInput)
	err = request.ReadEntity(&deployUpdateInput)

	if err != nil {
		errorText := fmt.Sprintf("PUT namespace %s kubeapiHost %s kubeapiPort %s failure with error %s", namespace, kubeapiHost, kubeapiPort, err)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	err = deploy.DeployUpdate(
		kubeapiHost,
		kubeapiPort,
		namespace,
		deployUpdateInput.ImageInformationName,
		deployUpdateInput.Version,
		deployUpdateInput.Description,
		deployUpdateInput.EnvironmentSlice,
	)

	if err != nil {
		errorText := fmt.Sprintf("Deploy update failure kubeapiHost %s kubeapiPort %s namespace %s deploy update input %s error %s", kubeapiHost, kubeapiPort, namespace, deployUpdateInput, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
}

func returns200AllDeployInformation(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", []deploy.DeployInformation{})
}
