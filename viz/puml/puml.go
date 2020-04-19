package puml

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strings"
	"text/template"

	"github.com/koderizer/arc/model"
)

const c4contextTemplate = "./templates/C4Context.puml.tpl"
const c4contextFile = "./C4Context.puml"

//C4Model type hold all data structure to render different diagrams
type C4Model struct {
	Title string
	Arc   model.ArcType
}

func cleanUp(s string) string {
	return strings.ReplaceAll(s, "\n", " ")
}
func cleanID(s string) string {
	return strings.ReplaceAll(s, "-", "")
}

//C4ContextPuml generate puml code for Context diagram using the given ArcType data
func C4ContextPuml(arcData model.ArcType) (string, error) {
	if arcData.App == "" || arcData.Desc == "" {
		return "", errors.New("Context require Application name and description")
	}
	funcMap := template.FuncMap{
		"CleanUp": cleanUp,
		"CleanID": cleanID,
	}
	contextTemplate, err := template.New(c4contextTemplate).Funcs(funcMap).ParseFiles(c4contextTemplate)
	if err != nil {
		log.Println("Fail to parse tpl file")
		return "", err
	}
	contextTemplate = contextTemplate.Funcs(funcMap)
	data := C4Model{Title: fmt.Sprintf("System Context Diagram for %s", arcData.App), Arc: arcData}
	puml := []byte{}
	wr := bytes.NewBuffer(puml)

	if err = contextTemplate.ExecuteTemplate(wr, "C4Context.puml.tpl", data); err != nil {
		return "", err
	}

	return wr.String(), nil
}
