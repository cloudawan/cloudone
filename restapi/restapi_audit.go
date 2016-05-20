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
	"github.com/cloudawan/cloudone/authorization"
	"github.com/cloudawan/cloudone/utility/configuration"
	"github.com/cloudawan/cloudone_utility/audit"
	"github.com/cloudawan/cloudone_utility/restclient"
	"github.com/emicklei/go-restful"
	"io/ioutil"
	"strconv"
)

func auditLog(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	token := req.Request.Header.Get("token")
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
		sendAuditLog(token, requestURI, method, path, string(requestBody), queryParameterMap, pathParameterMap, remoteAddress)
	}()

	chain.ProcessFilter(req, resp)
}

func auditLogWithoutBody(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	token := req.Request.Header.Get("token")
	requestURI := req.Request.URL.RequestURI()
	method := req.Request.Method
	path := req.SelectedRoutePath()
	queryParameterMap := req.Request.URL.Query()
	pathParameterMap := req.PathParameters()
	remoteAddress := req.Request.RemoteAddr

	// Don't record body for password related operation
	requestBody := ""

	go func() {
		sendAuditLog(token, requestURI, method, path, requestBody, queryParameterMap, pathParameterMap, remoteAddress)
	}()

	chain.ProcessFilter(req, resp)
}

func sendAuditLog(token string, requestURI string, method string, path string, requestBody string, queryParameterMap map[string][]string, pathParameterMap map[string]string, remoteAddress string) {
	// Get cache. If not exsiting, retrieving from authorization server.
	user := getCache(token)
	userName := "no_user"
	if user != nil {
		userName = user.Name
	}

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
	auditLog := audit.CreateAuditLog(componentName, path, userName, remoteAddress, queryParameterMap, pathParameterMap, method, requestURI, requestBody, nil)

	url := "https://" + cloudoneAnalysisHost + ":" + strconv.Itoa(cloudoneAnalysisPort) + "/api/v1/auditlogs"

	headerMap := make(map[string]string)
	headerMap["token"] = authorization.SystemAdminToken

	_, err := restclient.RequestPost(url, auditLog, headerMap, false)
	if err != nil {
		log.Error("Fail to send audit log with error %s", err)
	}
}
