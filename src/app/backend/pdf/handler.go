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

func getPdfList(request *restful.Request, response *restful.Response) {
	fmt.Print("Got request for pdf list. Request: ")
	fmt.Println(request)

	files, err := os.ReadDir(ReportDir)
	if err != nil {
		panic(err)
	}
	pdfList := make([]pdfDetail, len(files))
	for i, file := range files {
		pdfList[i].Name = file.Name()
	}
	fmt.Print("pdfList built. Contents sending: ")
	fmt.Println(pdfList)

	response.WriteHeaderAndEntity(http.StatusOK, pdfList)
}

func getPdf(request *restful.Request, response *restful.Response) {
	fmt.Print("Got request for a pdf. Request: ")
	fmt.Println(request)
	pdfname := request.PathParameter("pdfname")
	fmt.Print("Want pdf: ")
	fmt.Println(pdfname)

	content, err := os.ReadFile(ReportDir + "/" + pdfname)
	if errors.Is(err, os.ErrNotExist) {
		response.WriteHeaderAndEntity(http.StatusOK, pdfContent{Contents: nil})
	} else if err != nil {
		panic(err)
	} else {
		response.WriteHeaderAndEntity(http.StatusOK, pdfContent{Contents: content})
	}
}

func genPdf(request *restful.Request, response *restful.Response) {
	fmt.Print("Got request to generate a pdf. Request: ")
	fmt.Println(request)
	namespace := request.PathParameter("namespace")
	fmt.Print("Want from namespace: ")
	fmt.Println(namespace)

	// TODO: GENERATE HEALTH CHECK PDF ON DEMAND

	response.WriteHeader(http.StatusOK)
}

func genTestPdf(request *restful.Request, response *restful.Response) {
	fmt.Print("Got request to generate a TEST pdf. Request: ")
	fmt.Println(request)

	GenerateTestReport() // TODO: this doesn't seem to work correctly

	response.WriteHeader(http.StatusOK)
}

func CreatePdfApiHandler() (http.Handler, error) {
	fmt.Println("Initializing pdf api handler...")

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
			To(genTestPdf))
	pdfApiWs.Route(
		pdfApiWs.GET("/gen/{namespace}").
			To(genPdf))

	fmt.Println("pdf api handler initialized.")

	return wsContainer, nil
}
