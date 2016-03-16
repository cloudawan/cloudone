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
	"github.com/cloudawan/cloudone/utility/configuration"
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful/swagger"
	"net/http"
	"strconv"
)

func StartRestAPIServer() {
	registerWebServiceReplicationControllerMetric()
	registerWebServiceReplicationControllerAutoScaler()
	registerWebServiceKubernetesService()
	registerWebServiceReplicationController()
	registerWebServiceImageInformation()
	registerWebServiceImageRecord()
	registerWebServiceDeploy()
	registerWebServiceDeployBlueGreen()
	registerWebServiceDeployClusterApplication()
	registerWebServiceNodeMetric()
	registerWebServiceNamespace()
	registerWebServicePodLog()
	registerWebServiceReplicationControllerNotifier()
	registerWebServiceStatelessApplication()
	registerWebServiceClusterApplication()
	registerWebServiceGlusterfs()
	registerWebServiceHealthCheck()
	registerWebServiceHost()

	// You can install the Swagger Service which provides a nice Web UI on your REST API
	// You need to download the Swagger HTML5 assets and change the FilePath location in the config below.
	// Open http://localhost:8080/apidocs and enter http://localhost:8080/apidocs.json in the api input field.
	config := swagger.Config{
		WebServices: restful.DefaultContainer.RegisteredWebServices(), // you control what services are visible
		//WebServicesUrl: "http://localhost:8080",
		ApiPath: "/apidocs.json",

		// Optionally, specifiy where the UI is located
		SwaggerPath:     "/apidocs/",
		SwaggerFilePath: "swaggerui"}
	swagger.RegisterSwaggerService(config, restful.DefaultContainer)

	restapiPort, ok := configuration.LocalConfiguration.GetInt("restapiPort")
	if ok == false {
		log.Error("Can't find restapiPort")
		panic("Can't find restapiPort")
	}

	server := &http.Server{Addr: ":" + strconv.Itoa(restapiPort), Handler: restful.DefaultContainer}

	certificate, ok := configuration.LocalConfiguration.GetString("certificate")
	if ok == false {
		log.Error("Can't find certificate path")
		panic("Can't find certificate path")
	}
	key, ok := configuration.LocalConfiguration.GetString("key")
	if ok == false {
		log.Error("Can't find certificate path")
		panic("Can't find key path")
	}
	server.ListenAndServeTLS(certificate, key)
}

func returns200(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", nil)
}

func returns400(b *restful.RouteBuilder) {
	b.Returns(http.StatusBadRequest, "Bad request", nil)
}

func returns404(b *restful.RouteBuilder) {
	b.Returns(http.StatusNotFound, "Not found", nil)
}

func returns500(b *restful.RouteBuilder) {
	b.Returns(http.StatusInternalServerError, "Internal error", nil)
}
