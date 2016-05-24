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
	"github.com/cloudawan/cloudone/registry"
	"github.com/emicklei/go-restful"
	"net/http"
)

func registerWebServicePrivateRegistry() {
	ws := new(restful.WebService)
	ws.Path("/api/v1/privateregistries")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	restful.Add(ws)

	ws.Route(ws.GET("/servers/").Filter(authorize).Filter(auditLog).To(getAllPrivateRegistry).
		Doc("Get all of the private registry server configuration").
		Do(returns200AllPrivateRegistry, returns404, returns500))

	ws.Route(ws.POST("/servers/").Filter(authorize).Filter(auditLog).To(postPrivateRegistry).
		Doc("Create private registry server configuration").
		Do(returns200, returns400, returns409, returns422, returns500).
		Reads(registry.PrivateRegistry{}))

	ws.Route(ws.DELETE("/servers/{server}").Filter(authorize).Filter(auditLog).To(deletePrivateRegistry).
		Doc("Delete the private registry server configuration").
		Param(ws.PathParameter("server", "Server name").DataType("string")).
		Do(returns200, returns422, returns500))

	ws.Route(ws.PUT("/servers/{server}").Filter(authorize).Filter(auditLog).To(putPrivateRegistry).
		Doc("Modify private registry server configuration").
		Param(ws.PathParameter("server", "Server name").DataType("string")).
		Do(returns200, returns400, returns404, returns422, returns500).
		Reads(registry.PrivateRegistry{}))

	ws.Route(ws.GET("/servers/{server}").Filter(authorize).Filter(auditLog).To(getPrivateRegistry).
		Doc("Get the private registry server configuration").
		Param(ws.PathParameter("server", "Server name").DataType("string")).
		Do(returns200PrivateRegistry, returns404, returns500))

	ws.Route(ws.GET("/servers/{server}/repositories/").Filter(authorize).Filter(auditLog).To(getAllPrivateRegistryRepository).
		Doc("Get all of the repository in the private registry server").
		Param(ws.PathParameter("server", "Server name").DataType("string")).
		Do(returns200StringSlice, returns404, returns422, returns500))

	ws.Route(ws.DELETE("/servers/{server}/repositories/{repository}").Filter(authorize).Filter(auditLog).To(deleteAllImageInPrivateRegistryRepository).
		Doc("Delete all the images in the repository in the private registry server").
		Param(ws.PathParameter("server", "Server name").DataType("string")).
		Param(ws.PathParameter("repository", "Repository name").DataType("string")).
		Do(returns200, returns404, returns422, returns500))

	ws.Route(ws.GET("/servers/{server}/repositories/{repository}/tags").Filter(authorize).Filter(auditLog).To(getAllImageTagInPrivateRegistryRepository).
		Doc("Get all the image tags in the repository in the private registry server").
		Param(ws.PathParameter("server", "Server name").DataType("string")).
		Param(ws.PathParameter("repository", "Repository name").DataType("string")).
		Do(returns200StringSlice, returns404, returns422, returns500))

	ws.Route(ws.DELETE("/servers/{server}/repositories/{repository}/tags/{tag}").Filter(authorize).Filter(auditLog).To(deleteImageInPrivateRegistryRepository).
		Doc("Delete the image with the tag in the repository in the private registry server").
		Param(ws.PathParameter("server", "Server name").DataType("string")).
		Param(ws.PathParameter("repository", "Repository name").DataType("string")).
		Param(ws.PathParameter("tag", "Tag name").DataType("string")).
		Do(returns200, returns404, returns422, returns500))
}

func getAllPrivateRegistry(request *restful.Request, response *restful.Response) {
	privateRegistrySlice, err := registry.GetStorage().LoadAllPrivateRegistry()
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get all private registry server configuration failure"
		jsonMap["ErrorMessage"] = err.Error()
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(privateRegistrySlice, "[]PrivateRegistry")
}

func postPrivateRegistry(request *restful.Request, response *restful.Response) {
	privateRegistry := &registry.PrivateRegistry{}
	err := request.ReadEntity(&privateRegistry)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Read body failure"
		jsonMap["ErrorMessage"] = err.Error()
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	oldPrivateRegistry, _ := registry.GetStorage().LoadPrivateRegistry(privateRegistry.Name)
	if oldPrivateRegistry != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "The private registry server configuration to create already exists"
		jsonMap["name"] = privateRegistry.Name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(409, string(errorMessageByteSlice))
		return
	}

	err = registry.GetStorage().SavePrivateRegistry(privateRegistry)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Save private registry server configuration failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["privateRegistry"] = privateRegistry
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func deletePrivateRegistry(request *restful.Request, response *restful.Response) {
	server := request.PathParameter("server")

	err := registry.GetStorage().DeletePrivateRegistry(server)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Delete private registry server configuration failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["server"] = server
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func putPrivateRegistry(request *restful.Request, response *restful.Response) {
	server := request.PathParameter("server")

	privateRegistry := &registry.PrivateRegistry{}
	err := request.ReadEntity(&privateRegistry)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Read body failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["server"] = server
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	if server != privateRegistry.Name {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Path parameter name is different from name in the body"
		jsonMap["path"] = server
		jsonMap["body"] = privateRegistry.Name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	oldPrivateRegistry, _ := registry.GetStorage().LoadPrivateRegistry(server)
	if oldPrivateRegistry == nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "The private registry server configuration to update doesn't exist"
		jsonMap["name"] = server
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	err = registry.GetStorage().SavePrivateRegistry(privateRegistry)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Save private registry server configuration failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["privateRegistry"] = privateRegistry
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func getPrivateRegistry(request *restful.Request, response *restful.Response) {
	server := request.PathParameter("server")

	privateRegistry, err := registry.GetStorage().LoadPrivateRegistry(server)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get private registry server configuration failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["name"] = server
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(privateRegistry, "PrivateRegistry")
}

func getAllPrivateRegistryRepository(request *restful.Request, response *restful.Response) {
	server := request.PathParameter("server")

	privateRegistry, err := registry.GetStorage().LoadPrivateRegistry(server)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get private registry server configuration failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["server"] = server
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	repositorySlice, err := privateRegistry.GetAllRepository()
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get all repository failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["server"] = server
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(repositorySlice, "[]string")
}

func deleteAllImageInPrivateRegistryRepository(request *restful.Request, response *restful.Response) {
	server := request.PathParameter("server")
	repository := request.PathParameter("repository")

	privateRegistry, err := registry.GetStorage().LoadPrivateRegistry(server)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get private registry server configuration failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["server"] = server
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	err = privateRegistry.DeleteAllImageInRepository(repository)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Delete all image in repository failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["server"] = server
		jsonMap["repository"] = repository
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func getAllImageTagInPrivateRegistryRepository(request *restful.Request, response *restful.Response) {
	server := request.PathParameter("server")
	repository := request.PathParameter("repository")

	privateRegistry, err := registry.GetStorage().LoadPrivateRegistry(server)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get private registry server configuration failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["server"] = server
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	tagSlice, err := privateRegistry.GetAllImageTag(repository)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get all tags failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["server"] = server
		jsonMap["repository"] = repository
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(tagSlice, "[]string")
}

func deleteImageInPrivateRegistryRepository(request *restful.Request, response *restful.Response) {
	server := request.PathParameter("server")
	repository := request.PathParameter("repository")
	tag := request.PathParameter("tag")

	privateRegistry, err := registry.GetStorage().LoadPrivateRegistry(server)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get private registry server configuration failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["server"] = server
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	err = privateRegistry.DeleteImageInRepository(repository, tag)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Delete image in repository failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["server"] = server
		jsonMap["repository"] = repository
		jsonMap["tag"] = tag
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}
}

func returns200AllPrivateRegistry(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", []registry.PrivateRegistry{})
}

func returns200PrivateRegistry(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", registry.PrivateRegistry{})
}

func returns200StringSlice(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", make([]string, 0))
}
