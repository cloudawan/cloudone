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
	"github.com/cloudawan/cloudone/control"
	"github.com/cloudawan/cloudone/utility/configuration"
	"github.com/emicklei/go-restful"
	"net/http"
)

func registerWebServiceNode() {
	ws := new(restful.WebService)
	ws.Path("/api/v1/nodes")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	restful.Add(ws)

	ws.Route(ws.GET("/topology").Filter(authorize).Filter(auditLog).To(getNodeTopology).
		Doc("Get the node topology").
		Do(returns200NodeTopology, returns400, returns404, returns422, returns500))

}

func getNodeTopology(request *restful.Request, response *restful.Response) {
	kubeApiServerEndPoint, kubeApiServerToken, err := configuration.GetAvailablekubeApiServerEndPoint()
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get kube apiserver endpoint and token failure"
		jsonMap["ErrorMessage"] = err.Error()
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	regionSlice, err := control.GetNodeTopology(kubeApiServerEndPoint, kubeApiServerToken)

	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get node topology failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeApiServerEndPoint"] = kubeApiServerEndPoint
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(regionSlice, "[]Region")
}

func returns200NodeTopology(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", []control.Region{})
}
