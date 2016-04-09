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
	"github.com/cloudawan/cloudone/healthcheck"
	"github.com/emicklei/go-restful"
	"net/http"
)

func registerWebServiceHealthCheck() {
	ws := new(restful.WebService)
	ws.Path("/api/v1/healthchecks")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	restful.Add(ws)

	ws.Route(ws.GET("/").Filter(authorize).Filter(auditLog).To(getAllStatus).
		Doc("Get all status").
		Do(returns200Map, returns400, returns404, returns500))
}

func getAllStatus(request *restful.Request, response *restful.Response) {
	jsonMap, err := healthcheck.GetAllStatus()
	if err != nil {
		errorText := fmt.Sprintf("Fail to get the all status with error %s", err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}

	response.WriteJson(jsonMap, "{}")
}

func returns200Map(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", make(map[string]interface{}))
}
