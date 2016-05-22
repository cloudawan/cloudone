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
	"github.com/cloudawan/cloudone/monitor"
	"github.com/cloudawan/cloudone/utility/configuration"
	"github.com/emicklei/go-restful"
	"net/http"
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
		Do(returns200AllReplicationControllerMetric, returns400, returns404, returns422, returns500))

	ws.Route(ws.GET("/{namespace}/{replicationcontroller}").Filter(authorize).Filter(auditLog).To(getReplicationControllerMetric).
		Doc("Get the replication controller in the namespace").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.PathParameter("replicationcontroller", "Kubernetes replication controller name").DataType("string")).
		Do(returns200ReplicationControllerMetric, returns400, returns404, returns422, returns500))
}

func getAllReplicationControllerMetric(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")

	kubeApiServerEndPoint, kubeApiServerToken, err := configuration.GetAvailablekubeApiServerEndPoint()
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get kube apiserver endpoint and token failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["namespace"] = namespace
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	nameSlice, err := control.GetAllReplicationControllerName(kubeApiServerEndPoint, kubeApiServerToken, namespace)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get all replication controller name failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeApiServerEndPoint"] = kubeApiServerEndPoint
		jsonMap["namespace"] = namespace
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}

	replicationControllerMetricSlice := make([]monitor.ReplicationControllerMetric, 0)
	errorSlice := make([]error, 0)
	for _, name := range nameSlice {
		replicationControllerMetric, err := monitor.MonitorReplicationController(kubeApiServerEndPoint, kubeApiServerToken, namespace, name)
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
	namespace := request.PathParameter("namespace")
	replicationControllerName := request.PathParameter("replicationcontroller")

	kubeApiServerEndPoint, kubeApiServerToken, err := configuration.GetAvailablekubeApiServerEndPoint()
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get kube apiserver endpoint and token failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["namespace"] = namespace
		jsonMap["replicationControllerName"] = replicationControllerName
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	replicationControllerMetric, err := monitor.MonitorReplicationController(kubeApiServerEndPoint, kubeApiServerToken, namespace, replicationControllerName)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get replication controller metric failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeApiServerEndPoint"] = kubeApiServerEndPoint
		jsonMap["namespace"] = namespace
		jsonMap["replicationControllerName"] = replicationControllerName
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(replicationControllerMetric, "ReplicationControllerMetric")
}

func returns200AllReplicationControllerMetric(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", []monitor.ReplicationControllerMetric{})
}

func returns200ReplicationControllerMetric(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", monitor.ReplicationControllerMetric{})
}
