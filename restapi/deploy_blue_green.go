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
	"github.com/cloudawan/cloudone/deploy"
	"github.com/emicklei/go-restful"
	"net/http"
	"strconv"
)

func registerWebServiceDeployBlueGreen() {
	ws := new(restful.WebService)
	ws.Path("/api/v1/deploybluegreens")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	restful.Add(ws)

	ws.Route(ws.GET("/").To(getAllDeployBlueGreen).
		Doc("Get all of the blue green deployment").
		Do(returns200AllDeployBlueGreen, returns404, returns500))

	ws.Route(ws.DELETE("/{imageinformation}").To(deleteDeployBlueGreen).
		Doc("Delete blue green deployment").
		Param(ws.PathParameter("imageinformation", "Image information").DataType("string")).
		Param(ws.QueryParameter("kubeapihost", "Kubernetes host").DataType("string")).
		Param(ws.QueryParameter("kubeapiport", "Kubernetes port").DataType("int")).
		Do(returns200, returns404, returns500))

	ws.Route(ws.PUT("/").To(putDeployBlueGreen).
		Doc("Update blue green dployment to switch deployment").
		Param(ws.QueryParameter("kubeapihost", "Kubernetes host").DataType("string")).
		Param(ws.QueryParameter("kubeapiport", "Kubernetes port").DataType("int")).
		Do(returns200, returns400, returns404, returns500).
		Reads(deploy.DeployBlueGreen{}))

	ws.Route(ws.GET("/deployable/{imageinformation}").To(getAllDeployableNamespace).
		Doc("Get all of the deployable namespace").
		Param(ws.PathParameter("imageinformation", "Image information").DataType("string")).
		Do(returns200AllDeployableNamespace, returns404, returns500))
}

func getAllDeployBlueGreen(request *restful.Request, response *restful.Response) {
	deployBlueGreenSlice, err := deploy.LoadAllDeployBlueGreen()
	if err != nil {
		errorText := fmt.Sprintf("Get all blue green deployment failure %s", err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}

	response.WriteJson(deployBlueGreenSlice, "[]DeployBlueGreen")
}

func deleteDeployBlueGreen(request *restful.Request, response *restful.Response) {
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

	imageInformation := request.PathParameter("imageinformation")

	err = deploy.DeleteDeployBlueGreen(imageInformation)
	if err != nil {
		errorText := fmt.Sprintf("Delete blue green deployment imageInformation %s failure %s", imageInformation, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}

	err = deploy.CleanAllServiceUnderBlueGreenDeployment(kubeapiHost, kubeapiPort, imageInformation)
	if err != nil {
		errorText := fmt.Sprintf("Delete blue green deployment service on Kubernetes failure imageInformation %s error %s", imageInformation, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
}

func putDeployBlueGreen(request *restful.Request, response *restful.Response) {
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

	deployBlueGreen := new(deploy.DeployBlueGreen)
	err = request.ReadEntity(&deployBlueGreen)

	if err != nil {
		errorText := fmt.Sprintf("PUT kubeapiHost %s kubeapiPort %s failure with error %s", kubeapiHost, kubeapiPort, err)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	err = deploy.UpdateDeployBlueGreen(kubeapiHost, kubeapiPort, deployBlueGreen)
	if err != nil {
		errorText := fmt.Sprintf("Update blue green deployment failure kubeapiHost %s kubeapiPort %s deploy update input %s error %s", kubeapiHost, kubeapiPort, deployBlueGreen, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
}

func getAllDeployableNamespace(request *restful.Request, response *restful.Response) {
	imageInformation := request.PathParameter("imageinformation")

	namespaceSlice, err := deploy.GetAllBlueGreenDeployableNamespace(imageInformation)
	if err != nil {
		errorText := fmt.Sprintf("Get all deployable namespace failure %s", err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}

	response.WriteJson(namespaceSlice, "[]string")

}

func returns200AllDeployBlueGreen(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", []deploy.DeployBlueGreen{})
}

func returns200AllDeployableNamespace(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", []string{})
}
