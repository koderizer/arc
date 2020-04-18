package puml

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"

	"github.com/koderizer/arc/model"
)

const c4contextTemplate = "./templates/C4Context.puml.template"
const c4contextFile = "./C4Context.puml"

//C4Model type hold all data structure to render different diagrams
type C4Model struct {
	Title string
	Arc   model.ArcType
}

//C4ContextPuml generate puml code for Context diagram using the given ArcType data
func C4ContextPuml(arcData model.ArcType) (string, error) {
	if arcData.App == "" || arcData.Desc == "" {
		return "", errors.New("Context require Application name and description")
	}
	contextTemplate, err := template.ParseFiles(c4contextTemplate)
	if err != nil {
		return "", err
	}
	data := C4Model{Title: fmt.Sprintf("System Context Diagram for %s", arcData.App), Arc: arcData}
	puml := []byte{}
	wr := bytes.NewBuffer(puml)

	if err = contextTemplate.ExecuteTemplate(wr, contextTemplate.Name(), data); err != nil {
		return "", err
	}

	return wr.String(), nil
}
