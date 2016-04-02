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

type StatelessSerializableDescription struct {
	Name                      string
	Description               string
	ReplicationControllerJson interface{}
	ServiceJson               interface{}
	Environment               interface{}
}

func registerWebServiceStatelessApplication() {
	ws := new(restful.WebService)
	ws.Path("/api/v1/statelessapplications")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	restful.Add(ws)

	ws.Route(ws.GET("/").Filter(authorize).To(getAllStatelessApplication).
		Doc("Get all of the configuration of stateless application").
		Do(returns200AllStatelessApplication, returns404, returns500))

	ws.Route(ws.GET("/{statelessapplication}").Filter(authorize).To(getStatelessApplication).
		Doc("Get the configuration of stateless application").
		Param(ws.PathParameter("statelessapplication", "Stateless application name").DataType("string")).
		Do(returns200StatelessApplication, returns404, returns500))

	ws.Route(ws.POST("/").Filter(authorize).To(postStatelessApplication).
		Doc("Add a stateless application").
		Do(returns200, returns400, returns404, returns500).
		Reads(StatelessSerializableDescription{}))

	ws.Route(ws.DELETE("/{statelessapplication}").Filter(authorize).To(deleteStatelessApplication).
		Doc("Delete an stateless application").
		Param(ws.PathParameter("statelessapplication", "Stateless application name").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.POST("/launch/{namespace}/{statelessapplication}").Filter(authorize).To(postLaunchStatelessApplication).
		Doc("Launch a stateless application").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.PathParameter("statelessapplication", "stateless application").DataType("string")).
		Param(ws.QueryParameter("kubeapihost", "Kubernetes host").DataType("string")).
		Param(ws.QueryParameter("kubeapiport", "Kubernetes port").DataType("int")).
		Do(returns200, returns400, returns404, returns500).
		Reads(new(struct{})))
}

func getAllStatelessApplication(request *restful.Request, response *restful.Response) {
	statelessSlice, err := application.GetStorage().LoadAllStatelessApplication()
	if err != nil {
		errorText := fmt.Sprintf("Get read database fail with error %s", err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
	response.WriteJson(statelessSlice, "[]Stateless")
}

func getStatelessApplication(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("statelessapplication")

	statelessSerializable, err := application.RetrieveStatelessApplication(name)
	if err != nil {
		errorText := fmt.Sprintf("Get read database fail with error %s", err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}

	response.WriteJson(statelessSerializable, "Stateless")
}

func postStatelessApplication(request *restful.Request, response *restful.Response) {
	statelessSerializable := new(application.StatelessSerializable)
	err := request.ReadEntity(&statelessSerializable)

	if err != nil {
		errorText := fmt.Sprintf("POST read body failure with error %s", err)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	err = application.StoreStatelessApplication(statelessSerializable.Name, statelessSerializable.Description,
		statelessSerializable.ReplicationControllerJson, statelessSerializable.ServiceJson,
		statelessSerializable.Environment)
	if err != nil {
		errorText := fmt.Sprintf("POST fail to save %s to database with error %s", statelessSerializable, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
}

func deleteStatelessApplication(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("statelessapplication")
	err := application.GetStorage().DeleteStatelessApplication(name)
	if err != nil {
		errorText := fmt.Sprintf("Delete stateless application %s fail with error %s", name, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
}

func postLaunchStatelessApplication(request *restful.Request, response *restful.Response) {
	kubeapiHost := request.QueryParameter("kubeapihost")
	kubeapiPortText := request.QueryParameter("kubeapiport")
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("statelessapplication")

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

	environmentSlice := make([]interface{}, 0)
	err = request.ReadEntity(&environmentSlice)

	exist, err := monitor.ExistReplicationController(kubeapiHost, kubeapiPort, namespace, name)
	if exist {
		errorText := fmt.Sprintf("Replication controller already exists kubeapiHost %s, kubeapiPort %d, namespace %s, name %s", kubeapiHost, kubeapiPort, namespace, name)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	err = application.LaunchStatelessApplication(kubeapiHost, kubeapiPort, namespace, name, environmentSlice)
	if err != nil {
		errorText := fmt.Sprintf("Could not launch stateless application %s with kubeapiHost %s, kubeapiPort %d, namespace %s, error %s", name, kubeapiHost, kubeapiPort, namespace, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
}

func returns200AllStatelessApplication(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", []application.Stateless{})
}

func returns200StatelessApplication(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", StatelessSerializableDescription{})
}
