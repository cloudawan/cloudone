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

	ws.Route(ws.GET("/credentials/").Filter(authorize).To(getAllCredential).
		Doc("Get all of the credential").
		Do(returns200AllCredential, returns404, returns500))

	ws.Route(ws.POST("/credentials/").Filter(authorize).To(postCredential).
		Doc("Create the credential").
		Do(returns200, returns400, returns404, returns500).
		Reads(host.Credential{}))

	ws.Route(ws.DELETE("/credentials/{ip}").Filter(authorize).To(deleteCredential).
		Doc("Delete the credential").
		Param(ws.PathParameter("ip", "IP").DataType("string")).
		Do(returns200, returns404, returns500))

	ws.Route(ws.PUT("/credentials/{ip}").Filter(authorize).To(putCredential).
		Doc("Modify the credential").
		Param(ws.PathParameter("ip", "IP").DataType("string")).
		Do(returns200, returns400, returns404, returns500).
		Reads(host.Credential{}))

	ws.Route(ws.GET("/credentials/{ip}").Filter(authorize).To(getCredential).
		Doc("Get all of the credentials").
		Param(ws.PathParameter("ip", "IP").DataType("string")).
		Do(returns200Credential, returns404, returns500))
}

func getAllCredential(request *restful.Request, response *restful.Response) {
	credentialClusterSlice, err := host.GetStorage().LoadAllCredential()
	if err != nil {
		errorText := fmt.Sprintf("Could not get all credentials with error %s", err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}

	response.WriteJson(credentialClusterSlice, "[]Credential")
}

func postCredential(request *restful.Request, response *restful.Response) {
	credential := host.Credential{}
	err := request.ReadEntity(&credential)

	if err != nil {
		errorText := fmt.Sprintf("POST parse credential input failure with error %s", err)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	oldCredential, _ := host.GetStorage().LoadCredential(credential.IP)
	if oldCredential != nil {
		errorText := fmt.Sprintf("The credential with ip %s exists", credential.IP)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	err = host.GetStorage().SaveCredential(&credential)
	if err != nil {
		errorText := fmt.Sprintf("Save credential %v with error %s", credential, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
}

func putCredential(request *restful.Request, response *restful.Response) {
	ip := request.PathParameter("ip")

	credential := host.Credential{}
	err := request.ReadEntity(&credential)

	if err != nil {
		errorText := fmt.Sprintf("PUT parse credential input failure with error %s", err)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	if ip != credential.IP {
		errorText := fmt.Sprintf("PUT ip %s is different from ip %s in the body", ip, credential.IP)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	oldCredential, _ := host.GetStorage().LoadCredential(credential.IP)
	if oldCredential == nil {
		errorText := fmt.Sprintf("The credential with ip %s doesn't exist", credential.IP)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	err = host.GetStorage().SaveCredential(&credential)
	if err != nil {
		errorText := fmt.Sprintf("Save credential %v with error %s", credential, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
}

func deleteCredential(request *restful.Request, response *restful.Response) {
	ip := request.PathParameter("ip")

	err := host.GetStorage().DeleteCredential(ip)
	if err != nil {
		errorText := fmt.Sprintf("Delete credential with ip %s error %s", ip, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
}

func getCredential(request *restful.Request, response *restful.Response) {
	ip := request.PathParameter("ip")

	credential, err := host.GetStorage().LoadCredential(ip)
	if err != nil {
		errorText := fmt.Sprintf("Could not get credential %s with error %s", credential, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
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
