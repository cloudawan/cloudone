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
	"bytes"
	"encoding/json"
	"github.com/cloudawan/cloudone/authorization"
	"github.com/cloudawan/cloudone/utility/configuration"
	"github.com/cloudawan/cloudone_utility/audit"
	"github.com/cloudawan/cloudone_utility/rbac"
	"github.com/cloudawan/cloudone_utility/restclient"
	"github.com/emicklei/go-restful"
	"io/ioutil"
	"net/http"
	"strconv"
)

type UserData struct {
	Username string
	Password string
}

type TokenData struct {
	Token string
}

func registerWebServiceAuthorization() {
	ws := new(restful.WebService)
	ws.Path("/api/v1/authorizations")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	restful.Add(ws)

	// Used for authorization token so don't need to be check authorization
	// No audit since this is used by sytem rather than user
	ws.Route(ws.GET("/tokens/{token}/components/{component}").To(getUserFromToken).
		Doc("Get user data with the token").
		Param(ws.PathParameter("token", "Token").DataType("string")).
		Param(ws.PathParameter("component", "Component").DataType("string")).
		Do(returns200User, returns404, returns500))

	// Used for authorization token so don't need to be check authorization
	ws.Route(ws.POST("/tokens/").Filter(auditLogWithoutVerified).To(postToken).
		Doc("Create the token").
		Do(returns200Token, returns400, returns422, returns500).
		Reads(UserData{}))

	ws.Route(ws.GET("/tokens/expired").Filter(authorize).Filter(auditLog).To(getAllTokenExpiredTime).
		Doc("Get all token's expired time").
		Do(returns200AllTokenExpiredTime, returns500))

	ws.Route(ws.GET("/users/").Filter(authorize).Filter(auditLog).To(getAllUser).
		Doc("Get all of the user").
		Do(returns200AllUser, returns404, returns500))

	ws.Route(ws.POST("/users/").Filter(authorize).Filter(auditLogWithoutBody).To(postUser).
		Doc("Create the user").
		Do(returns200, returns400, returns409, returns422, returns500).
		Reads(rbac.User{}))

	ws.Route(ws.DELETE("/users/{name}").Filter(authorize).Filter(auditLog).To(deleteUser).
		Doc("Delete the user").
		Param(ws.PathParameter("name", "Name").DataType("string")).
		Do(returns200, returns404, returns500))

	ws.Route(ws.PUT("/users/{name}").Filter(authorize).Filter(auditLogWithoutBody).To(putUser).
		Doc("Modify the user").
		Param(ws.PathParameter("name", "Name").DataType("string")).
		Do(returns200, returns400, returns404, returns422, returns500).
		Reads(rbac.User{}))

	ws.Route(ws.GET("/users/{name}").Filter(authorize).Filter(auditLog).To(getUser).
		Doc("Get all of the users").
		Param(ws.PathParameter("name", "Name").DataType("string")).
		Do(returns200User, returns404, returns500))

	ws.Route(ws.PUT("/users/{name}/metadata").Filter(authorize).Filter(auditLogWithoutBody).To(putUserMetaData).
		Doc("Modify the user metadata").
		Param(ws.PathParameter("name", "Name").DataType("string")).
		Do(returns200, returns400, returns404, returns422, returns500).
		Reads(make(map[string]string)))

	ws.Route(ws.GET("/roles/").Filter(authorize).Filter(auditLog).To(getAllRole).
		Doc("Get all of the role").
		Do(returns200AllRole, returns404, returns500))

	ws.Route(ws.POST("/roles/").Filter(authorize).Filter(auditLog).To(postRole).
		Doc("Create the role").
		Do(returns200, returns400, returns409, returns422, returns500).
		Reads(rbac.Role{}))

	ws.Route(ws.DELETE("/roles/{name}").Filter(authorize).Filter(auditLog).To(deleteRole).
		Doc("Delete the role").
		Param(ws.PathParameter("name", "Name").DataType("string")).
		Do(returns200, returns404, returns500))

	ws.Route(ws.PUT("/roles/{name}").Filter(authorize).Filter(auditLog).To(putRole).
		Doc("Modify the role").
		Param(ws.PathParameter("name", "Name").DataType("string")).
		Do(returns200, returns400, returns404, returns422, returns500).
		Reads(rbac.Role{}))

	ws.Route(ws.GET("/roles/{name}").Filter(authorize).Filter(auditLog).To(getRole).
		Doc("Get all of the roles").
		Param(ws.PathParameter("name", "Name").DataType("string")).
		Do(returns200Role, returns404, returns500))
}

func getUserFromToken(request *restful.Request, response *restful.Response) {
	token := request.PathParameter("token")
	component := request.PathParameter("component")

	user, err := authorization.GetUserFromToken(token)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get user with token failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["token"] = token
		jsonMap["component"] = component
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	partialUser := user.CopyPartialUserDataForComponent(component)

	response.WriteJson(partialUser, "User")
}

func postToken(request *restful.Request, response *restful.Response) {
	userData := UserData{}
	err := request.ReadEntity(&userData)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Read body failure"
		jsonMap["ErrorMessage"] = err.Error()
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	token, err := authorization.CreateToken(userData.Username, userData.Password)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Create token failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["userData"] = userData
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(TokenData{token}, "[]TokenOutput")
}

func getAllTokenExpiredTime(request *restful.Request, response *restful.Response) {
	expiredMap := authorization.GetAllTokenExpiredTime()

	response.WriteJson(expiredMap, "{}")
}

func getAllUser(request *restful.Request, response *restful.Response) {
	userSlice, err := authorization.GetStorage().LoadAllUser()
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get all user failure"
		jsonMap["ErrorMessage"] = err.Error()
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(userSlice, "[]User")
}

func postUser(request *restful.Request, response *restful.Response) {
	user := rbac.User{}
	err := request.ReadEntity(&user)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Read body failure"
		jsonMap["ErrorMessage"] = err.Error()
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	oldUser, _ := authorization.GetStorage().LoadUser(user.Name)
	if oldUser != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "The user to create already exists"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["name"] = user.Name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(409, string(errorMessageByteSlice))
		return
	}

	createdUser := rbac.CreateUser(user.Name, user.EncodedPassword, user.RoleSlice, user.ResourceSlice, user.Description, user.MetaDataMap, user.ExpiredTime, user.Disabled)

	err = authorization.GetStorage().SaveUser(createdUser)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Save user failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["user"] = user
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func putUser(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")

	user := rbac.User{}
	err := request.ReadEntity(&user)
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

	if name != user.Name {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Path parameter name is different from name in the body"
		jsonMap["path"] = name
		jsonMap["body"] = user.Name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	oldUser, _ := authorization.GetStorage().LoadUser(name)
	if oldUser == nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "The user to update deosn't exist"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["name"] = name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	createdUser := rbac.CreateUser(user.Name, user.EncodedPassword, user.RoleSlice, user.ResourceSlice, user.Description, user.MetaDataMap, user.ExpiredTime, user.Disabled)

	err = authorization.GetStorage().SaveUser(createdUser)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Save user failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["user"] = user
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func deleteUser(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")

	err := authorization.GetStorage().DeleteUser(name)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Delete user failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["name"] = name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}
}

func getUser(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")

	user, err := authorization.GetStorage().LoadUser(name)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get user failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["name"] = name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(user, "User")
}

func putUserMetaData(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")

	metaDataMap := make(map[string]string)
	err := request.ReadEntity(&metaDataMap)
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

	user, _ := authorization.GetStorage().LoadUser(name)
	if user == nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "The user to update deosn't exist"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["name"] = name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	user.MetaDataMap = metaDataMap

	err = authorization.GetStorage().SaveUser(user)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Save user metadata failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["user"] = user
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func getAllRole(request *restful.Request, response *restful.Response) {
	roleSlice, err := authorization.GetStorage().LoadAllRole()
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get all role failure"
		jsonMap["ErrorMessage"] = err.Error()
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(roleSlice, "[]Role")
}

func postRole(request *restful.Request, response *restful.Response) {
	role := rbac.Role{}
	err := request.ReadEntity(&role)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Read body failure"
		jsonMap["ErrorMessage"] = err.Error()
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	oldRole, _ := authorization.GetStorage().LoadRole(role.Name)
	if oldRole != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "The role to create already exists"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["name"] = role.Name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(409, string(errorMessageByteSlice))
		return
	}

	err = authorization.GetStorage().SaveRole(&role)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Save role failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["role"] = role
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func putRole(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")

	role := rbac.Role{}
	err := request.ReadEntity(&role)
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

	if name != role.Name {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Path parameter name is different from name in the body"
		jsonMap["path"] = name
		jsonMap["body"] = role.Name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	oldRole, _ := authorization.GetStorage().LoadRole(role.Name)
	if oldRole == nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "The role to update deosn't exist"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["name"] = role.Name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	err = authorization.GetStorage().SaveRole(&role)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Save role failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["role"] = role
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func deleteRole(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")

	err := authorization.GetStorage().DeleteRole(name)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Delete role failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["name"] = name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}
}

func getRole(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")

	role, err := authorization.GetStorage().LoadRole(name)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get role failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["name"] = name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(role, "Role")
}

func returns200Token(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", TokenData{})
}

func returns200AllUser(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", []rbac.User{})
}

func returns200AllTokenExpiredTime(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", make(map[interface{}]string))
}

func returns200User(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", rbac.User{})
}

func returns200AllRole(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", []rbac.Role{})
}

func returns200Role(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", rbac.Role{})
}

func auditLogWithoutVerified(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	requestURI := req.Request.URL.RequestURI()
	method := req.Request.Method
	path := req.SelectedRoutePath()
	queryParameterMap := req.Request.URL.Query()
	pathParameterMap := req.PathParameters()
	remoteAddress := req.Request.RemoteAddr

	requestBody, _ := ioutil.ReadAll(req.Request.Body)
	// Write data back for the later use
	req.Request.Body = ioutil.NopCloser(bytes.NewReader(requestBody))

	go func() {
		userData := UserData{}
		req.ReadEntity(&userData)
		userName := userData.Username

		cloudoneAnalysisHost, ok := configuration.LocalConfiguration.GetString("cloudoneAnalysisHost")
		if ok == false {
			log.Error("Fail to get configuration cloudoneAnalysisHost")
			return
		}
		cloudoneAnalysisPort, ok := configuration.LocalConfiguration.GetInt("cloudoneAnalysisPort")
		if ok == false {
			log.Error("Fail to get configuration cloudoneAnalysisPort")
			return
		}

		// Header is not used since the header has no useful information for now
		auditLog := audit.CreateAuditLog(componentName, path, userName, remoteAddress, queryParameterMap, pathParameterMap, method, requestURI, string(requestBody), nil)

		url := "https://" + cloudoneAnalysisHost + ":" + strconv.Itoa(cloudoneAnalysisPort) + "/api/v1/auditlogs"

		headerMap := make(map[string]string)
		headerMap["token"] = authorization.SystemAdminToken

		_, err := restclient.RequestPost(url, auditLog, headerMap, false)
		if err != nil {
			log.Error("Fail to send audit log with error %s", err)
		}
	}()

	chain.ProcessFilter(req, resp)
}
