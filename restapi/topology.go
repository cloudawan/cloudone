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
		Do(returns200AllTopology, returns404, returns500))

	ws.Route(ws.POST("/").Filter(authorize).Filter(auditLog).To(postTopology).
		Doc("Create topology").
		Do(returns200, returns400, returns404, returns500).
		Reads(topology.Topology{}))

	ws.Route(ws.DELETE("/{topology}").Filter(authorize).Filter(auditLog).To(deleteTopology).
		Doc("Delete the topology").
		Param(ws.PathParameter("topology", "Topology name").DataType("string")).
		Do(returns200, returns404, returns500))

	ws.Route(ws.PUT("/{topology}").Filter(authorize).Filter(auditLog).To(putTopology).
		Doc("Modify topology").
		Param(ws.PathParameter("topology", "Topology name").DataType("string")).
		Do(returns200, returns400, returns404, returns500).
		Reads(topology.Topology{}))

	ws.Route(ws.GET("/{topology}").Filter(authorize).Filter(auditLog).To(getTopology).
		Doc("Get the topology").
		Param(ws.PathParameter("topology", "Topology name").DataType("string")).
		Do(returns200Topology, returns404, returns500))
}

func getAllTopology(request *restful.Request, response *restful.Response) {
	topologySlice, err := topology.GetStorage().LoadAllTopology()
	if err != nil {
		errorText := fmt.Sprintf("Could not get all topology with error %s", err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}

	response.WriteJson(topologySlice, "[]Topology")
}

func postTopology(request *restful.Request, response *restful.Response) {
	topologyInput := &topology.Topology{}
	err := request.ReadEntity(&topologyInput)

	if err != nil {
		errorText := fmt.Sprintf("POST parse topology failure with error %s", err)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	oldTopology, _ := topology.GetStorage().LoadTopology(topologyInput.Name)
	if oldTopology != nil {
		errorText := fmt.Sprintf("The topology with name %s exists", topologyInput.Name)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	err = topology.GetStorage().SaveTopology(topologyInput)
	if err != nil {
		errorText := fmt.Sprintf("Save topology %v with error %s", topologyInput, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
}

func putTopology(request *restful.Request, response *restful.Response) {
	topologyName := request.PathParameter("topology")

	topologyInput := &topology.Topology{}
	err := request.ReadEntity(&topologyInput)

	if err != nil {
		errorText := fmt.Sprintf("PUT parse topology failure with error %s", err)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	if topologyName != topologyInput.Name {
		errorText := fmt.Sprintf("PUT name %s is different from name %s in the body", topologyName, topologyInput.Name)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	oldTopology, _ := topology.GetStorage().LoadTopology(topologyInput.Name)
	if oldTopology == nil {
		errorText := fmt.Sprintf("The topology with name %s doesn't exist", topologyInput.Name)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	err = topology.GetStorage().SaveTopology(topologyInput)
	if err != nil {
		errorText := fmt.Sprintf("Save topology %v with error %s", topologyInput, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
}

func deleteTopology(request *restful.Request, response *restful.Response) {
	topologyName := request.PathParameter("topology")

	err := topology.GetStorage().DeleteTopology(topologyName)
	if err != nil {
		errorText := fmt.Sprintf("Delete topology %s with error %s", topologyName, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
}

func getTopology(request *restful.Request, response *restful.Response) {
	topologyName := request.PathParameter("topology")

	topology, err := topology.GetStorage().LoadTopology(topologyName)
	if err != nil {
		errorText := fmt.Sprintf("Could not get topology %s with error %s", topologyName, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
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
