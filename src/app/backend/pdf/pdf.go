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
	"title.generated": {30, 60},
	"title.namespace": {30, 80},
}

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

func GenerateTestTitlePage() error {
	pdf := gofpdf.New(gofpdf.OrientationPortrait, "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Helvetica", "", 16)
	title_page := gofpdi.ImportPage(pdf, "src/app/backend/pdf/templates/title_page.pdf", 1, "/MediaBox")
	gofpdi.UseImportedTemplate(pdf, title_page, 0, 0, 210, 297)

	reportGenerated := time.Now().Format("01-02-2006_15-04-05")
	addText(pdf, "title.generated", reportGenerated)
	addText(pdf, "title.namespace", "SAMPLE-NAMESPACE")

	err := pdf.OutputFileAndClose("/tmp/Report-" + reportGenerated + ".pdf")
	return err
}

func addText(pdf *gofpdf.Fpdf, key string, text string) {
	location := pointMap[key]
	pdf.Text(location.x, location.y, text)
}
