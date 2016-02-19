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
	"github.com/cloudawan/cloudone/control"
	"github.com/emicklei/go-restful"
	"net/http"
	"strconv"
)

func registerWebServicePodLog() {
	ws := new(restful.WebService)
	ws.Path("/api/v1/podlogs")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	restful.Add(ws)

	ws.Route(ws.GET("/{namespace}/{pod}").To(getPodLog).
		Doc("Get log for pod").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.PathParameter("pod", "Kubernetes pod").DataType("string")).
		Param(ws.QueryParameter("kubeapihost", "Kubernetes host").DataType("string")).
		Param(ws.QueryParameter("kubeapiport", "Kubernetes port").DataType("int")).
		Do(returns200PodLog, returns400, returns404, returns500))
}

func getPodLog(request *restful.Request, response *restful.Response) {
	kubeapiHost := request.QueryParameter("kubeapihost")
	kubeapiPortText := request.QueryParameter("kubeapiport")
	namespace := request.PathParameter("namespace")
	pod := request.PathParameter("pod")
	if kubeapiHost == "" || kubeapiPortText == "" || namespace == "" || pod == "" {
		errorText := fmt.Sprintf("Input text is incorrect kubeapiHost %s kubeapiPort %s namespace %s pod %s", kubeapiHost, kubeapiPortText, namespace, pod)
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

	logJsonMap, err := control.GetPodLog(kubeapiHost, kubeapiPort, namespace, pod)

	if err != nil {
		errorText := fmt.Sprintf("Get pod log failure kubeapiHost %s kubeapiPort %s namespace %s pod %s", kubeapiHost, kubeapiPortText, namespace, pod)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}

	response.WriteJson(logJsonMap, "{}")
}

func returns200PodLog(b *restful.RouteBuilder) {
	jsonMap := make(map[string]interface{})
	b.Returns(http.StatusOK, "OK", jsonMap)
}
