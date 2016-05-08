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
	"github.com/cloudawan/cloudone/topology"
	"github.com/emicklei/go-restful"
	"net/http"
)

func registerWebServiceTopology() {
	ws := new(restful.WebService)
	ws.Path("/api/v1/topology")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	restful.Add(ws)

	ws.Route(ws.GET("/").Filter(authorize).Filter(auditLog).To(getAllTopology).
		Doc("Get all of the topology").
		Do(returns200AllTopology, returns422, returns500))

	ws.Route(ws.POST("/").Filter(authorize).Filter(auditLog).To(postTopology).
		Doc("Create topology").
		Do(returns200, returns400, returns409, returns422, returns500).
		Reads(topology.Topology{}))

	ws.Route(ws.PUT("/{topology}").Filter(authorize).Filter(auditLog).To(putTopology).
		Doc("Modify topology").
		Param(ws.PathParameter("topology", "Topology name").DataType("string")).
		Do(returns200, returns400, returns404, returns422, returns500).
		Reads(topology.Topology{}))

	ws.Route(ws.DELETE("/{topology}").Filter(authorize).Filter(auditLog).To(deleteTopology).
		Doc("Delete the topology").
		Param(ws.PathParameter("topology", "Topology name").DataType("string")).
		Do(returns200, returns422, returns500))

	ws.Route(ws.GET("/{topology}").Filter(authorize).Filter(auditLog).To(getTopology).
		Doc("Get the topology").
		Param(ws.PathParameter("topology", "Topology name").DataType("string")).
		Do(returns200Topology, returns422, returns500))
}

func getAllTopology(request *restful.Request, response *restful.Response) {
	topologySlice, err := topology.GetStorage().LoadAllTopology()
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get all topology failure"
		jsonMap["ErrorMessage"] = err.Error()
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(topologySlice, "[]Topology")
}

func postTopology(request *restful.Request, response *restful.Response) {
	topologyInput := &topology.Topology{}
	err := request.ReadEntity(&topologyInput)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Read body failure"
		jsonMap["ErrorMessage"] = err.Error()
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	oldTopology, _ := topology.GetStorage().LoadTopology(topologyInput.Name)
	if oldTopology != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "The topology to create already exists"
		jsonMap["name"] = topologyInput.Name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(409, string(errorMessageByteSlice))
		return
	}

	err = topology.GetStorage().SaveTopology(topologyInput)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Save topology failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["topologyInput"] = topologyInput
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func putTopology(request *restful.Request, response *restful.Response) {
	topologyName := request.PathParameter("topology")

	topologyInput := &topology.Topology{}
	err := request.ReadEntity(&topologyInput)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Read body failure"
		jsonMap["ErrorMessage"] = err.Error()
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	if topologyName != topologyInput.Name {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Path parameter name is different from name in the body"
		jsonMap["path"] = topologyName
		jsonMap["body"] = topologyInput.Name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	oldTopology, _ := topology.GetStorage().LoadTopology(topologyInput.Name)
	if oldTopology == nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "The topology to update doesn't exist"
		jsonMap["name"] = topologyInput.Name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	err = topology.GetStorage().SaveTopology(topologyInput)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Save topology failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["topologyInput"] = topologyInput
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func deleteTopology(request *restful.Request, response *restful.Response) {
	topologyName := request.PathParameter("topology")

	err := topology.GetStorage().DeleteTopology(topologyName)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Delete topology failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["topologyName"] = topologyName
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func getTopology(request *restful.Request, response *restful.Response) {
	topologyName := request.PathParameter("topology")

	topology, err := topology.GetStorage().LoadTopology(topologyName)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get topology failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["topologyName"] = topologyName
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(topology, "Topology")
}

func returns200AllTopology(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", []topology.Topology{})
}

func returns200Topology(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", topology.Topology{})
}
