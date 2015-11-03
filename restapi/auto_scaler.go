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
	"github.com/cloudawan/cloudone/autoscaler"
	"github.com/cloudawan/cloudone/execute"
	"github.com/cloudawan/cloudone/monitor"
	"github.com/emicklei/go-restful"
	"net/http"
)

func registerWebServiceReplicationControllerAutoScaler() {
	ws := new(restful.WebService)
	ws.Path("/api/v1/autoscalers")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	restful.Add(ws)

	ws.Route(ws.GET("/").To(getAllReplicationControllerAutoScaler).
		Doc("Get all of the configuration of auto scaler").
		Do(returns200AllReplicationControllerAutoScaler, returns500))

	ws.Route(ws.GET("/{namespace}/{kind}/{name}").To(getReplicationControllerAutoScaler).
		Doc("Get the configuration of auto scaler for the replication controller in the namespace").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.PathParameter("kind", "selector or replicationController").DataType("string")).
		Param(ws.PathParameter("name", "name").DataType("string")).
		Do(returns200ReplicationControllerAutoScaler, returns404, returns500))

	ws.Route(ws.PUT("/").To(putReplicationControllerAutoScaler).
		Doc("Add (if not existing) or update an auto scaler for the replication controller in the namespace").
		Do(returns200, returns400, returns404, returns500).
		Reads(autoscaler.ReplicationControllerAutoScaler{}))

	ws.Route(ws.DELETE("/{namespace}/{kind}/{name}").To(deleteReplicationControllerAutoScaler).
		Doc("Delete an auto scaler for the replication controller in the namespace").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.PathParameter("kind", "selector or replicationController").DataType("string")).
		Param(ws.PathParameter("name", "name").DataType("string")).
		Do(returns200, returns500))
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
	if exist {
		response.WriteJson(replicationControllerAutoScaler, "ReplicationControllerAutoScaler")
	} else {
		response.WriteErrorString(404, `{"Error": "No such ReplicationControllerAutoScaler exists"}`)
	}
}

func putReplicationControllerAutoScaler(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespaces")
	replicationcontroller := request.PathParameter("replicationcontrollers")

	replicationControllerAutoScaler := new(autoscaler.ReplicationControllerAutoScaler)
	err := request.ReadEntity(&replicationControllerAutoScaler)

	if err != nil {
		errorText := fmt.Sprintf("PUT namespace %s replicationcontroller %s failure with error %s", namespace, replicationcontroller, err)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	switch replicationControllerAutoScaler.Kind {
	case "selector":
		nameSlice, err := monitor.GetReplicationControllerNameFromSelector(replicationControllerAutoScaler.KubeapiHost, replicationControllerAutoScaler.KubeapiPort, replicationControllerAutoScaler.Namespace, replicationControllerAutoScaler.Name)
		if err != nil {
			for _, name := range nameSlice {
				exist, err := monitor.ExistReplicationController(replicationControllerAutoScaler.KubeapiHost, replicationControllerAutoScaler.KubeapiPort, replicationControllerAutoScaler.Namespace, name)
				if exist == false {
					errorText := fmt.Sprintf("PUT autoscaler %s fail to test the existence of replication controller with error %s", replicationControllerAutoScaler, err)
					log.Error(errorText)
					response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
					return
				}
			}
		}
	case "replicationController":
		exist, err := monitor.ExistReplicationController(replicationControllerAutoScaler.KubeapiHost, replicationControllerAutoScaler.KubeapiPort, replicationControllerAutoScaler.Namespace, replicationControllerAutoScaler.Name)
		if exist == false {
			errorText := fmt.Sprintf("PUT autoscaler %s fail to test the existence of replication controller with error %s", replicationControllerAutoScaler, err)
			log.Error(errorText)
			response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
			return
		}
	default:
		errorText := fmt.Sprintf("PUT autoscaler %s has no such kind %s", replicationControllerAutoScaler, replicationControllerAutoScaler.Kind)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}

	err = autoscaler.SaveReplicationControllerAutoScaler(replicationControllerAutoScaler)
	if err != nil {
		errorText := fmt.Sprintf("PUT autoscaler %s fail to save to database with error %s", replicationControllerAutoScaler, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}

	execute.AddReplicationControllerAutoScaler(replicationControllerAutoScaler)
}

func deleteReplicationControllerAutoScaler(request *restful.Request, response *restful.Response) {
	replicationControllerAutoScaler := new(autoscaler.ReplicationControllerAutoScaler)
	replicationControllerAutoScaler.Namespace = request.PathParameter("namespace")
	replicationControllerAutoScaler.Kind = request.PathParameter("kind")
	replicationControllerAutoScaler.Name = request.PathParameter("name")
	replicationControllerAutoScaler.Check = false
	err := autoscaler.DeleteReplicationControllerAutoScaler(replicationControllerAutoScaler.Namespace, replicationControllerAutoScaler.Kind, replicationControllerAutoScaler.Name)
	if err != nil {
		errorText := fmt.Sprintf("Delete namespace %s replicationcontroller %s fail to delete with error %s", replicationControllerAutoScaler, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
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
