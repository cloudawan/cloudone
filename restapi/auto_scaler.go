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
	"github.com/cloudawan/cloudone/autoscaler"
	"github.com/cloudawan/cloudone/deploy"
	"github.com/cloudawan/cloudone/execute"
	"github.com/cloudawan/cloudone/monitor"
	"github.com/cloudawan/cloudone/utility/configuration"
	"github.com/emicklei/go-restful"
	"net/http"
)

func registerWebServiceReplicationControllerAutoScaler() {
	ws := new(restful.WebService)
	ws.Path("/api/v1/autoscalers")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	restful.Add(ws)

	ws.Route(ws.GET("/").Filter(authorize).Filter(auditLog).To(getAllReplicationControllerAutoScaler).
		Doc("Get all of the configuration of auto scaler").
		Do(returns200AllReplicationControllerAutoScaler, returns500))

	ws.Route(ws.GET("/{namespace}/{kind}/{name}").Filter(authorize).Filter(auditLog).To(getReplicationControllerAutoScaler).
		Doc("Get the configuration of auto scaler for the replication controller in the namespace").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.PathParameter("kind", "selector or replicationController").DataType("string")).
		Param(ws.PathParameter("name", "name").DataType("string")).
		Do(returns200ReplicationControllerAutoScaler, returns404, returns500))

	ws.Route(ws.PUT("/").Filter(authorize).Filter(auditLog).To(putReplicationControllerAutoScaler).
		Doc("Add (if not existing) or update an auto scaler for the replication controller in the namespace").
		Do(returns200, returns400, returns404, returns422, returns500).
		Reads(autoscaler.ReplicationControllerAutoScaler{}))

	ws.Route(ws.DELETE("/{namespace}/{kind}/{name}").Filter(authorize).Filter(auditLog).To(deleteReplicationControllerAutoScaler).
		Doc("Delete an auto scaler for the replication controller in the namespace").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.PathParameter("kind", "selector or replicationController").DataType("string")).
		Param(ws.PathParameter("name", "name").DataType("string")).
		Do(returns200, returns422, returns500))
}

func getAllReplicationControllerAutoScaler(request *restful.Request, response *restful.Response) {
	replicationControllerAutoScalerMap := execute.GetReplicationControllerAutoScalerMap()
	replicationControllerAutoScalerSlice := make([]autoscaler.ReplicationControllerAutoScaler, 0)
	for _, replicationControllerAutoScaler := range replicationControllerAutoScalerMap {
		replicationControllerAutoScalerSlice = append(replicationControllerAutoScalerSlice, replicationControllerAutoScaler)
	}

	response.WriteJson(replicationControllerAutoScalerSlice, "[]ReplicationControllerAutoScaler")
}

func getReplicationControllerAutoScaler(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	kind := request.PathParameter("kind")
	name := request.PathParameter("name")

	exist, replicationControllerAutoScaler := execute.GetReplicationControllerAutoScaler(namespace, kind, name)
	if exist == false {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "The replication controller autoscaler doesn't exist"
		jsonMap["namespace"] = namespace
		jsonMap["kind"] = kind
		jsonMap["name"] = name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(replicationControllerAutoScaler, "ReplicationControllerAutoScaler")
}

func putReplicationControllerAutoScaler(request *restful.Request, response *restful.Response) {
	replicationControllerAutoScaler := new(autoscaler.ReplicationControllerAutoScaler)
	err := request.ReadEntity(&replicationControllerAutoScaler)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Read body failure"
		jsonMap["ErrorMessage"] = err.Error()
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

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

	replicationControllerAutoScaler.KubeApiServerEndPoint = kubeApiServerEndPoint
	replicationControllerAutoScaler.KubeApiServerToken = kubeApiServerToken

	switch replicationControllerAutoScaler.Kind {
	case "application":
		_, err := deploy.GetStorage().LoadDeployInformation(replicationControllerAutoScaler.Namespace, replicationControllerAutoScaler.Name)
		if err != nil {
			jsonMap := make(map[string]interface{})
			jsonMap["Error"] = "Check whether the application exists or not failure"
			jsonMap["ErrorMessage"] = err.Error()
			jsonMap["replicationControllerAutoScaler"] = replicationControllerAutoScaler
			errorMessageByteSlice, _ := json.Marshal(jsonMap)
			log.Error(jsonMap)
			response.WriteErrorString(422, string(errorMessageByteSlice))
			return
		}
	case "selector":
		nameSlice, err := monitor.GetReplicationControllerNameFromSelector(replicationControllerAutoScaler.KubeApiServerEndPoint, replicationControllerAutoScaler.KubeApiServerToken, replicationControllerAutoScaler.Namespace, replicationControllerAutoScaler.Name)
		if err != nil {
			for _, name := range nameSlice {
				exist, err := monitor.ExistReplicationController(replicationControllerAutoScaler.KubeApiServerEndPoint, replicationControllerAutoScaler.KubeApiServerToken, replicationControllerAutoScaler.Namespace, name)
				if err != nil {
					jsonMap := make(map[string]interface{})
					jsonMap["Error"] = "Check whether the replication controller exists or not failure"
					jsonMap["ErrorMessage"] = err.Error()
					jsonMap["replicationControllerAutoScaler"] = replicationControllerAutoScaler
					errorMessageByteSlice, _ := json.Marshal(jsonMap)
					log.Error(jsonMap)
					response.WriteErrorString(422, string(errorMessageByteSlice))
					return
				}
				if exist == false {
					jsonMap := make(map[string]interface{})
					jsonMap["Error"] = "The replication controller to auto scale doesn't exist"
					jsonMap["ErrorMessage"] = err.Error()
					jsonMap["replicationControllerAutoScaler"] = replicationControllerAutoScaler
					errorMessageByteSlice, _ := json.Marshal(jsonMap)
					log.Error(jsonMap)
					response.WriteErrorString(404, string(errorMessageByteSlice))
					return
				}
			}
		}
	case "replicationController":
		exist, err := monitor.ExistReplicationController(replicationControllerAutoScaler.KubeApiServerEndPoint, replicationControllerAutoScaler.KubeApiServerToken, replicationControllerAutoScaler.Namespace, replicationControllerAutoScaler.Name)
		if err != nil {
			jsonMap := make(map[string]interface{})
			jsonMap["Error"] = "Check whether the replication controller exists or not failure"
			jsonMap["ErrorMessage"] = err.Error()
			jsonMap["replicationControllerAutoScaler"] = replicationControllerAutoScaler
			errorMessageByteSlice, _ := json.Marshal(jsonMap)
			log.Error(jsonMap)
			response.WriteErrorString(422, string(errorMessageByteSlice))
			return
		}
		if exist == false {
			jsonMap := make(map[string]interface{})
			jsonMap["Error"] = "The replication controller to auto scale doesn't exist"
			jsonMap["ErrorMessage"] = err.Error()
			jsonMap["replicationControllerAutoScaler"] = replicationControllerAutoScaler
			errorMessageByteSlice, _ := json.Marshal(jsonMap)
			log.Error(jsonMap)
			response.WriteErrorString(404, string(errorMessageByteSlice))
			return
		}
	default:
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "No such kind"
		jsonMap["replicationControllerAutoScaler"] = replicationControllerAutoScaler
		jsonMap["kind"] = replicationControllerAutoScaler.Kind
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	err = autoscaler.GetStorage().SaveReplicationControllerAutoScaler(replicationControllerAutoScaler)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Save replication controller autoscaler failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["replicationControllerAutoScaler"] = replicationControllerAutoScaler
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}

	execute.AddReplicationControllerAutoScaler(replicationControllerAutoScaler)
}

func deleteReplicationControllerAutoScaler(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	kind := request.PathParameter("kind")
	name := request.PathParameter("name")

	replicationControllerAutoScaler := &autoscaler.ReplicationControllerAutoScaler{}
	replicationControllerAutoScaler.Namespace = namespace
	replicationControllerAutoScaler.Kind = kind
	replicationControllerAutoScaler.Name = name
	replicationControllerAutoScaler.Check = false

	err := autoscaler.GetStorage().DeleteReplicationControllerAutoScaler(replicationControllerAutoScaler.Namespace, replicationControllerAutoScaler.Kind, replicationControllerAutoScaler.Name)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Delete replication controller autoscaler failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["namespace"] = namespace
		jsonMap["kind"] = kind
		jsonMap["name"] = name
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}

	execute.AddReplicationControllerAutoScaler(replicationControllerAutoScaler)
}

func returns200AllReplicationControllerAutoScaler(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", []autoscaler.ReplicationControllerAutoScaler{})
}

func returns200ReplicationControllerAutoScaler(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", autoscaler.ReplicationControllerAutoScaler{})
}
