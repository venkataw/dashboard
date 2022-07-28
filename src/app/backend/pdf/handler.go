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
	"archive/zip"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/emicklei/go-restful/v3"

	authApi "github.com/kubernetes/dashboard/src/app/backend/auth/api"
)

// types to send to frontend
type pdfDetail struct {
	Name string `json:"name"`
}
type pdfContent struct {
	Contents []byte `json:"contents"`
}
type pdfRequestStatus struct {
	Status       string `json:"status"`
	ErrorMessage string `json:"error"`
	FileName     string `json:"file"`
}
type pdfTemplate struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayname"`
}
type pdfZip struct {
	Status   string `json:"status"`
	Error    string `json:"error"`
	Contents []byte `json:"contents"`
}

var (
	ApiPort      = 9090
	Secure       = false
	tokenManager authApi.TokenManager
)

// list of templates that the frontend can request for generation
var templateList []pdfTemplate = []pdfTemplate{
	{"healthcheck", "Health Check Report"},
	{"test", "Test Report"},
}

// helper functions
func getReportDirListing() []pdfDetail {
	files, err := os.ReadDir(ReportDir)
	if err != nil {
		log.Printf("Failed to list report dir '%s', error: %v", ReportDir, err)
		return nil
	}
	pdfList := make([]pdfDetail, len(files))
	for i, file := range files {
		pdfList[i].Name = file.Name()
	}
	return pdfList
}
func jweFormatCookieString(cookie string) string {
	// jwe decrypt expects json format, so convert cookie format into json
	cookie = strings.ReplaceAll(cookie, "%7B", "{")
	cookie = strings.ReplaceAll(cookie, "%22", "\"")
	cookie = strings.ReplaceAll(cookie, "%3A", ":")
	cookie = strings.ReplaceAll(cookie, "%2C", ",")
	cookie = strings.ReplaceAll(cookie, "%7D", "}")
	return cookie
}

// handler functions
func getPdfList(request *restful.Request, response *restful.Response) {
	log.Printf("Sending list of reports in ReportDir '%s'", ReportDir)

	pdfList := getReportDirListing()

	response.WriteHeaderAndEntity(http.StatusOK, pdfList)
}

func getPdf(request *restful.Request, response *restful.Response) {
	pdfname := request.PathParameter("pdfname")
	log.Printf("Sending pdf '%s'", pdfname)

	content, err := os.ReadFile(ReportDir + "/" + pdfname)
	if errors.Is(err, os.ErrNotExist) {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, pdfContent{Contents: nil})
	} else if err != nil {
		panic(err)
	} else {
		response.WriteHeaderAndEntity(http.StatusOK, pdfContent{Contents: content})
	}
}

func getTemplates(_ *restful.Request, response *restful.Response) {
	log.Print("Sending template list")
	response.WriteHeaderAndEntity(http.StatusOK, templateList)
}

// Feedback for pdf generation (pdf generated = OK, some error = 500 + message)
// TODO: 500 error might be problematic for frontend; check back later
func genHealthCheckPdf(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	log.Printf("Generating health check pdf for %v", namespace)

	// decrypt bearer token from cookie
	cookie, err := request.Request.Cookie("jweToken")
	if err != nil {
		log.Printf("Error getting cookie 'jweToken' from request: %v", err)
		log.Printf("Trying to generate health check report without bearer token")
		setBearerToken("")
	} else {
		encrypted := jweFormatCookieString(cookie.Value)
		authInfo, err := tokenManager.Decrypt(encrypted)
		if err != nil {
			log.Printf("Error decrypting bearer token: %v", err)
		}
		setBearerToken(authInfo.Token)
	}

	fileName, err := GenerateHealthCheckReport(namespace)

	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, pdfRequestStatus{Status: "error", ErrorMessage: fmt.Sprint(err)})
	} else {
		response.WriteHeaderAndEntity(http.StatusOK, pdfRequestStatus{Status: "ok", FileName: fileName})
	}
}

func genTestPdf(_ *restful.Request, response *restful.Response) {
	log.Printf("Generating test pdf")

	fileName, err := GenerateTestReport()

	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, pdfRequestStatus{Status: "error", ErrorMessage: fmt.Sprint(err)})
	} else {
		response.WriteHeaderAndEntity(http.StatusOK, pdfRequestStatus{Status: "ok", FileName: fileName})
	}
}

func zipAllReports(_ *restful.Request, response *restful.Response) {
	log.Printf("Zipping all reports")
	reports := getReportDirListing()
	archive, err := os.Create(ReportDir + "/archive.zip")
	if err != nil {
		log.Printf("Unable to create zip archive, error: %v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, pdfZip{Status: "error", Error: fmt.Sprint(err)})
		return
	}
	writer := zip.NewWriter(archive)
	for _, report := range reports {
		// open file
		file, err := os.Open(ReportDir + "/" + report.Name)
		if err != nil {
			log.Printf("Failed to open file %s, skipping. Error: %v", report.Name, err)
			continue
		}
		// write file to archive
		write, err := writer.Create(report.Name)
		if err != nil {
			log.Printf("Failed to create file %s in zip archive, error: %v", report.Name, err)
			continue
		}
		if _, err := io.Copy(write, file); err != nil {
			log.Printf("Error copying file %s to zip archive: %v", report.Name, err)
		}
		file.Close()
	}
	writer.Close()
	archive.Close()

	// read zip contents and send
	content, err := os.ReadFile(ReportDir + "/archive.zip")
	if err != nil {
		log.Printf("Error reading zip contents archive.zip: %v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, pdfZip{Status: "error", Error: fmt.Sprint(err)})
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, pdfZip{Status: "ok", Contents: content})

	// delete archive.zip
	err = os.Remove(ReportDir + "/archive.zip")
	if err != nil {
		log.Printf("Error deleting archive.zip: %v", err)
	}
}

func deleteReport(request *restful.Request, response *restful.Response) {
	file := request.PathParameter("file")
	log.Printf("Deleting file %s from report dir", file)

	err := os.Remove(ReportDir + "/" + file)
	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, pdfRequestStatus{Status: "error", ErrorMessage: fmt.Sprint(err)})
	} else {
		response.WriteHeaderAndEntity(http.StatusOK, pdfRequestStatus{Status: "ok"})
	}
}

func SetTokenManager(tokenMgr authApi.TokenManager) {
	tokenManager = tokenMgr
}

func CreatePdfApiHandler(port int, isSecure bool) (http.Handler, error) {
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
			Writes([]pdfDetail{}))
	pdfApiWs.Route(
		pdfApiWs.GET("/pdf/{pdfname}").
			To(getPdf).
			Writes(pdfContent{}))
	pdfApiWs.Route(
		pdfApiWs.GET("/templates").
			To(getTemplates).
			Writes([]pdfTemplate{}))
	pdfApiWs.Route(
		pdfApiWs.GET("/gen").
			To(genTestPdf).
			Writes(pdfRequestStatus{}))
	pdfApiWs.Route(
		pdfApiWs.GET("/gen/test").
			To(genTestPdf).
			Writes(pdfRequestStatus{}))
	pdfApiWs.Route(
		pdfApiWs.GET("/gen/test/{namespace}").
			To(genTestPdf).
			Writes(pdfRequestStatus{}))
	pdfApiWs.Route(
		pdfApiWs.GET("/gen/healthcheck/{namespace}").
			To(genHealthCheckPdf).
			Writes(pdfRequestStatus{}))
	pdfApiWs.Route(
		pdfApiWs.GET("/zip").
			To(zipAllReports).
			Writes(pdfZip{}))
	pdfApiWs.Route(
		pdfApiWs.GET("/delete/{file}").
			To(deleteReport).
			Writes(pdfRequestStatus{}))

	// TODO: Remove this. Ignores self-signed certs for testing purposes only
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	// initialize http client in clusterinfo.go.
	initializeHttpClient()

	log.Print("PDF API handler initialized.")

	return wsContainer, nil
}
