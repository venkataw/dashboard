// Copyright 2017 The Kubernetes Authors.
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

package pdf

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/emicklei/go-restful/v3"
)

type pdfDetail struct {
	Name string `json:"name"`
}
type pdfContent struct {
	Contents []byte `json:"contents"`
}
type pdfRequestStatus struct {
	Status       string `json:"status"`
	ErrorMessage string `json:"error"`
}

var (
	ApiPort = 9090
	Secure  = false
)

func getPdfList(request *restful.Request, response *restful.Response) {
	log.Printf("Got request for pdf list. Request: %v", request)

	files, err := os.ReadDir(ReportDir)
	if err != nil {
		panic(err)
	}
	pdfList := make([]pdfDetail, len(files))
	for i, file := range files {
		pdfList[i].Name = file.Name()
	}
	log.Printf("pdfList built. Contents sending: %v", pdfList)

	response.WriteHeaderAndEntity(http.StatusOK, pdfList)
}

func getPdf(request *restful.Request, response *restful.Response) {
	log.Printf("Got request for a pdf. Request: %v", request)
	pdfname := request.PathParameter("pdfname")
	log.Printf("Want pdf: %v", pdfname)

	content, err := os.ReadFile(ReportDir + "/" + pdfname)
	if errors.Is(err, os.ErrNotExist) {
		response.WriteHeaderAndEntity(http.StatusOK, pdfContent{Contents: nil})
	} else if err != nil {
		panic(err)
	} else {
		response.WriteHeaderAndEntity(http.StatusOK, pdfContent{Contents: content})
	}
}

// Feedback for pdf generation (pdf generated = OK, some error = 500 + message)
// TODO: 500 error might be problematic for frontend; check back later
func genPdf(request *restful.Request, response *restful.Response) {
	log.Printf("Got request to generate a pdf. Request: %v", request)
	namespace := request.PathParameter("namespace")
	log.Printf("Want from namespace: %v", namespace)

	err := GenerateHealthCheckReport(namespace)

	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, pdfRequestStatus{Status: "error", ErrorMessage: fmt.Sprint(err)})
	} else {
		response.WriteHeaderAndEntity(http.StatusOK, pdfRequestStatus{Status: "ok"})
	}
}

func genTestPdf(request *restful.Request, response *restful.Response) {
	log.Printf("Got request to generate a TEST pdf. Request: %v", request)

	GenerateTestReport()

	response.WriteHeader(http.StatusOK)
}

func CreatePdfApiHandler(port int, isSecure bool) (http.Handler, error) {
	log.Print("Initializing pdf api handler...")

	ApiPort = port
	Secure = isSecure

	err := os.MkdirAll(ReportDir, os.ModePerm)
	if err != nil {
		panic(err)
	}

	wsContainer := restful.NewContainer()
	pdfApiWs := new(restful.WebService)
	pdfApiWs.Path("/pdf").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	wsContainer.Add(pdfApiWs)

	pdfApiWs.Route(
		pdfApiWs.GET("/").
			To(getPdfList).
			Writes(pdfDetail{}))
	pdfApiWs.Route(
		pdfApiWs.GET("/pdf/{pdfname}").
			To(getPdf).
			Writes(pdfContent{}))
	pdfApiWs.Route(
		pdfApiWs.GET("/gen").
			To(genTestPdf).
			Writes(pdfRequestStatus{}))
	pdfApiWs.Route(
		pdfApiWs.GET("/gen/{namespace}").
			To(genPdf).
			Writes(pdfRequestStatus{}))

	log.Print("pdf api handler initialized.")

	return wsContainer, nil
}
