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
	"bytes"
	"encoding/json"
	"github.com/cloudawan/cloudone/control"
	"github.com/cloudawan/cloudone/utility/configuration"
	"github.com/emicklei/go-restful"
	"net/http"
)

type SizeInput struct {
	Size int
}

func registerWebServiceKubernetesService() {
	ws := new(restful.WebService)
	ws.Path("/api/v1/replicationcontrollers")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	restful.Add(ws)

	ws.Route(ws.GET("/{namespace}").Filter(authorize).Filter(auditLog).To(getAllReplicationController).
		Doc("Get all replication controllers in the namespace").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.QueryParameter("kubeapihost", "Kubernetes host").DataType("string")).
		Param(ws.QueryParameter("kubeapiport", "Kubernetes port").DataType("int")).
		Do(returns200AllReplicationController, returns400, returns422, returns500))

	ws.Route(ws.POST("/{namespace}").Filter(authorize).Filter(auditLog).To(postReplicationController).
		Doc("Add a replication controller in the namespace").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Do(returns200, returns400, returns404, returns422, returns500).
		Reads(control.ReplicationController{}))

	ws.Route(ws.GET("/{namespace}/{replicationcontroller}").Filter(authorize).Filter(auditLog).To(getReplicationController).
		Doc("Get replication controllers in the namespace").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.PathParameter("replicationcontroller", "Kubernetes replication controller name").DataType("string")).
		Do(returns200ReplicationController, returns400, returns404, returns422, returns500))

	ws.Route(ws.DELETE("/{namespace}/{replicationcontroller}").Filter(authorize).Filter(auditLog).To(deleteReplicationController).
		Doc("Delete the replication controller in the namespace").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.PathParameter("replicationcontroller", "Kubernetes replication controller name").DataType("string")).
		Do(returns200, returns400, returns404, returns422, returns500))

	ws.Route(ws.PUT("/size/{namespace}/{replicationcontroller}").Filter(authorize).Filter(auditLog).To(putReplicationControllerSize).
		Doc("Configure the replication controller replica amount").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.PathParameter("replicationcontroller", "Kubernetes replication controller name").DataType("string")).
		Do(returns200, returns400, returns404, returns422, returns500).
		Reads(SizeInput{}))

	ws.Route(ws.POST("/json/{namespace}").Filter(authorize).Filter(auditLog).To(postReplicationControllerFromJson).
		Doc("Add a replication controller in the namespace from json source").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Do(returns200, returns400, returns404, returns422, returns500).
		Reads(new(struct{})))

	ws.Route(ws.PUT("/json/{namespace}/{replicationcontroller}").Filter(authorize).Filter(auditLog).To(putReplicationControllerFromJson).
		Doc("Update a replication controller in the namespace from json source").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.PathParameter("replicationcontroller", "Kubernetes replication controller name").DataType("string")).
		Do(returns200, returns400, returns404, returns422, returns500).
		Reads(new(struct{})))
}

func getAllReplicationController(request *restful.Request, response *restful.Response) {
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

	replicationControllerAndRelatedPodSlice, err := control.GetAllReplicationControllerAndRelatedPodSlice(kubeApiServerEndPoint, kubeApiServerToken, namespace)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get all replication controller and related pods failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeApiServerEndPoint"] = kubeApiServerEndPoint
		jsonMap["namespace"] = namespace
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(replicationControllerAndRelatedPodSlice, "[]ReplicationControllerAndRelatedPodSlice")
}

func postReplicationController(request *restful.Request, response *restful.Response) {
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

	replicationController := new(control.ReplicationController)
	err = request.ReadEntity(&replicationController)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Read body failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeApiServerEndPoint"] = kubeApiServerEndPoint
		jsonMap["namespace"] = namespace
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	err = control.CreateReplicationController(kubeApiServerEndPoint, kubeApiServerToken, namespace, *replicationController)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Create replication controller failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeApiServerEndPoint"] = kubeApiServerEndPoint
		jsonMap["namespace"] = namespace
		jsonMap["replicationController"] = replicationController
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func getReplicationController(request *restful.Request, response *restful.Response) {
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

	replicationController, err := control.GetReplicationController(kubeApiServerEndPoint, kubeApiServerToken, namespace, replicationControllerName)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get replication controller failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeApiServerEndPoint"] = kubeApiServerEndPoint
		jsonMap["namespace"] = namespace
		jsonMap["replicationController"] = replicationControllerName
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}

	response.WriteJson(replicationController, "[]ReplicationController")
}

func deleteReplicationController(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	replicationController := request.PathParameter("replicationcontroller")

	kubeApiServerEndPoint, kubeApiServerToken, err := configuration.GetAvailablekubeApiServerEndPoint()
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get kube apiserver endpoint and token failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["namespace"] = namespace
		jsonMap["replicationController"] = replicationController
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	err = control.DeleteReplicationControllerAndRelatedPod(kubeApiServerEndPoint, kubeApiServerToken, namespace, replicationController)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Delete replication controller failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeApiServerEndPoint"] = kubeApiServerEndPoint
		jsonMap["namespace"] = namespace
		jsonMap["replicationController"] = replicationController
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func putReplicationControllerSize(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	replicationController := request.PathParameter("replicationcontroller")

	kubeApiServerEndPoint, kubeApiServerToken, err := configuration.GetAvailablekubeApiServerEndPoint()
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get kube apiserver endpoint and token failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["namespace"] = namespace
		jsonMap["replicationController"] = replicationController
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	sizeInput := SizeInput{}
	err = request.ReadEntity(&sizeInput)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Read body failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeApiServerEndPoint"] = kubeApiServerEndPoint
		jsonMap["namespace"] = namespace
		jsonMap["replicationController"] = replicationController
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	err = control.UpdateReplicationControllerSize(kubeApiServerEndPoint, kubeApiServerToken, namespace, replicationController, sizeInput.Size)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Resize replication controller failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeApiServerEndPoint"] = kubeApiServerEndPoint
		jsonMap["namespace"] = namespace
		jsonMap["replicationController"] = replicationController
		jsonMap["size"] = sizeInput.Size
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func postReplicationControllerFromJson(request *restful.Request, response *restful.Response) {
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

	replicationController := make(map[string]interface{})
	err = request.ReadEntity(&replicationController)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Read body failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeApiServerEndPoint"] = kubeApiServerEndPoint
		jsonMap["namespace"] = namespace
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	err = control.CreateReplicationControllerWithJson(kubeApiServerEndPoint, kubeApiServerToken, namespace, replicationController)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Create replication controller failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeApiServerEndPoint"] = kubeApiServerEndPoint
		jsonMap["namespace"] = namespace
		jsonMap["replicationController"] = replicationController
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func putReplicationControllerFromJson(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	replicationcontrollerName := request.PathParameter("replicationcontroller")

	kubeApiServerEndPoint, kubeApiServerToken, err := configuration.GetAvailablekubeApiServerEndPoint()
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get kube apiserver endpoint and token failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["namespace"] = namespace
		jsonMap["replicationcontrollerName"] = replicationcontrollerName
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(404, string(errorMessageByteSlice))
		return
	}

	replicationController := make(map[string]interface{})
	err = request.ReadEntity(&replicationController)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Read body failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeApiServerEndPoint"] = kubeApiServerEndPoint
		jsonMap["namespace"] = namespace
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(400, string(errorMessageByteSlice))
		return
	}

	err = control.UpdateReplicationControllerWithJson(kubeApiServerEndPoint, kubeApiServerToken, namespace, replicationcontrollerName, replicationController)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Update replication controller failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeApiServerEndPoint"] = kubeApiServerEndPoint
		jsonMap["namespace"] = namespace
		jsonMap["replicationController"] = replicationController
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}

	podNameSlice, err := control.GetAllPodNameBelongToReplicationController(kubeApiServerEndPoint, kubeApiServerToken, namespace, replicationcontrollerName)
	if err != nil {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Get all pod name belonging to replication controller failure"
		jsonMap["ErrorMessage"] = err.Error()
		jsonMap["kubeApiServerEndPoint"] = kubeApiServerEndPoint
		jsonMap["namespace"] = namespace
		jsonMap["replicationController"] = replicationController
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}

	hasError := false
	errorByteBuffer := bytes.Buffer{}
	for _, podName := range podNameSlice {
		err := control.DeletePod(kubeApiServerEndPoint, kubeApiServerToken, namespace, podName)
		if err != nil {
			hasError = true
			errorByteBuffer.WriteString(err.Error())
			errorByteBuffer.WriteString(" ")
		}
	}

	if hasError {
		jsonMap := make(map[string]interface{})
		jsonMap["Error"] = "Delete pods belonging to replication controller failure"
		jsonMap["ErrorMessage"] = errorByteBuffer.String()
		jsonMap["kubeApiServerEndPoint"] = kubeApiServerEndPoint
		jsonMap["namespace"] = namespace
		jsonMap["replicationController"] = replicationController
		errorMessageByteSlice, _ := json.Marshal(jsonMap)
		log.Error(jsonMap)
		response.WriteErrorString(422, string(errorMessageByteSlice))
		return
	}
}

func returns200AllReplicationController(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", []control.ReplicationControllerAndRelatedPod{})
}

func returns200ReplicationController(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", control.ReplicationController{})
}
