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
	"github.com/cloudawan/cloudone/monitor"
	"github.com/cloudawan/cloudone/utility/configuration"
	"github.com/emicklei/go-restful"
	"net/http"
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

	ws.Route(ws.GET("/").Filter(authorize).Filter(auditLog).To(getAllStatelessApplication).
		Doc("Get all of the configuration of stateless application").
		Do(returns200AllStatelessApplication, returns404, returns500))

	ws.Route(ws.GET("/{statelessapplication}").Filter(authorize).Filter(auditLog).To(getStatelessApplication).
		Doc("Get the configuration of stateless application").
		Param(ws.PathParameter("statelessapplication", "Stateless application name").DataType("string")).
		Do(returns200StatelessApplication, returns404, returns500))

	ws.Route(ws.POST("/").Filter(authorize).Filter(auditLog).To(postStatelessApplication).
		Doc("Add a stateless application").
		Do(returns200, returns400, returns404, returns500).
		Reads(StatelessSerializableDescription{}))

	ws.Route(ws.DELETE("/{statelessapplication}").Filter(authorize).Filter(auditLog).To(deleteStatelessApplication).
		Doc("Delete an stateless application").
		Param(ws.PathParameter("statelessapplication", "Stateless application name").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.POST("/launch/{namespace}/{statelessapplication}").Filter(authorize).Filter(auditLog).To(postLaunchStatelessApplication).
		Doc("Launch a stateless application").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.PathParameter("statelessapplication", "stateless application").DataType("string")).
		Do(returns200, returns400, returns404, returns500).
		Reads(new(struct{})))
}

func getAllStatelessApplication(request *restful.Request, response *restful.Response) {
	statelessSlice, err := application.GetStorage().LoadAllStatelessApplication()
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get all stateless application failure"
		jsonMap["ErrorMessage"] = err.Error()
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}
	response.WriteJson(statelessSlice, "[]Stateless")
}

func getStatelessApplication(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("statelessapplication")

	statelessSerializable, err := application.RetrieveStatelessApplication(name)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get stateless application failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["name"] = name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(statelessSerializable, "Stateless")
}

func postStatelessApplication(request *restful.Request, response *restful.Response) {
	statelessSerializable := new(application.StatelessSerializable)
	err := request.ReadEntity(&statelessSerializable)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Read body failure"
		jsonMap["ErrorMessage"] = err.Error()
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	err = application.StoreStatelessApplication(statelessSerializable.Name, statelessSerializable.Description,
		statelessSerializable.ReplicationControllerJson, statelessSerializable.ServiceJson,
		statelessSerializable.Environment)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Save stateless application failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["statelessSerializable"] = statelessSerializable
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func deleteStatelessApplication(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("statelessapplication")
	err := application.GetStorage().DeleteStatelessApplication(name)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Delete stateless application failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["name"] = name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}
}

func postLaunchStatelessApplication(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	name := request.PathParameter("statelessapplication")

	kubeApiServerEndPoint, kubeApiServerToken, err := configuration.GetAvailablekubeApiServerEndPoint()
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get kube apiserver endpoint and token failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["namespace"] = namespace
		jsonMap["name"] = name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	environmentSlice := make([]interface{}, 0)
	err = request.ReadEntity(&environmentSlice)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Read body failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeApiServerEndPoint"] = kubeApiServerEndPoint
		jsonMap["namespace"] = namespace
		jsonMap["name"] = name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	exist, err := monitor.ExistReplicationController(kubeApiServerEndPoint, kubeApiServerToken, namespace, name)
	if exist {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "The replication controller to use already exists"
		jsonMap["kubeApiServerEndPoint"] = kubeApiServerEndPoint
		jsonMap["namespace"] = namespace
		jsonMap["name"] = name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(409, string(errorMessageByteSlice))
		return
	}

	err = application.LaunchStatelessApplication(kubeApiServerEndPoint, kubeApiServerToken, namespace, name, environmentSlice)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Launch stateless application failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeApiServerEndPoint"] = kubeApiServerEndPoint
		jsonMap["namespace"] = namespace
		jsonMap["name"] = name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func returns200AllStatelessApplication(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", []application.Stateless{})
}

func returns200StatelessApplication(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", StatelessSerializableDescription{})
}
