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
	"github.com/cloudawan/cloudone/slb"
	"github.com/cloudawan/cloudone/utility/configuration"
	"github.com/emicklei/go-restful"
	"net/http"
)

func registerWebServiceDeployBlueGreen() {
	ws := new(restful.WebService)
	ws.Path("/api/v1/deploybluegreens")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	restful.Add(ws)

	ws.Route(ws.GET("/").Filter(authorize).Filter(auditLog).To(getAllDeployBlueGreen).
		Doc("Get all of the blue green deployment").
		Do(returns200AllDeployBlueGreen, returns404, returns500))

	ws.Route(ws.PUT("/").Filter(authorize).Filter(auditLog).To(putDeployBlueGreen).
		Doc("Update blue green dployment to switch deployment").
		Do(returns200, returns400, returns404, returns422, returns500).
		Reads(deploy.DeployBlueGreen{}))

	ws.Route(ws.DELETE("/{imageinformation}").Filter(authorize).Filter(auditLog).To(deleteDeployBlueGreen).
		Doc("Delete blue green deployment").
		Param(ws.PathParameter("imageinformation", "Image information").DataType("string")).
		Do(returns200, returns400, returns404, returns422, returns500))

	ws.Route(ws.GET("/{imageinformation}").Filter(authorize).Filter(auditLog).To(getDeployBlueGreen).
		Doc("Get the blue green deployment with the image information").
		Param(ws.PathParameter("imageinformation", "Image information").DataType("string")).
		Do(returns200DeployBlueGreen, returns404, returns500))

	ws.Route(ws.GET("/deployable/{imageinformation}").Filter(authorize).Filter(auditLog).To(getAllDeployableNamespace).
		Doc("Get all of the deployable namespace").
		Param(ws.PathParameter("imageinformation", "Image information").DataType("string")).
		Do(returns200AllDeployableNamespace, returns404, returns500))
}

func getAllDeployBlueGreen(request *restful.Request, response *restful.Response) {
	deployBlueGreenSlice, err := deploy.GetStorage().LoadAllDeployBlueGreen()
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "UpdaGet all blue green deployment failure"
		jsonMap["ErrorMessage"] = err.Error()
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(deployBlueGreenSlice, "[]DeployBlueGreen")
}

func putDeployBlueGreen(request *restful.Request, response *restful.Response) {
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

	deployBlueGreen := new(deploy.DeployBlueGreen)
	err = request.ReadEntity(&deployBlueGreen)
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

	err = deploy.UpdateDeployBlueGreen(kubeApiServerEndPoint, kubeApiServerToken, deployBlueGreen)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Update blue green deployment failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeApiServerEndPoint"] = kubeApiServerEndPoint
		jsonMap["deployBlueGreen"] = deployBlueGreen
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}

	err = slb.SendCommandToAllSLBDaemon()
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Configure SLB failure"
		jsonMap["ErrorMessage"] = err.Error()
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func deleteDeployBlueGreen(request *restful.Request, response *restful.Response) {
	imageInformation := request.PathParameter("imageinformation")

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

	err = deploy.GetStorage().DeleteDeployBlueGreen(imageInformation)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Delete blue green deployment failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeApiServerEndPoint"] = kubeApiServerEndPoint
		jsonMap["imageInformation"] = imageInformation
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}

	err = deploy.CleanAllServiceUnderBlueGreenDeployment(kubeApiServerEndPoint, kubeApiServerToken, imageInformation)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Delete all services under blue green deployment failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeApiServerEndPoint"] = kubeApiServerEndPoint
		jsonMap["imageInformation"] = imageInformation
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}

	err = slb.SendCommandToAllSLBDaemon()
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Configure SLB failure"
		jsonMap["ErrorMessage"] = err.Error()
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func getDeployBlueGreen(request *restful.Request, response *restful.Response) {
	imageInformation := request.PathParameter("imageinformation")

	deployBlueGreen, err := deploy.GetStorage().LoadDeployBlueGreen(imageInformation)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get blue green deployment failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["imageInformation"] = imageInformation
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(deployBlueGreen, "DeployBlueGreen")
}

func getAllDeployableNamespace(request *restful.Request, response *restful.Response) {
	imageInformation := request.PathParameter("imageinformation")

	namespaceSlice, err := deploy.GetAllBlueGreenDeployableNamespace(imageInformation)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get all deployable namespace failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["imageInformation"] = imageInformation
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(namespaceSlice, "[]string")

}

func returns200AllDeployBlueGreen(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", []deploy.DeployBlueGreen{})
}

func returns200DeployBlueGreen(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", deploy.DeployBlueGreen{})
}

func returns200AllDeployableNamespace(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", []string{})
}
