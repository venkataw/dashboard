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
	"strconv"
	"time"

	"github.com/phpdave11/gofpdf"
	"github.com/phpdave11/gofpdf/contrib/gofpdi"
)

type Point struct {
	x, y float64
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

func GenerateHealthCheckReport(namespace string) error {
	// Check if namespace exists first
	exists, _ := namespaceExists(namespace)
	if !exists {
		log.Printf("Namespace %s doesn't exist, cannot create health check report.", namespace)
		return errors.New("Namespace " + namespace + " doesn't exist")
	}

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

			events, err := getPodEvents(namespace, pod.ObjectMeta.Name)
			if err != nil {
				log.Printf("Error getting events for pod %s in namespace %s. Error: %v", pod.ObjectMeta.Name, namespace, err)
			}
			log.Printf("Pod events gotten: %v", events.Events)

			pvc, err := getPodPvc(namespace, pod.ObjectMeta.Name)
			if err != nil {
				log.Printf("Error getting pvc for pod %s in namespace %s. Error: %v", pod.ObjectMeta.Name, namespace, err)
			}
			log.Printf("Pod pvc gotten: %v", pvc.Items)

			// TODO: Implement pod taint
			// problem: not directly available via pod API. Maybe use Node api and try to match?
			addPodDetailPage(pdf, podDetailPageId, pod.ObjectMeta.Name, labels, "tbi", pod.ContainerImages, formatSimplePvcList(pvc.Items), pod.NodeName, formatEventListArray(events))
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
			conditions := nodeInfo.Conditions
			var (
				networkUnavailable string
				memoryPressure     string
				diskPressure       string
				pidPressure        string
				ready              string
			)
			for _, item := range conditions {
				switch item.Type {
				case "MemoryPressure":
					memoryPressure = string(item.Status)
				case "DiskPressure":
					diskPressure = string(item.Status)
				case "PIDPressure":
					pidPressure = string(item.Status)
				case "Ready":
					ready = string(item.Status)
				}
			}
			networkUnavailable = "tbi"
			// TODO: Implement NetworkUnavailable
			// problem: how to get?
			addNodePage(pdf, nodePageId, node.ObjectMeta.Name, labels, taints, nodeInfo.NodeInfo.OSImage, internalIps, !nodeInfo.Unschedulable, networkUnavailable, memoryPressure, diskPressure, pidPressure, ready, events)
		}
	}

	pvcList, err := getPvcDetail(namespace)
	if err != nil {
		log.Printf("Error getting pvc list. Skipping. Error: %v", err)
	} else {
		for _, pvc := range pvcList.Items {
			log.Printf("pvc detail gotten, %v", pvc)
			labels := formatLabelString(pvc.ObjectMeta.Labels)
			// TODO: Implement PVC events
			// Problem: not available to API
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
		false, "false", "false", "false", "false", "True", []string{"EVENT1", "EVENT2", "EVENT3"})

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
func addNodePage(pdf *gofpdf.Fpdf, nodePageId int, nodeName, labels, taints, osimage, ip string, schedulable bool, networkunavailable, memorypressure, diskpressure, pidpressure, ready string, events []string) {
	pdf.AddPage()
	importer.UseImportedTemplate(pdf, nodePageId, 0, 0, reportWidth, reportHeight)
	addText(pdf, "node.name", nodeName)
	addText(pdf, "node.labels", labels)
	addText(pdf, "node.taints", taints)
	addText(pdf, "node.osimage", osimage)
	addText(pdf, "node.ip", ip)
	addText(pdf, "node.schedulable", strconv.FormatBool(schedulable))
	addText(pdf, "node.state.networkunavailable", networkunavailable)
	addText(pdf, "node.state.memorypressure", memorypressure)
	addText(pdf, "node.state.diskpressure", diskpressure)
	addText(pdf, "node.state.pidpressure", pidpressure)
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
