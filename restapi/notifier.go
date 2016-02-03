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
	"github.com/cloudawan/cloudone/execute"
	"github.com/cloudawan/cloudone/monitor"
	"github.com/cloudawan/cloudone/notification"
	"github.com/emicklei/go-restful"
	"net/http"
)

func registerWebServiceReplicationControllerNotifier() {
	ws := new(restful.WebService)
	ws.Path("/api/v1/notifiers")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	restful.Add(ws)

	ws.Route(ws.GET("/").To(getAllReplicationControllerNotifier).
		Doc("Get all of the configuration of notifier").
		Do(returns200AllReplicationControllerNotifier, returns500))

	ws.Route(ws.GET("/{namespace}/{kind}/{name}").To(getReplicationControllerNotifier).
		Doc("Get the configuration of notifier for the replication controller in the namespace").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.PathParameter("kind", "selector or replicationController").DataType("string")).
		Param(ws.PathParameter("name", "name").DataType("string")).
		Do(returns200ReplicationControllerNotifier, returns404, returns500))

	ws.Route(ws.PUT("/").To(putReplicationControllerNotifier).
		Doc("Add (if not existing) or update an notifier for the replication controller in the namespace").
		Do(returns200, returns400, returns404, returns500).
		Reads(notification.ReplicationControllerNotifierSerializable{}))

	ws.Route(ws.DELETE("/{namespace}/{kind}/{name}").To(deleteReplicationControllerNotifier).
		Doc("Delete an notifier for the replication controller in the namespace").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.PathParameter("kind", "selector or replicationController").DataType("string")).
		Param(ws.PathParameter("name", "name").DataType("string")).
		Do(returns200, returns500))
}

func getAllReplicationControllerNotifier(request *restful.Request, response *restful.Response) {
	replicationControllerNotifierMap := execute.GetReplicationControllerNotifierMap()
	replicationControllerNotifierSerializableSlice := make([]notification.ReplicationControllerNotifierSerializable, 0)
	for _, replicationControllerNotifier := range replicationControllerNotifierMap {
		replicationControllerNotifierSerializable, err := notification.ConvertToSerializable(replicationControllerNotifier)
		if err != nil {
			log.Error(err)
		} else {
			replicationControllerNotifierSerializableSlice = append(replicationControllerNotifierSerializableSlice, replicationControllerNotifierSerializable)
		}
	}
	response.WriteJson(replicationControllerNotifierSerializableSlice, "[]ReplicationControllerNotifierSerializable")
}

func getReplicationControllerNotifier(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	kind := request.PathParameter("kind")
	name := request.PathParameter("name")
	exist, replicationControllerNotifier := execute.GetReplicationControllerNotifier(namespace, kind, name)
	if exist {
		replicationControllerNotifierSerializable, err := notification.ConvertToSerializable(replicationControllerNotifier)
		if err != nil {
			log.Error(err)
		} else {
			response.WriteJson(replicationControllerNotifierSerializable, "ReplicationControllerNotifierSerializable")
		}
	} else {
		response.WriteErrorString(404, `{"Error": "No such ReplicationControllerNotifier exists"}`)
	}
}

func putReplicationControllerNotifier(request *restful.Request, response *restful.Response) {

	replicationControllerNotifierSerializable := new(notification.ReplicationControllerNotifierSerializable)
	err := request.ReadEntity(&replicationControllerNotifierSerializable)

	if err != nil {
		errorText := fmt.Sprintf("PUT notifier %s failure with error %s", replicationControllerNotifierSerializable, err)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	switch replicationControllerNotifierSerializable.Kind {
	case "selector":
		nameSlice, err := monitor.GetReplicationControllerNameFromSelector(replicationControllerNotifierSerializable.KubeapiHost, replicationControllerNotifierSerializable.KubeapiPort, replicationControllerNotifierSerializable.Namespace, replicationControllerNotifierSerializable.Name)
		if err != nil {
			for _, name := range nameSlice {
				exist, err := monitor.ExistReplicationController(replicationControllerNotifierSerializable.KubeapiHost, replicationControllerNotifierSerializable.KubeapiPort, replicationControllerNotifierSerializable.Namespace, name)
				if exist == false {
					errorText := fmt.Sprintf("PUT notifier %s fail to test the existence of replication controller with error %s", replicationControllerNotifierSerializable, err)
					log.Error(errorText)
					response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
					return
				}
			}
		}
	case "replicationController":
		exist, err := monitor.ExistReplicationController(replicationControllerNotifierSerializable.KubeapiHost, replicationControllerNotifierSerializable.KubeapiPort, replicationControllerNotifierSerializable.Namespace, replicationControllerNotifierSerializable.Name)
		if exist == false {
			errorText := fmt.Sprintf("PUT notifier %s fail to test the existence of replication controller with error %s", replicationControllerNotifierSerializable, err)
			log.Error(errorText)
			response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
			return
		}
	default:
		errorText := fmt.Sprintf("PUT notifier %s has no such kind %s", replicationControllerNotifierSerializable, replicationControllerNotifierSerializable.Kind)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}

	replicationControllerNotifier, err := notification.ConvertFromSerializable(*replicationControllerNotifierSerializable)
	if err != nil {
		errorText := fmt.Sprintf("PUT notifier %s fail to convert with error %s", replicationControllerNotifierSerializable, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}

	err = notification.GetStorage().SaveReplicationControllerNotifierSerializable(replicationControllerNotifierSerializable)
	if err != nil {
		errorText := fmt.Sprintf("PUT notifier %s fail to save to database with error %s", replicationControllerNotifierSerializable, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}

	execute.AddReplicationControllerNotifier(&replicationControllerNotifier)
}

func deleteReplicationControllerNotifier(request *restful.Request, response *restful.Response) {
	replicationControllerNotifier := new(notification.ReplicationControllerNotifier)
	replicationControllerNotifier.Namespace = request.PathParameter("namespace")
	replicationControllerNotifier.Kind = request.PathParameter("kind")
	replicationControllerNotifier.Name = request.PathParameter("name")
	replicationControllerNotifier.Check = false
	err := notification.GetStorage().DeleteReplicationControllerNotifierSerializable(replicationControllerNotifier.Namespace, replicationControllerNotifier.Kind, replicationControllerNotifier.Name)
	if err != nil {
		errorText := fmt.Sprintf("Delete namespace %s kind %s name %s fail to delete with error %s", replicationControllerNotifier.Namespace, replicationControllerNotifier.Kind, replicationControllerNotifier.Name, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
	execute.AddReplicationControllerNotifier(replicationControllerNotifier)
}

func returns200AllReplicationControllerNotifier(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", []notification.ReplicationControllerNotifierSerializable{})
}

func returns200ReplicationControllerNotifier(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", notification.ReplicationControllerNotifierSerializable{})
}
