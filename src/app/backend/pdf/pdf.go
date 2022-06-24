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
	"node.state.networkunavailable": {90, 140},
	"node.state.memorypressure":     {90, 150},
	"node.state.diskpressure":       {90, 160},
	"node.state.pidpressure":        {90, 170},
	"node.state.ready":              {90, 180},
	"node.events":                   {30, 195},
}

const reportHeight float64 = 297
const reportWidth float64 = 210

func GenerateTestReport() error {
	pdf := gofpdf.New(gofpdf.OrientationPortrait, "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Helvetica", "", 14)

	titlePage := gofpdi.ImportPage(pdf, "templates/title_page.pdf", 1, "/MediaBox")
	gofpdi.UseImportedTemplate(pdf, titlePage, 0, 0, reportWidth, reportHeight)
	reportGenerated := time.Now().Format("01-02-2006_15-04-05")
	addText(pdf, "title.generated", reportGenerated)
	addText(pdf, "title.namespace", "SAMPLE-NAMESPACE")

	pdf.AddPage()
	podDetailPage := gofpdi.ImportPage(pdf, "templates/pod_detail.pdf", 1, "/MediaBox")
	gofpdi.UseImportedTemplate(pdf, podDetailPage, 0, 0, reportWidth, reportHeight)
	addText(pdf, "poddetail.name", "SAMPLE-POD-ABCDEF1234567890")
	addText(pdf, "poddetail.labels", "LABEL1, LABEL2, LABEL3, LABEL4, LABEL5")
	addText(pdf, "poddetail.taints", "SAMPLE.SAMPLE/SAMPLE:SAMPLE op=Exists for 300s")
	addText(pdf, "poddetail.containers", "CONTAINER1", "CONTAINER2", "CONTAINER3")
	addText(pdf, "poddetail.pvc", "SAMPLE PVC")
	addText(pdf, "poddetail.nodes", "NODE1, NODE2, NODE3, NODE4, NODE5")
	addText(pdf, "poddetail.events", "EVENT1", "EVENT2", "EVENT3")

	pdf.AddPage()
	podLogsPage := gofpdi.ImportPage(pdf, "templates/pod_logs.pdf", 1, "/MediaBox")
	gofpdi.UseImportedTemplate(pdf, podLogsPage, 0, 0, reportWidth, reportHeight)
	addText(pdf, "podlogs.name", "SAMPLE-POD-ABCDEF1234567890")
	addText(pdf, "podlogs.logs", "LOG1", "LOG2", "LOG3", "LOG4", "LOG5")

	pdf.AddPage()
	nodePage := gofpdi.ImportPage(pdf, "templates/node_detail.pdf", 1, "/MediaBox")
	gofpdi.UseImportedTemplate(pdf, nodePage, 0, 0, reportWidth, reportHeight)
	addText(pdf, "node.name", "SAMPLE-NODE1")
	addText(pdf, "node.labels", "LABEL1, LABEL2, LABEL3")
	addText(pdf, "node.taints", "SAMPLE.SAMPLE/SAMPLE:SAMPLE op=Exists for 300s")
	addText(pdf, "node.osimage", "Linux Shminux 22.04 LTS")
	addText(pdf, "node.ip", "123.123.123.123")
	addText(pdf, "node.schedulable", "Truefalse")
	addText(pdf, "node.state.networkunavailable", "Truefalse")
	addText(pdf, "node.state.memorypressure", "Truefalse")
	addText(pdf, "node.state.diskpressure", "Truefalse")
	addText(pdf, "node.state.pidpressure", "Truefalse")
	addText(pdf, "node.state.ready", "Truefalse")
	addText(pdf, "node.events", "EVENT1", "EVENT2", "EVENT3")

	pdf.AddPage()
	pvcPage := gofpdi.ImportPage(pdf, "templates/pvc_detail.pdf", 1, "/MediaBox")
	gofpdi.UseImportedTemplate(pdf, pvcPage, 0, 0, reportWidth, reportHeight)
	addText(pdf, "pvc.name", "SAMPLE-PVC-1")
	addText(pdf, "pvc.state", "bound")
	addText(pdf, "pvc.storageclass", "local-storage")
	addText(pdf, "pvc.volume", "SAMPLE-PV-1")
	addText(pdf, "pvc.labels", "LABEL1, LABEL2, LABEL3, LABEL4")
	addText(pdf, "pvc.capacity", "100Gi")
	addText(pdf, "pvc.events", "EVENT1", "EVENT2", "EVENT3")

	err := pdf.OutputFileAndClose("/tmp/Report-" + reportGenerated + ".pdf")
	return err
}

func addText(pdf *gofpdf.Fpdf, key string, text ...string) {
	location := pointMap[key]
	for i, item := range text {
		pdf.Text(location.x, location.y+float64(5*i), item)
	}
}
