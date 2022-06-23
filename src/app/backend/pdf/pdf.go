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
	"title.generated":      {30, 60},
	"title.namespace":      {30, 80},
	"poddetail.name":       {60, 33},
	"poddetail.labels":     {30, 50},
	"poddetail.taints":     {30, 70},
	"poddetail.containers": {30, 90},
	"poddetail.pvc":        {30, 130},
	"poddetail.nodes":      {30, 150},
	"poddetail.events":     {30, 170},
	"podlogs.name":         {60, 33},
	"podlogs.logs":         {30, 50},
}

const reportHeight float64 = 297
const reportWidth float64 = 210

func GenerateTestPdf() error {
	pdf := gofpdf.New(gofpdf.OrientationPortrait, "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Helvetica", "B", 16)
	pdf.Cell(20, 20, "Hello World!")
	err := pdf.OutputFileAndClose("/tmp/test.pdf")
	return err
}

func GenerateTemplatePdf() error {
	pdf := gofpdf.New(gofpdf.OrientationPortrait, "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Helvetica", "", 16)
	title_page := gofpdi.ImportPage(pdf, "src/app/backend/pdf/templates/title_page.pdf", 1, "/MediaBox")
	gofpdi.UseImportedTemplate(pdf, title_page, 0, 0, 210, 297)

	pdf.Text(50, 50, "I added this text here!")

	err := pdf.OutputFileAndClose("/tmp/test2.pdf")
	return err
}

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
	addText(pdf, "poddetail.containers", "CONTAINER1\r\nCONTAINER2\r\nCONTAINER3")
	addText(pdf, "poddetail.pvc", "SAMPLE PVC")
	addText(pdf, "poddetail.nodes", "NODE1, NODE2, NODE3, NODE4, NODE5")
	addText(pdf, "poddetail.events", "EVENT1\r\nEVENT2\r\nEVENT3")

	pdf.AddPage()
	podLogsPage := gofpdi.ImportPage(pdf, "templates/pod_logs.pdf", 1, "/MediaBox")
	gofpdi.UseImportedTemplate(pdf, podLogsPage, 0, 0, reportWidth, reportHeight)
	addText(pdf, "podlogs.name", "SAMPLE-POD-ABCDEF1234567890")
	addText(pdf, "podlogs.logs", "LOG1\nLOG2\nLOG3")

	err := pdf.OutputFileAndClose("/tmp/Report-" + reportGenerated + ".pdf")
	return err
}

func addText(pdf *gofpdf.Fpdf, key string, text string) {
	location := pointMap[key]
	pdf.Text(location.x, location.y, text)
}
