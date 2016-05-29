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
	"github.com/cloudawan/cloudone/deploy"
	"github.com/cloudawan/cloudone/utility/configuration"
	"github.com/emicklei/go-restful"
	"net/http"
	"strconv"
)

type DeployCreateInput struct {
	ImageInformationName  string
	Version               string
	Description           string
	ReplicaAmount         int
	PortSlice             []deploy.DeployContainerPort
	EnvironmentSlice      []control.ReplicationControllerContainerEnvironment
	ResourceMap           map[string]interface{}
	ExtraJsonMap          map[string]interface{}
	AutoUpdateForNewBuild bool
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

	ws.Route(ws.GET("/").Filter(authorize).Filter(auditLog).To(getAllDeployInformation).
		Doc("Get all of the deplpoy information").
		Do(returns200AllDeployInformation, returns404, returns500))

	ws.Route(ws.GET("/{namespace}").Filter(authorize).Filter(auditLog).To(getDeployInformationInNamespace).
		Doc("Get all of the deplpoy information in the namespace").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Do(returns200AllDeployInformation, returns404, returns500))

	ws.Route(ws.GET("/{namespace}/{imageinformation}").Filter(authorize).Filter(auditLog).To(getDeployInformation).
		Doc("Get the deplpoy information").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.PathParameter("imageinformation", "Image information").DataType("string")).
		Do(returns200DeployInformation, returns404, returns500))

	ws.Route(ws.DELETE("/{namespace}/{imageinformation}").Filter(authorize).Filter(auditLog).To(deleteDeployInformation).
		Doc("Delete deploy information").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.PathParameter("imageinformation", "Image information").DataType("string")).
		Do(returns200, returns400, returns404, returns422, returns500))

	ws.Route(ws.POST("/create/{namespace}").Filter(authorize).Filter(auditLog).To(postDeployCreate).
		Doc("Create dployment from selected image build and version").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Do(returns200, returns400, returns404, returns422, returns500).
		Reads(DeployCreateInput{}))

	ws.Route(ws.PUT("/update/{namespace}").Filter(authorize).Filter(auditLog).To(putDeployUpdate).
		Doc("Update dployment from selected image build and version").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Do(returns200, returns400, returns404, returns422, returns500).
		Reads(DeployUpdateInput{}))

	ws.Route(ws.PUT("/resize/{namespace}/{imageinformation}").Filter(authorize).Filter(auditLog).To(putDeployResize).
		Doc("Resize dployment from selected image build and version").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.PathParameter("imageinformation", "Image information").DataType("string")).
		Param(ws.QueryParameter("size", "Size").DataType("int")).
		Do(returns200, returns400, returns404, returns422, returns500))
}

func getAllDeployInformation(request *restful.Request, response *restful.Response) {
	deployInformationSlice, err := deploy.GetStorage().LoadAllDeployInformation()
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get all deployment failure"
		jsonMap["ErrorMessage"] = err.Error()
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(deployInformationSlice, "[]DeployInformation")
}

func getDeployInformationInNamespace(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")

	deployInformationSlice, err := deploy.GetDeployInformationInNamespace(namespace)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get all deployment in the namespace failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["namespace"] = namespace
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(deployInformationSlice, "[]DeployInformation")
}

func getDeployInformation(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	imageInformation := request.PathParameter("imageinformation")

	deployInformation, err := deploy.GetStorage().LoadDeployInformation(namespace, imageInformation)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get deployment failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["namespace"] = namespace
		jsonMap["imageInformation"] = imageInformation
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(deployInformation, "DeployInformation")
}

func deleteDeployInformation(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	imageInformation := request.PathParameter("imageinformation")

	kubeApiServerEndPoint, kubeApiServerToken, err := configuration.GetAvailablekubeApiServerEndPoint()
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get kube apiserver endpoint and token failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["namespace"] = namespace
		jsonMap["imageInformation"] = imageInformation
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	err = deploy.DeployDelete(kubeApiServerEndPoint, kubeApiServerToken, namespace, imageInformation)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Delete deployment failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeApiServerEndPoint"] = kubeApiServerEndPoint
		jsonMap["namespace"] = namespace
		jsonMap["imageInformation"] = imageInformation
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func postDeployCreate(request *restful.Request, response *restful.Response) {
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

	deployCreateInput := new(DeployCreateInput)
	err = request.ReadEntity(&deployCreateInput)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Read body failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeApiServerEndPoint"] = kubeApiServerEndPoint
		jsonMap["namespace"] = namespace
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	deploymentInformation, _ := deploy.GetStorage().LoadDeployInformation(namespace, deployCreateInput.ImageInformationName)
	if deploymentInformation != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Duplicate deployment error"
		jsonMap["ErrorMessage"] = "Already exists"
		jsonMap["kubeApiServerToken"] = kubeApiServerToken
		jsonMap["namespace"] = namespace
		jsonMap["deployCreateInput"] = deployCreateInput
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}

	err = deploy.DeployCreate(
		kubeApiServerEndPoint,
		kubeApiServerToken,
		namespace,
		deployCreateInput.ImageInformationName,
		deployCreateInput.Version,
		deployCreateInput.Description,
		deployCreateInput.ReplicaAmount,
		deployCreateInput.PortSlice,
		deployCreateInput.EnvironmentSlice,
		deployCreateInput.ResourceMap,
		deployCreateInput.ExtraJsonMap,
		deployCreateInput.AutoUpdateForNewBuild,
	)

	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Create deployment failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeApiServerToken"] = kubeApiServerToken
		jsonMap["namespace"] = namespace
		jsonMap["deployCreateInput"] = deployCreateInput
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func putDeployUpdate(request *restful.Request, response *restful.Response) {
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

	deployUpdateInput := new(DeployUpdateInput)
	err = request.ReadEntity(&deployUpdateInput)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Read body failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeApiServerEndPoint"] = kubeApiServerEndPoint
		jsonMap["namespace"] = namespace
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	err = deploy.DeployUpdate(
		kubeApiServerEndPoint,
		kubeApiServerToken,
		namespace,
		deployUpdateInput.ImageInformationName,
		deployUpdateInput.Version,
		deployUpdateInput.Description,
		deployUpdateInput.EnvironmentSlice,
	)

	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Update deployment failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeApiServerEndPoint"] = kubeApiServerEndPoint
		jsonMap["namespace"] = namespace
		jsonMap["deployUpdateInput"] = deployUpdateInput
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func putDeployResize(request *restful.Request, response *restful.Response) {
	sizeText := request.QueryParameter("size")
	namespace := request.PathParameter("namespace")
	imageinformation := request.PathParameter("imageinformation")

	kubeApiServerEndPoint, kubeApiServerToken, err := configuration.GetAvailablekubeApiServerEndPoint()
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get kube apiserver endpoint and token failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["namespace"] = namespace
		jsonMap["imageinformation"] = imageinformation
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	if sizeText == "" {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Input is incorrect. The fields sizeText is required."
		jsonMap["sizeText"] = sizeText
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

	err = deploy.DeployResize(kubeApiServerEndPoint, kubeApiServerToken, namespace, imageinformation, size)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Resize deployment failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeApiServerEndPoint"] = kubeApiServerEndPoint
		jsonMap["namespace"] = namespace
		jsonMap["imageinformation"] = imageinformation
		jsonMap["size"] = size
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func returns200AllDeployInformation(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", []deploy.DeployInformation{})
}

func returns200DeployInformation(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", deploy.DeployInformation{})
}
