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

type SizeInput struct {
	Size int
}

func registerWebServiceKubernetesService() {
	ws := new(restful.WebService)
	ws.Path("/api/v1/replicationcontrollers")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	restful.Add(ws)

	ws.Route(ws.GET("/{namespace}").To(getAllReplicationController).
		Doc("Get all replication controllers in the namespace").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.QueryParameter("kubeapihost", "Kubernetes host").DataType("string")).
		Param(ws.QueryParameter("kubeapiport", "Kubernetes port").DataType("int")).
		Do(returns200AllReplicationController, returns400, returns404, returns500))

	ws.Route(ws.POST("/{namespace}").To(postReplicationController).
		Doc("Add a replication controller in the namespace").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.QueryParameter("kubeapihost", "Kubernetes host").DataType("string")).
		Param(ws.QueryParameter("kubeapiport", "Kubernetes port").DataType("int")).
		Do(returns200, returns400, returns404, returns500).
		Reads(control.ReplicationController{}))

	ws.Route(ws.GET("/{namespace}/{replicationcontroller}").To(getReplicationController).
		Doc("Get replication controllers in the namespace").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.PathParameter("replicationcontroller", "Kubernetes replication controller name").DataType("string")).
		Param(ws.QueryParameter("kubeapihost", "Kubernetes host").DataType("string")).
		Param(ws.QueryParameter("kubeapiport", "Kubernetes port").DataType("int")).
		Do(returns200ReplicationController, returns400, returns404, returns500))

	ws.Route(ws.DELETE("/{namespace}/{replicationcontroller}").To(deleteReplicationController).
		Doc("Delete the replication controller in the namespace").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.PathParameter("replicationcontroller", "Kubernetes replication controller name").DataType("string")).
		Param(ws.QueryParameter("kubeapihost", "Kubernetes host").DataType("string")).
		Param(ws.QueryParameter("kubeapiport", "Kubernetes port").DataType("int")).
		Do(returns200, returns400, returns404, returns500))

	ws.Route(ws.PUT("/size/{namespace}/{replicationcontroller}").To(putReplicationControllerSize).
		Doc("Configure the replication controller replica amount").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.PathParameter("replicationcontroller", "Kubernetes replication controller name").DataType("string")).
		Param(ws.QueryParameter("kubeapihost", "Kubernetes host").DataType("string")).
		Param(ws.QueryParameter("kubeapiport", "Kubernetes port").DataType("int")).
		Do(returns200, returns400, returns404, returns500).
		Reads(SizeInput{}))

	ws.Route(ws.POST("/json/{namespace}").To(postReplicationControllerFromJson).
		Doc("Add a replication controller in the namespace from json source").
		Param(ws.PathParameter("namespace", "Kubernetes namespace").DataType("string")).
		Param(ws.QueryParameter("kubeapihost", "Kubernetes host").DataType("string")).
		Param(ws.QueryParameter("kubeapiport", "Kubernetes port").DataType("int")).
		Do(returns200, returns400, returns404, returns500).
		Reads(new(struct{})))
}

func getAllReplicationController(request *restful.Request, response *restful.Response) {
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

	replicationControllerAndRelatedPodSlice, err := control.GetAllReplicationControllerAndRelatedPodSlice(kubeapiHost, kubeapiPort, namespace)
	if err != nil {
		errorText := fmt.Sprintf("Could not load all replication controller and related pod with kubeapiHost %s kubeapiPort %d namespace %s error %s", kubeapiHost, kubeapiPort, namespace, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}

	response.WriteJson(replicationControllerAndRelatedPodSlice, "[]ReplicationControllerAndRelatedPodSlice")
}

func postReplicationController(request *restful.Request, response *restful.Response) {
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

	replicationController := new(control.ReplicationController)
	err = request.ReadEntity(&replicationController)

	if err != nil {
		errorText := fmt.Sprintf("POST namespace %s kubeapiHost %s kubeapiPort %s failure with error %s", namespace, kubeapiHost, kubeapiPort, err)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	err = control.CreateReplicationController(kubeapiHost, kubeapiPort, namespace, *replicationController)

	if err != nil {
		errorText := fmt.Sprintf("Create replication controller failure kubeapiHost %s kubeapiPort %s namespace %s replication controller %s error %s", kubeapiHost, kubeapiPort, namespace, replicationController, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
}

func getReplicationController(request *restful.Request, response *restful.Response) {
	kubeapiHost := request.QueryParameter("kubeapihost")
	kubeapiPortText := request.QueryParameter("kubeapiport")
	namespace := request.PathParameter("namespace")
	replicationcontroller := request.PathParameter("replicationcontroller")
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

	replicationController, err := control.GetReplicationController(kubeapiHost, kubeapiPort, namespace, replicationcontroller)
	if err != nil {
		errorText := fmt.Sprintf("Could not get replication controller %s with kubeapiHost %s kubeapiPort %d namespace %s error %s", replicationcontroller, kubeapiHost, kubeapiPort, namespace, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}

	response.WriteJson(replicationController, "[]ReplicationController")
}

func deleteReplicationController(request *restful.Request, response *restful.Response) {
	kubeapiHost := request.QueryParameter("kubeapihost")
	kubeapiPortText := request.QueryParameter("kubeapiport")
	namespace := request.PathParameter("namespace")
	replicationcontroller := request.PathParameter("replicationcontroller")
	if kubeapiHost == "" || kubeapiPortText == "" || namespace == "" || replicationcontroller == "" {
		errorText := fmt.Sprintf("Input text is incorrect kubeapiHost %s kubeapiHost %s namespace %s replication controller %s", kubeapiHost, kubeapiPortText, namespace, replicationcontroller)
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

	err = control.DeleteReplicationControllerAndRelatedPod(kubeapiHost, kubeapiPort, namespace, replicationcontroller)

	if err != nil {
		errorText := fmt.Sprintf("Delete replication controller failure kubeapiHost %s kubeapiPort %s namespace %s replication controller %s", kubeapiHost, kubeapiPort, namespace, replicationcontroller)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
}

func putReplicationControllerSize(request *restful.Request, response *restful.Response) {
	kubeapiHost := request.QueryParameter("kubeapihost")
	kubeapiPortText := request.QueryParameter("kubeapiport")
	namespace := request.PathParameter("namespace")
	replicationcontroller := request.PathParameter("replicationcontroller")
	if kubeapiHost == "" || kubeapiPortText == "" || namespace == "" || replicationcontroller == "" {
		errorText := fmt.Sprintf("Input text is incorrect kubeapiHost %s kubeapiHost %s namespace %s replication controller %s", kubeapiHost, kubeapiPortText, namespace, replicationcontroller)
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

	sizeInput := SizeInput{}
	err = request.ReadEntity(&sizeInput)

	if err != nil {
		errorText := fmt.Sprintf("PUT namespace %s replicationcontroller %s kubeapiHost %s kubeapiPort %s failure with error %s", namespace, replicationcontroller, kubeapiHost, kubeapiPort, err)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	err = control.UpdateReplicationControllerSize(kubeapiHost, kubeapiPort, namespace, replicationcontroller, sizeInput.Size)

	if err != nil {
		errorText := fmt.Sprintf("Fail to resize to size %d namespace %s replicationcontroller %s kubeapiHost %s kubeapiPort %s failure with error %s", sizeInput.Size, namespace, replicationcontroller, kubeapiHost, kubeapiPort, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
}

func postReplicationControllerFromJson(request *restful.Request, response *restful.Response) {
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

	replicationController := make(map[string]interface{})
	err = request.ReadEntity(&replicationController)

	if err != nil {
		errorText := fmt.Sprintf("POST namespace %s kubeapiHost %s kubeapiPort %s failure with error %s", namespace, kubeapiHost, kubeapiPort, err)
		log.Error(errorText)
		response.WriteErrorString(400, `{"Error": "`+errorText+`"}`)
		return
	}

	err = control.CreateReplicationControllerWithJson(kubeapiHost, kubeapiPort, namespace, replicationController)

	if err != nil {
		errorText := fmt.Sprintf("Create replication controller failure kubeapiHost %s kubeapiPort %s namespace %s replication controller %s error %s", kubeapiHost, kubeapiPort, namespace, replicationController, err)
		log.Error(errorText)
		response.WriteErrorString(404, `{"Error": "`+errorText+`"}`)
		return
	}
}

func returns200AllReplicationController(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", []control.ReplicationControllerAndRelatedPod{})
}

func returns200ReplicationController(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", control.ReplicationController{})
}
