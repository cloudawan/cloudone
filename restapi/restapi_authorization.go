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
	"github.com/cloudawan/cloudone_utility/rbac"
	"github.com/emicklei/go-restful"
)

const (
	componentName = "cloudone"
)

func getCache(token string) *rbac.User {
	// This is special case since cloudone own the authorization server so it doesn't need to ask authorization server and cache but just get data.
	return rbac.GetCache(token)
}

func authorize(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	token := req.Request.Header.Get("token")

	// Get cache. If not exsiting, retrieving from authorization server.
	user := getCache(token)

	// Verify
	if user != nil {
		authorized := false
		if user.HasPermission(componentName, req.Request.Method, req.SelectedRoutePath()) {
			// Resource check
			namespace := req.PathParameter("namespace")
			namespacePass := false
			if namespace != "" {
				if user.HasResource(componentName, "/namespaces/"+namespace) {
					namespacePass = true
				}
			} else {
				namespacePass = true
			}
			if namespacePass {
				authorized = true
			}
		}

		if authorized {
			chain.ProcessFilter(req, resp)
		} else {
			jsonMap := make(map[string]interface{})
			jsonMap["Error"] = "Not Authorized"
			jsonMap["Format"] = "Put correct token in the header token"
			resp.WriteHeaderAndJson(401, jsonMap, "{}")
		}
	} else {
		// Cache doesn't exist
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Token doesn't exist"
		jsonMap["ErrorMessage"] = "Token is incorrect or expired. Please get token with username and password again."
		resp.WriteHeaderAndJson(401, jsonMap, "{}")
	}
}
