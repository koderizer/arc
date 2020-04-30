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
	data, err := C4ContextParse(arcData)
	if err != nil {
		log.Println(err)
		return "", err
	}
	puml := []byte{}
	wr := bytes.NewBuffer(puml)

	if err = contextTemplate.ExecuteTemplate(wr, "C4Context.puml.tpl", data); err != nil {
		return "", err
	}

	return wr.String(), nil
}

//C4ContextParse prepare the data structure to render C4 Context diagram
func C4ContextParse(arcData model.ArcType) (C4Model, error) {
	/* Map up all primary top path between 2 systems
	by folding all relations into parent systems */
	sys := make(map[string][]string)

	for _, r := range arcData.Relations {
		s := strings.Split(r.Subject, ".")
		o := strings.Split(r.Object, ".")
		key := fmt.Sprintf("%s-%s", cleanID(s[0]), cleanID(o[0]))
		if _, ok := sys[key]; !ok {
			sys[key] = []string{r.Pointer}
			continue
		}
		if len(s) == 1 && len(o) == 1 {
			if path, ok := sys[key]; ok {
				sys[key] = append(path, "")
				copy(sys[key][1:], sys[key][0:])
				sys[key][0] = r.Pointer
			}
			continue
		}
		sys[key] = append(sys[key], r.Pointer)
	}
	relations := make([]model.Relation, 0)
	for k, v := range sys {
		so := strings.Split(k, "-")
		//Skip self-referencing
		if so[0] == so[1] {
			continue
		}
		relations = append(relations, model.Relation{
			Subject: so[0],
			Pointer: strings.Join(v, ","),
			Object:  so[1],
		})
		log.Println(relations)
	}
	arc := arcData
	arc.Relations = relations
	return C4Model{Title: fmt.Sprintf("System Context Diagram for %s", arcData.App), Arc: arc}, nil
}

//Utilities function for template map
func cleanUp(s string) string {
	return strings.ReplaceAll(s, "\n", " ")
}
func cleanID(s string) string {
	return strings.ReplaceAll(s, "-", "")
}
