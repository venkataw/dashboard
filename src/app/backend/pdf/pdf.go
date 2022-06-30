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
	"github.com/kubernetes/dashboard/src/app/backend/resource/node"
	"github.com/kubernetes/dashboard/src/app/backend/resource/persistentvolumeclaim"
	"github.com/kubernetes/dashboard/src/app/backend/resource/pod"
	"github.com/phpdave11/gofpdf"
	"github.com/phpdave11/gofpdf/contrib/gofpdi"
	v1 "k8s.io/api/core/v1"
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

var importer *gofpdi.Importer

const (
	reportHeight float64 = 297
	reportWidth  float64 = 210
	ReportDir    string  = "/tmp/pdf"
)

func GenerateReport(namespace string) error {
	pdf := gofpdf.New(gofpdf.OrientationPortrait, "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Helvetica", "", 12)

	// import templates
	importer = gofpdi.NewImporter()
	titlePageId := importer.ImportPage(pdf, "templates/title_page.pdf", 1, "/MediaBox")
	podDetailPageId := importer.ImportPage(pdf, "templates/pod_detail.pdf", 1, "/MediaBox")
	podLogsPageId := importer.ImportPage(pdf, "templates/pod_logs.pdf", 1, "/MediaBox")
	nodePageId := importer.ImportPage(pdf, "templates/node_detail.pdf", 1, "/MediaBox")
	pvcPageId := importer.ImportPage(pdf, "templates/pvc_detail.pdf", 1, "/MediaBox")

	addTitlePage(pdf, titlePageId, namespace)

	// pod detail and logs
	resp, err := getPodDetail(namespace)
	if err != nil {
		log.Printf("Error getting pod detail. Skipping. Error: %v", err)
	} else {
		for _, pod := range resp.Pods {
			log.Printf("Pod detail gotten: %v", pod)

			labels := formatLabelString(pod.ObjectMeta.Labels)
			logDetail, err := getPodLogs(namespace, pod.ObjectMeta.Name)
			if err != nil {
				log.Printf("Error getting logs for pod %s in namespace %s. Error: %v", pod.ObjectMeta.Name, namespace, err)
			}

			logArr := make([]string, len(logDetail.LogLines))
			for i, value := range logDetail.LogLines {
				logArr[i] = string(value.Timestamp) + "---" + value.Content
			}

			// TODO: IMPLEMENT TAINT, PVC, EVENTS
			addPodDetailPage(pdf, podDetailPageId, pod.ObjectMeta.Name, labels, "tbi", pod.ContainerImages, "tbi", pod.NodeName, []string{"tbi"})
			addPodLogsPage(pdf, podLogsPageId, pod.ObjectMeta.Name, logArr)
		}
	}

	nodes, err := getNodeList()
	if err != nil {
		log.Printf("Error getting node list. Skipping. Error: %v", err)
	} else {
		for _, node := range nodes.Nodes {
			// get more specific
			nodeInfo, err := getNodeDetail(node.ObjectMeta.Name)
			if err != nil {
				log.Printf("Error getting more specific node detail for %s, skipping. Error: %v", node.ObjectMeta.Name, err)
			}
			log.Printf("node detail gotten: %v", nodeInfo)
			labels := formatLabelString(node.ObjectMeta.Labels)
			taints := formatTaintString(nodeInfo.Taints)
			internalIps := formatInternalIpString(nodeInfo.Addresses)
			events := formatEventListArray(nodeInfo.EventList)
			// TODO: IMPLEMENT NODE PRESSURE INFO
			addNodePage(pdf, nodePageId, node.ObjectMeta.Name, labels, taints, nodeInfo.NodeInfo.OSImage, internalIps, !nodeInfo.Unschedulable, false, false, false, false, string(nodeInfo.Ready), events)
		}
	}

	pvcList, err := getPvcDetail(namespace)
	if err != nil {
		log.Printf("Error getting pvc list. Skipping. Error: %v", err)
	} else {
		for _, pvc := range pvcList.Items {
			log.Printf("pvc detail gotten, %v", pvc)
			labels := formatLabelString(pvc.ObjectMeta.Labels)
			addPvcPage(pdf, pvcPageId, pvc.ObjectMeta.Name, pvc.Status, *pvc.StorageClass, pvc.Volume, labels, fmt.Sprint(pvc.Capacity.Storage()), []string{"tbi"})
		}
	}

	err = pdf.OutputFileAndClose(ReportDir + "/HealthCheck-" + namespace + "-" + time.Now().Format("01-02-2006_15-04-05") + ".pdf")
	return err
}
func GenerateTestReport() error {
	pdf := gofpdf.New(gofpdf.OrientationPortrait, "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Helvetica", "", 12)

	// import templates
	importer = gofpdi.NewImporter()
	titlePageId := importer.ImportPage(pdf, "templates/title_page.pdf", 1, "/MediaBox")
	podDetailPageId := importer.ImportPage(pdf, "templates/pod_detail.pdf", 1, "/MediaBox")
	podLogsPageId := importer.ImportPage(pdf, "templates/pod_logs.pdf", 1, "/MediaBox")
	nodePageId := importer.ImportPage(pdf, "templates/node_detail.pdf", 1, "/MediaBox")
	pvcPageId := importer.ImportPage(pdf, "templates/pvc_detail.pdf", 1, "/MediaBox")

	addTitlePage(pdf, titlePageId, "SAMPLE-NAMESPACE")

	addPodDetailPage(pdf, podDetailPageId, "SAMPLE-POD-ABCDEF1234567890", "LABEL1, LABEL2, LABEL3, LABEL4, LABEL5", "SAMPLE.SAMPLE/SAMPLE:SAMPLE op=Exists for 300s",
		[]string{"CONTAINER1", "CONTAINER2", "CONTAINER3"}, "SAMPLE-PVC-1", "NODE1, NODE2, NODE3", []string{"EVENT1", "EVENT2", "EVENT3"})

	addPodLogsPage(pdf, podLogsPageId, "SAMPLE-POD-ABCDEF1234567890", []string{"LOG1", "LOG2", "LOG3", "LOG4", "LOG5"})

	addNodePage(pdf, nodePageId, "NODE1", "LABEL1, LABEL2, LABEL3", "SAMPLE.SAMPLE/SAMPLE:SAMPLE op=Exists for 300s", "Linux Shminux 22.04 LTS", "123.123.123.123",
		false, false, false, false, false, "True", []string{"EVENT1", "EVENT2", "EVENT3"})

	addPvcPage(pdf, pvcPageId, "SAMPLE-PVC-1", "bound", "local-storage", "SAMPLE-PV-1", "LABEL1, LABEL2, LABEL3", "100Ti", []string{"EVENT1", "EVENT2", "EVENT3"})

	err := pdf.OutputFileAndClose(ReportDir + "/Test-SAMPLE-NAMESPACE-" + time.Now().Format("01-02-2006_15-04-05") + ".pdf")
	return err
}
func addTitlePage(pdf *gofpdf.Fpdf, titlePageId int, namespace string) {
	importer.UseImportedTemplate(pdf, titlePageId, 0, 0, reportWidth, reportHeight)
	addText(pdf, "title.generated", time.Now().String())
	addText(pdf, "title.namespace", namespace)
}
func addPodDetailPage(pdf *gofpdf.Fpdf, podDetailPageId int, podName, podLabels, podTaints string, podContainers []string, podPvc, podNodes string, podEvents []string) {
	pdf.AddPage()
	importer.UseImportedTemplate(pdf, podDetailPageId, 0, 0, reportWidth, reportHeight)
	addText(pdf, "poddetail.name", podName)
	addText(pdf, "poddetail.labels", podLabels)
	addText(pdf, "poddetail.taints", podTaints)
	addMultilineText(pdf, "poddetail.containers", podContainers)
	addText(pdf, "poddetail.pvc", podPvc)
	addText(pdf, "poddetail.nodes", podNodes)
	addMultilineText(pdf, "poddetail.events", podEvents)
}
func addPodLogsPage(pdf *gofpdf.Fpdf, podLogsPageId int, podName string, logs []string) {
	pdf.AddPage()
	importer.UseImportedTemplate(pdf, podLogsPageId, 0, 0, reportWidth, reportHeight)
	addText(pdf, "podlogs.name", podName)
	addMultilineText(pdf, "podlogs.logs", logs)
}
func addNodePage(pdf *gofpdf.Fpdf, nodePageId int, nodeName, labels, taints, osimage, ip string, schedulable, networkunavailable, memorypressure, diskpressure, pidpressure bool, ready string, events []string) {
	pdf.AddPage()
	importer.UseImportedTemplate(pdf, nodePageId, 0, 0, reportWidth, reportHeight)
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
	addText(pdf, "node.state.ready", ready)
	addMultilineText(pdf, "node.events", events)
}
func addPvcPage(pdf *gofpdf.Fpdf, pvcPageId int, pvcName, state, storageclass, volume, labels, capacity string, events []string) {
	pdf.AddPage()
	importer.UseImportedTemplate(pdf, pvcPageId, 0, 0, reportWidth, reportHeight)
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
		log.Printf("Error getting pod logs in namespace %s, error: %v", namespace, err)
		return logs.LogDetails{}, err
	}
	bodyBytes, err := parseHtmlToBytes(resp)
	if err != nil {
		log.Printf("Error parsing html of pod logs in namespace %s, error: %v", namespace, err)
		return logs.LogDetails{}, err
	}

	var logDetails logs.LogDetails = logs.LogDetails{}
	err = json.Unmarshal(bodyBytes, &logDetails)
	if err != nil {
		log.Printf("Error parsing json of pod logs in namespace %s, error: %v", namespace, err)
		return logs.LogDetails{}, err
	}

	return logDetails, nil
}
func getNodeList() (nodes node.NodeList, err error) {
	resp, err := getHttp("node")
	if err != nil {
		log.Printf("Error getting node list, error: %v", err)
		return node.NodeList{}, err
	}
	bodyBytes, err := parseHtmlToBytes(resp)
	if err != nil {
		log.Printf("Error parsing html of node list, error: %v", err)
		return node.NodeList{}, err
	}

	var nodeDetails node.NodeList = node.NodeList{}
	err = json.Unmarshal(bodyBytes, &nodeDetails)
	if err != nil {
		log.Printf("Error parsing json of node list, error: %v", err)
		return node.NodeList{}, err
	}

	return nodeDetails, nil
}
func getNodeDetail(nodeName string) (nodeDetail node.NodeDetail, err error) {
	resp, err := getHttp("node/" + nodeName)
	if err != nil {
		log.Printf("Error getting node detail for node %s, error: %v", nodeName, err)
		return node.NodeDetail{}, err
	}
	bodyBytes, err := parseHtmlToBytes(resp)
	if err != nil {
		log.Printf("Error parsing html of node detail %s, error: %v", nodeName, err)
		return node.NodeDetail{}, err
	}

	var nodeDetails node.NodeDetail = node.NodeDetail{}
	err = json.Unmarshal(bodyBytes, &nodeDetails)
	if err != nil {
		log.Printf("Error parsing json of node detail %s, error: %v", nodeName, err)
		return node.NodeDetail{}, err
	}

	return nodeDetails, nil
}
func getPvcDetail(namespace string) (claimList persistentvolumeclaim.PersistentVolumeClaimList, err error) {
	resp, err := getHttp("persistentvolumeclaim/" + namespace)
	if err != nil {
		log.Printf("Error getting pvc list for namespace %s, error: %v", namespace, err)
		return persistentvolumeclaim.PersistentVolumeClaimList{}, err
	}
	bodyBytes, err := parseHtmlToBytes(resp)
	if err != nil {
		log.Printf("Error parsing html of pvc list in namespace %s, error: %v", namespace, err)
		return persistentvolumeclaim.PersistentVolumeClaimList{}, err
	}

	var pvcList persistentvolumeclaim.PersistentVolumeClaimList = persistentvolumeclaim.PersistentVolumeClaimList{}
	err = json.Unmarshal(bodyBytes, &pvcList)
	if err != nil {
		log.Printf("Error parsing json of pvc list in namespace %s, error: %v", namespace, err)
		return persistentvolumeclaim.PersistentVolumeClaimList{}, err
	}

	return pvcList, nil
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
func formatLabelString(labels map[string]string) string {
	var formatted string = ""
	if len(labels) == 0 {
		return "No labels"
	}
	for key, value := range labels {
		formatted += key + ":" + value + ", "
	}
	formatted = formatted[0 : len(formatted)-2]
	return formatted
}
func formatTaintString(taints []v1.Taint) string {
	var formatted string = ""
	if len(taints) == 0 {
		return "No taints"
	}
	for _, taint := range taints {
		formatted += taint.Key + ":" + taint.Value + ", "
	}
	formatted = formatted[0 : len(formatted)-2]
	return formatted
}
func formatInternalIpString(addresses []v1.NodeAddress) string {
	var formatted string = ""
	if len(addresses) == 0 {
		return "No addresses"
	}
	for _, address := range addresses {
		if address.Type == "InternalIP" {
			formatted += address.Address + ", "
		}
	}
	formatted = formatted[0 : len(formatted)-2]
	return formatted
}
func formatEventListArray(events common.EventList) []string {
	eventArr := make([]string, len(events.Events))
	for i, event := range events.Events {
		eventArr[i] = event.Message + ", Reason: " + event.Reason // TODO: ADD MORE DETAILS TO EVENT LIST
	}
	if len(eventArr) == 0 {
		return []string{"No events"}
	}
	return eventArr
}
