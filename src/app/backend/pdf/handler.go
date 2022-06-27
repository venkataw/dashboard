package pdf

import (
	"fmt"
	"net/http"
	"os"

	"github.com/emicklei/go-restful/v3"
)

type pdfDetail struct {
	Name string `json:"name"`
}

/*var testPdfs = []pdfDetail{
	{Name: "Report-06-27-2022_12-33-44.pdf"},
	{Name: "notreal.pdf"},
	{Name: "notrealagain.pdf"},
}*/

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

	fmt.Println("pdf api handler initialized.")

	return wsContainer, nil
}
