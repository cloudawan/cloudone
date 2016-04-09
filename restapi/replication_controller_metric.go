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
	"github.com/cloudawan/cloudone/monitor"
	"github.com/emicklei/go-restful"
	"net/http"
	"strconv"
)

func registerWebServiceReplicationControllerMetric() {
	ws := new(restful.WebService)
	ws.Path("/api/v1/replicationcontrollermetrics")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	restful.Add(ws)

	ws.Route(ws.GET("/{namespace}").Filter(authorize).Filter(auditLog).To(getAllReplicationControllerMetric).
		Doc("Get all replication controllers in the namespace").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.QueryParameter("kubeapihost", "Kubernetes host").DataType("string")).
		Param(ws.QueryParameter("kubeapiport", "Kubernetes port").DataType("int")).
		Do(returns200AllReplicationControllerMetric, returns400, returns404, returns500))

	ws.Route(ws.GET("/{namespace}/{replicationcontroller}").Filter(authorize).Filter(auditLog).To(getReplicationControllerMetric).
		Doc("Get the replication controller in the namespace").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.PathParameter("replicationcontroller", "Kubernetes replication controller name").DataType("string")).
		Param(ws.QueryParameter("kubeapihost", "Kubernetes host").DataType("string")).
		Param(ws.QueryParameter("kubeapiport", "Kubernetes port").DataType("int")).
		Do(returns200ReplicationControllerMetric, returns400, returns404, returns500))
}

func getAllReplicationControllerMetric(request *restful.Request, response *restful.Response) {
	kubeapiHost := request.QueryParameter("kubeapihost")
	kubeapiPortText := request.QueryParameter("kubeapiport")
	namespace := request.PathParameter("namespace")
	if kubeapiHost == "" || kubeapiPortText == "" || namespace == "" {
		errorText := fmt.Sprintf("Input text is incorrect kubeapiHost %s kubeapiPort %s namespace %s", kubeapiHost, kubeapiPortText, namespace)
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

	nameSlice, err := control.GetAllReplicationControllerName(kubeapiHost, kubeapiPort, namespace)
	if err != nil {
		errorText := fmt.Sprintf("Get all replication controller name failure kubeapiHost %s kubeapiHost %s namespace %s", kubeapiHost, kubeapiPortText, namespace)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}

	replicationControllerMetricSlice := make([]monitor.ReplicationControllerMetric, 0)
	errorSlice := make([]error, 0)
	for _, name := range nameSlice {
		replicationControllerMetric, err := monitor.MonitorReplicationController(kubeapiHost, kubeapiPort, namespace, name)
		if replicationControllerMetric != nil {
			replicationControllerMetricSlice = append(replicationControllerMetricSlice, *replicationControllerMetric)
		}
		errorSlice = append(errorSlice, err)
	}

	returnedJsonMap := make(map[string]interface{})
	returnedJsonMap["ReplicationControllerMetricSlice"] = replicationControllerMetricSlice
	returnedJsonMap["ErrorSlice"] = errorSlice
	response.WriteJson(returnedJsonMap, "ReplicationControllerMetricSlice/ErrorSlice")
}

func getReplicationControllerMetric(request *restful.Request, response *restful.Response) {
	kubeapiHost := request.QueryParameter("kubeapihost")
	kubeapiPortText := request.QueryParameter("kubeapiport")
	namespace := request.PathParameter("namespace")
	replicationControllerName := request.PathParameter("replicationcontroller")
	if kubeapiPortText == "" || kubeapiPortText == "" || namespace == "" || replicationControllerName == "" {
		errorText := fmt.Sprintf("Input text is incorrect kubeapiHost %s kubeapiPort %s namespace %s replicationControllerName %s", kubeapiHost, kubeapiPortText, namespace, replicationControllerName)
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
	replicationControllerMetric, err := monitor.MonitorReplicationController(kubeapiHost, kubeapiPort, namespace, replicationControllerName)
	if err != nil {
		errorText := fmt.Sprintf("Fail to get ReplicationController kubeapiHost %s kubeapiPort %s namespace %s replicationControllerName %s with error %s", kubeapiHost, kubeapiPortText, namespace, replicationControllerName, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	} else {
		response.WriteJson(replicationControllerMetric, "ReplicationControllerMetric")
	}
}

func returns200AllReplicationControllerMetric(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", []monitor.ReplicationControllerMetric{})
}

func returns200ReplicationControllerMetric(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", monitor.ReplicationControllerMetric{})
}
