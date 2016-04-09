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
	"github.com/cloudawan/cloudone/monitor"
	"github.com/emicklei/go-restful"
	"net/http"
	"strconv"
)

func registerWebServiceNodeMetric() {
	ws := new(restful.WebService)
	ws.Path("/api/v1/nodemetrics")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	restful.Add(ws)

	ws.Route(ws.GET("/").Filter(authorize).Filter(auditLog).To(getAllNodeMetric).
		Doc("Get the node metric").
		Param(ws.QueryParameter("kubeapihost", "Kubernetes host").DataType("string")).
		Param(ws.QueryParameter("kubeapiport", "Kubernetes port").DataType("int")).
		Do(returns200AllNodeMetric, returns400, returns404, returns500))

}

func getAllNodeMetric(request *restful.Request, response *restful.Response) {
	kubeapiHost := request.QueryParameter("kubeapihost")
	kubeapiPortText := request.QueryParameter("kubeapiport")
	if kubeapiHost == "" || kubeapiPortText == "" {
		errorText := fmt.Sprintf("Input text is incorrect kubeapiHost %s kubeapiPort %s", kubeapiHost, kubeapiPortText)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}
	kubeapiPort, err := strconv.Atoi(kubeapiPortText)
	if err != nil {
		errorText := fmt.Sprintf("Could not parse kubeapiPortText %s with error %s", kubeapiPortText, err)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	nodeMetricSlice, err := monitor.MonitorNode(kubeapiHost, kubeapiPort)

	if err != nil {
		errorText := fmt.Sprintf("Fail to get Node metric kubeapiHost %s kubeapiPort %s with error %s", kubeapiHost, kubeapiPortText, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	} else {
		response.WriteJson(nodeMetricSlice, "[]NodeMetric")
	}
}

func returns200AllNodeMetric(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", []monitor.NodeMetric{})
}
