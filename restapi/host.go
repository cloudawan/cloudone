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
	"github.com/cloudawan/cloudone/host"
	"github.com/emicklei/go-restful"
	"net/http"
)

func registerWebServiceHost() {
	ws := new(restful.WebService)
	ws.Path("/api/v1/hosts")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	restful.Add(ws)

	ws.Route(ws.GET("/credentials/").Filter(authorize).Filter(auditLog).To(getAllCredential).
		Doc("Get all of the credential").
		Do(returns200AllCredential, returns422, returns500))

	ws.Route(ws.POST("/credentials/").Filter(authorize).Filter(auditLogWithoutBody).To(postCredential).
		Doc("Create the credential").
		Do(returns200, returns400, returns409, returns422, returns500).
		Reads(host.Credential{}))

	ws.Route(ws.DELETE("/credentials/{ip}").Filter(authorize).Filter(auditLog).To(deleteCredential).
		Doc("Delete the credential").
		Param(ws.PathParameter("ip", "IP").DataType("string")).
		Do(returns200, returns422, returns500))

	ws.Route(ws.PUT("/credentials/{ip}").Filter(authorize).Filter(auditLogWithoutBody).To(putCredential).
		Doc("Modify the credential").
		Param(ws.PathParameter("ip", "IP").DataType("string")).
		Do(returns200, returns400, returns404, returns422, returns500).
		Reads(host.Credential{}))

	ws.Route(ws.GET("/credentials/{ip}").Filter(authorize).Filter(auditLog).To(getCredential).
		Doc("Get all of the credentials").
		Param(ws.PathParameter("ip", "IP").DataType("string")).
		Do(returns200Credential, returns422, returns500))
}

func getAllCredential(request *restful.Request, response *restful.Response) {
	credentialClusterSlice, err := host.GetStorage().LoadAllCredential()
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get all credential failure"
		jsonMap["ErrorMessage"] = err.Error()
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(credentialClusterSlice, "[]Credential")
}

func postCredential(request *restful.Request, response *restful.Response) {
	credential := host.Credential{}
	err := request.ReadEntity(&credential)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Read body failure"
		jsonMap["ErrorMessage"] = err.Error()
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	oldCredential, _ := host.GetStorage().LoadCredential(credential.IP)
	if oldCredential != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "The credential to create already exists"
		jsonMap["ip"] = credential.IP
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(409, string(errorMessageByteSlice))
		return
	}

	err = host.GetStorage().SaveCredential(&credential)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Save credential failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["credential"] = credential
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func putCredential(request *restful.Request, response *restful.Response) {
	ip := request.PathParameter("ip")

	credential := host.Credential{}
	err := request.ReadEntity(&credential)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Read body failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["ip"] = ip
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	if ip != credential.IP {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Path parameter ip is different from ip in the body"
		jsonMap["path"] = ip
		jsonMap["body"] = credential.IP
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	oldCredential, _ := host.GetStorage().LoadCredential(credential.IP)
	if oldCredential == nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "The credential to update doesn't exist"
		jsonMap["ip"] = credential.IP
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	err = host.GetStorage().SaveCredential(&credential)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Save credential failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["credential"] = credential
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func deleteCredential(request *restful.Request, response *restful.Response) {
	ip := request.PathParameter("ip")

	err := host.GetStorage().DeleteCredential(ip)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Delete credential failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["ip"] = ip
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func getCredential(request *restful.Request, response *restful.Response) {
	ip := request.PathParameter("ip")

	credential, err := host.GetStorage().LoadCredential(ip)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get credential failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["ip"] = ip
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(credential, "Credential")
}

func returns200AllCredential(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", []host.Credential{})
}

func returns200Credential(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", host.Credential{})
}
