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
	"github.com/cloudawan/cloudone/slb"
	"github.com/emicklei/go-restful"
	"net/http"
)

func registerWebServiceSLB() {
	ws := new(restful.WebService)
	ws.Path("/api/v1/slbs")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	restful.Add(ws)

	ws.Route(ws.GET("/daemons/").Filter(authorize).Filter(auditLog).To(getAllSLBDaemon).
		Doc("Get all of the slb daemon").
		Do(returns200AllSLBDaemon, returns422, returns500))

	ws.Route(ws.POST("/daemons/").Filter(authorize).Filter(auditLogWithoutBody).To(postSLBDaemon).
		Doc("Create the slb daemon").
		Do(returns200, returns400, returns409, returns422, returns500).
		Reads(slb.SLBDaemon{}))

	ws.Route(ws.DELETE("/daemons/{name}").Filter(authorize).Filter(auditLog).To(deleteSLBDaemon).
		Doc("Delete the slb daemon").
		Param(ws.PathParameter("name", "Name").DataType("string")).
		Do(returns200, returns422, returns500))

	ws.Route(ws.PUT("/daemons/{name}").Filter(authorize).Filter(auditLogWithoutBody).To(putSLBDaemon).
		Doc("Modify the slb daemon").
		Param(ws.PathParameter("name", "Name").DataType("string")).
		Do(returns200, returns400, returns404, returns422, returns500).
		Reads(slb.SLBDaemon{}))

	ws.Route(ws.GET("/daemons/{name}").Filter(authorize).Filter(auditLog).To(getSLBDaemon).
		Doc("Get all of the slb daemons").
		Param(ws.PathParameter("name", "Name").DataType("string")).
		Do(returns200SLBDaemon, returns422, returns500))
}

func getAllSLBDaemon(request *restful.Request, response *restful.Response) {
	slbDaemonClusterSlice, err := slb.GetStorage().LoadAllSLBDaemon()
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get all slbDaemon failure"
		jsonMap["ErrorMessage"] = err.Error()
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(slbDaemonClusterSlice, "[]SLBDaemon")
}

func postSLBDaemon(request *restful.Request, response *restful.Response) {
	slbDaemon := slb.SLBDaemon{}
	err := request.ReadEntity(&slbDaemon)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Read body failure"
		jsonMap["ErrorMessage"] = err.Error()
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	oldSLBDaemon, _ := slb.GetStorage().LoadSLBDaemon(slbDaemon.Name)
	if oldSLBDaemon != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "The slbDaemon to create already exists"
		jsonMap["name"] = slbDaemon.Name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(409, string(errorMessageByteSlice))
		return
	}

	err = slb.GetStorage().SaveSLBDaemon(&slbDaemon)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Save slbDaemon failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["slbDaemon"] = slbDaemon
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func putSLBDaemon(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")

	slbDaemon := slb.SLBDaemon{}
	err := request.ReadEntity(&slbDaemon)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Read body failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["name"] = name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	if name != slbDaemon.Name {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Path parameter name is different from name in the body"
		jsonMap["path"] = name
		jsonMap["body"] = slbDaemon.Name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	oldSLBDaemon, _ := slb.GetStorage().LoadSLBDaemon(slbDaemon.Name)
	if oldSLBDaemon == nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "The slbDaemon to update doesn't exist"
		jsonMap["name"] = slbDaemon.Name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	err = slb.GetStorage().SaveSLBDaemon(&slbDaemon)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Save slbDaemon failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["slbDaemon"] = slbDaemon
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func deleteSLBDaemon(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")

	err := slb.GetStorage().DeleteSLBDaemon(name)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Delete slbDaemon failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["name"] = name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func getSLBDaemon(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")

	slbDaemon, err := slb.GetStorage().LoadSLBDaemon(name)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get slbDaemon failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["name"] = name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(slbDaemon, "SLBDaemon")
}

func returns200AllSLBDaemon(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", []slb.SLBDaemon{})
}

func returns200SLBDaemon(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", slb.SLBDaemon{})
}
