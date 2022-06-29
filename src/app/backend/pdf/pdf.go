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
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/kubernetes/dashboard/src/app/backend/api"
	metricapi "github.com/kubernetes/dashboard/src/app/backend/integration/metric/api"
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/logs"
	"github.com/kubernetes/dashboard/src/app/backend/resource/pod"
	"github.com/phpdave11/gofpdf"
	"github.com/phpdave11/gofpdf/contrib/gofpdi"
)

type Point struct {
	x, y float64
}

type PodResponse struct {
	ListMeta          api.ListMeta          `json:"listMeta"`
	CumulativeMetrics []metricapi.Metric    `json:"cumulativeMetrics"`
	Status            common.ResourceStatus `json:"status"`
	Pods              []pod.PodDetail       `json:"pods"`
}

var pointMap = map[string]Point{
	"title.generated":               {30, 60},
	"title.namespace":               {30, 80},
	"poddetail.name":                {60, 33},
	"poddetail.labels":              {30, 50},
	"poddetail.taints":              {30, 70},
	"poddetail.containers":          {30, 90},
	"poddetail.pvc":                 {30, 130},
	"poddetail.nodes":               {30, 150},
	"poddetail.events":              {30, 170},
	"podlogs.name":                  {60, 33},
	"podlogs.logs":                  {30, 50},
	"pvc.name":                      {60, 33},
	"pvc.state":                     {30, 50},
	"pvc.storageclass":              {30, 70},
	"pvc.volume":                    {30, 90},
	"pvc.labels":                    {30, 110},
	"pvc.capacity":                  {30, 130},
	"pvc.events":                    {30, 150},
	"node.name":                     {65, 33},
	"node.labels":                   {30, 50},
	"node.taints":                   {30, 70},
	"node.osimage":                  {30, 90},
	"node.ip":                       {30, 110},
	"node.schedulable":              {70, 120},
	"node.state.networkunavailable": {90, 137},
	"node.state.memorypressure":     {90, 147},
	"node.state.diskpressure":       {90, 157},
	"node.state.pidpressure":        {90, 167},
	"node.state.ready":              {90, 177},
	"node.events":                   {30, 195},
}

const (
	reportHeight float64 = 297
	reportWidth  float64 = 210
	ReportDir    string  = "/tmp/pdf"
)

func GenerateReport(namespace string) error {
	pdf := gofpdf.New(gofpdf.OrientationPortrait, "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Helvetica", "", 14)

	addTitlePage(pdf, namespace)

	resp, err := getPodDetail(namespace)
	if err != nil {
		log.Printf("Error getting pod detail. Skipping. Error: %v", err)
	} else {
		for _, pod := range resp.Pods {
			log.Printf("Pod detail gotten: %v", pod)
			var labels string = ""
			for key, value := range pod.ObjectMeta.Labels {
				labels += key + ":" + value + ", "
			}
			labels = labels[0 : len(labels)-2]
			logDetail, err := getPodLogs(namespace, pod.ObjectMeta.Name)
			if err != nil {
				log.Printf("Error getting logs for pod %s in namespace %s. Error: %v", pod.ObjectMeta.Name, namespace, err)
			}

			logArr := make([]string, len(logDetail.LogLines))
			for i, value := range logDetail.LogLines {
				logArr[i] = string(value.Timestamp) + "---" + value.Content
			}

			addPodDetailPage(pdf, pod.ObjectMeta.Name, labels, "tbi", pod.ContainerImages, "tbi", pod.NodeName, []string{"tbi"}) // TODO: IMPLEMENT TAINT, PVC, EVENTS
			addPodLogsPage(pdf, pod.ObjectMeta.Name, logArr)
		}
	}

	err = pdf.OutputFileAndClose(ReportDir + "/Report-" + time.Now().Format("01-02-2006_15-04-05") + ".pdf")
	return err
}
func GenerateTestReport() error {
	pdf := gofpdf.New(gofpdf.OrientationPortrait, "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Helvetica", "", 14)

	addTitlePage(pdf, "SAMPLE-NAMESPACE")

	addPodDetailPage(pdf, "SAMPLE-POD-ABCDEF1234567890", "LABEL1, LABEL2, LABEL3, LABEL4, LABEL5", "SAMPLE.SAMPLE/SAMPLE:SAMPLE op=Exists for 300s",
		[]string{"CONTAINER1", "CONTAINER2", "CONTAINER3"}, "SAMPLE-PVC-1", "NODE1, NODE2, NODE3", []string{"EVENT1", "EVENT2", "EVENT3"})

	addPodLogsPage(pdf, "SAMPLE-POD-ABCDEF1234567890", []string{"LOG1", "LOG2", "LOG3", "LOG4", "LOG5"})

	addNodePage(pdf, "NODE1", "LABEL1, LABEL2, LABEL3", "SAMPLE.SAMPLE/SAMPLE:SAMPLE op=Exists for 300s", "Linux Shminux 22.04 LTS", "123.123.123.123",
		false, false, false, false, false, true, []string{"EVENT1", "EVENT2", "EVENT3"})

	addPvcPage(pdf, "SAMPLE-PVC-1", "bound", "local-storage", "SAMPLE-PV-1", "LABEL1, LABEL2, LABEL3", "100Ti", []string{"EVENT1", "EVENT2", "EVENT3"})

	err := pdf.OutputFileAndClose(ReportDir + "/Report-" + time.Now().Format("01-02-2006_15-04-05") + ".pdf")
	return err
}
func addTitlePage(pdf *gofpdf.Fpdf, namespace string) {
	titlePage := gofpdi.ImportPage(pdf, "templates/title_page.pdf", 1, "/MediaBox")
	gofpdi.UseImportedTemplate(pdf, titlePage, 0, 0, reportWidth, reportHeight)
	addText(pdf, "title.generated", time.Now().String())
	addText(pdf, "title.namespace", namespace)
}
func addPodDetailPage(pdf *gofpdf.Fpdf, podName, podLabels, podTaints string, podContainers []string, podPvc, podNodes string, podEvents []string) {
	pdf.AddPage()
	podDetailPage := gofpdi.ImportPage(pdf, "templates/pod_detail.pdf", 1, "/MediaBox")
	gofpdi.UseImportedTemplate(pdf, podDetailPage, 0, 0, reportWidth, reportHeight)
	addText(pdf, "poddetail.name", podName)
	addText(pdf, "poddetail.labels", podLabels)
	addText(pdf, "poddetail.taints", podTaints)
	addMultilineText(pdf, "poddetail.containers", podContainers)
	addText(pdf, "poddetail.pvc", podPvc)
	addText(pdf, "poddetail.nodes", podNodes)
	addMultilineText(pdf, "poddetail.events", podEvents)
}
func addPodLogsPage(pdf *gofpdf.Fpdf, podName string, logs []string) {
	pdf.AddPage()
	podLogsPage := gofpdi.ImportPage(pdf, "templates/pod_logs.pdf", 1, "/MediaBox")
	gofpdi.UseImportedTemplate(pdf, podLogsPage, 0, 0, reportWidth, reportHeight)
	addText(pdf, "podlogs.name", podName)
	addMultilineText(pdf, "podlogs.logs", logs)
}
func addNodePage(pdf *gofpdf.Fpdf, nodeName, labels, taints, osimage, ip string, schedulable, networkunavailable, memorypressure, diskpressure, pidpressure, ready bool, events []string) {
	pdf.AddPage()
	nodePage := gofpdi.ImportPage(pdf, "templates/node_detail.pdf", 1, "/MediaBox")
	gofpdi.UseImportedTemplate(pdf, nodePage, 0, 0, reportWidth, reportHeight)
	addText(pdf, "node.name", nodeName)
	addText(pdf, "node.labels", labels)
	addText(pdf, "node.taints", taints)
	addText(pdf, "node.osimage", osimage)
	addText(pdf, "node.ip", ip)
	addText(pdf, "node.schedulable", strconv.FormatBool(schedulable))
	addText(pdf, "node.state.networkunavailable", strconv.FormatBool(networkunavailable))
	addText(pdf, "node.state.memorypressure", strconv.FormatBool(memorypressure))
	addText(pdf, "node.state.diskpressure", strconv.FormatBool(diskpressure))
	addText(pdf, "node.state.pidpressure", strconv.FormatBool(pidpressure))
	addText(pdf, "node.state.ready", strconv.FormatBool(ready))
	addMultilineText(pdf, "node.events", events)
}
func addPvcPage(pdf *gofpdf.Fpdf, pvcName, state, storageclass, volume, labels, capacity string, events []string) {
	pdf.AddPage()
	pvcPage := gofpdi.ImportPage(pdf, "templates/pvc_detail.pdf", 1, "/MediaBox")
	gofpdi.UseImportedTemplate(pdf, pvcPage, 0, 0, reportWidth, reportHeight)
	addText(pdf, "pvc.name", pvcName)
	addText(pdf, "pvc.state", state)
	addText(pdf, "pvc.storageclass", storageclass)
	addText(pdf, "pvc.volume", volume)
	addText(pdf, "pvc.labels", labels)
	addText(pdf, "pvc.capacity", capacity)
	addMultilineText(pdf, "pvc.events", events)
}
func addText(pdf *gofpdf.Fpdf, key string, text string) {
	location := pointMap[key]
	pdf.Text(location.x, location.y, text)
}
func addMultilineText(pdf *gofpdf.Fpdf, key string, text []string) {
	location := pointMap[key]
	for i, item := range text {
		pdf.Text(location.x, location.y+float64(5*i), item)
	}
}

func getPodDetail(namespace string) (response pod.PodList, err error) {
	resp, err := getHttp("pod/" + namespace)
	if err != nil {
		log.Printf("Error getting pod detail in namespace %s, error: %v", namespace, err)
		return pod.PodList{}, err
	}
	bodyBytes, err := parseHtmlToBytes(resp)
	if err != nil {
		log.Printf("Error parsing html of pod detail in namespace %s, error: %v", namespace, err)
		return pod.PodList{}, err
	}

	var detail pod.PodList = pod.PodList{}
	err = json.Unmarshal(bodyBytes, &detail)
	if err != nil {
		log.Printf("Error parsing json of pod detail in namespace %s, error: %v", namespace, err)
		return pod.PodList{}, err
	}

	return detail, nil
}
func getPodLogs(namespace string, pod string) (response logs.LogDetails, err error) {
	resp, err := getHttp("log/" + namespace + "/" + pod)
	if err != nil {
		log.Printf("Error getting pod detail in namespace %s, error: %v", namespace, err)
		return logs.LogDetails{}, err
	}
	bodyBytes, err := parseHtmlToBytes(resp)
	if err != nil {
		log.Printf("Error parsing html of pod detail in namespace %s, error: %v", namespace, err)
		return logs.LogDetails{}, err
	}

	var logDetails logs.LogDetails = logs.LogDetails{}
	err = json.Unmarshal(bodyBytes, &logDetails)
	if err != nil {
		log.Printf("Error parsing json of pod detail in namespace %s, error: %v", namespace, err)
		return logs.LogDetails{}, err
	}

	return logDetails, nil
}
func getProtocol() string {
	if Secure {
		return "https://"
	} else {
		return "http://"
	}
}
func getHttp(path string) (resp *http.Response, err error) {
	return http.Get(getProtocol() + "localhost:" + fmt.Sprint(ApiPort) + "/api/v1/" + path)
}
func parseHtmlToBytes(response *http.Response) (result []byte, err error) {
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Printf("Error reading http body, error: %v", err)
		return nil, err
	}
	bodyBytes := []byte(body)
	return bodyBytes, nil
}
