package puml

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"text/template"

	"github.com/koderizer/arc/model"
)

//C4Context type hold all data structure to render Context diagrams
type C4Context struct {
	Title     string
	Arc       model.ArcType
	Relations []C4Relation
}

//C4SystemContainer type hold data structure to render Container diagrams
type C4SystemContainer struct {
	Title     string
	Systems   map[string][]model.Container
	Users     []model.User
	Relations []C4Relation
	Neighbors []model.ExternalSystem
}

//C4Neighbor is the generic presentation for any partnering elements
type C4Neighbor struct {
	Name string
	Desc string
}

//C4Relation is the data struct to draw relation in C4
type C4Relation struct {
	Subject     string
	Object      string
	Pointer     string
	PointerTech string
}

//C4ContextPuml generate puml code for Context diagram using the given ArcType data
func C4ContextPuml(arcData model.ArcType, targets ...string) (string, error) {
	if arcData.App == "" || arcData.Desc == "" {
		return "", errors.New("Context require Application name and description")
	}
	funcMap := template.FuncMap{
		"CleanUp": cleanUp,
		"CleanID": cleanID,
	}
	contextTemplate, err := template.New("c4ContextTemplate").Funcs(funcMap).Parse(c4ContextTemplate)
	if err != nil {
		log.Println("Fail to parse tpl file")
		return "", err
	}
	data, err := c4ContextParse(arcData, targets...)
	if err != nil {
		log.Println(err)
		return "", err
	}
	puml := []byte{}
	wr := bytes.NewBuffer(puml)

	if err = contextTemplate.ExecuteTemplate(wr, "c4ContextTemplate", data); err != nil {
		return "", err
	}

	return wr.String(), nil
}

//c4ContextParse prepare the data structure to render C4 full Landscape or targeted Context diagram
func c4ContextParse(arcData model.ArcType, targets ...string) (C4Context, error) {
	// sys := relMap(arcData, targets...)
	relations := make([]C4Relation, 0)
	for _, relation := range arcData.Relations {

		relations = append(relations, C4Relation{
			Subject:     relation.Subject,
			Pointer:     cleanRelation(relation.Pointer),
			PointerTech: parseRelationTech(relation.Pointer),
			Object:      relation.Object,
		})

	}
	var title string
	if len(targets) != 0 && len(targets) != len(arcData.InternalSystems) {
		title = fmt.Sprintf("System Context view for: %s", strings.Join(targets, ", "))
	} else {
		title = fmt.Sprintf("System Landscape view for: %s", arcData.App)
	}
	return C4Context{
		Title:     title,
		Arc:       arcData,
		Relations: relations,
	}, nil
}

//C4ContainerPuml generate the C4 plantUml code from ArcType data to draw Container diagram for target Systems
func C4ContainerPuml(arcData model.ArcType, targets ...string) (string, error) {

	funcMap := template.FuncMap{
		"CleanUp": cleanUp,
		"CleanID": cleanID,
	}
	containerTemplate, err := template.New("c4ContainerTemplate").Funcs(funcMap).Parse(c4ContainerTemplate)
	if err != nil {
		log.Println("Fail to parse tpl")
		return "", err
	}
	data, err := c4ContainerParse(arcData)
	if err != nil {
		log.Println(err)
		return "", err
	}
	data.Title = fmt.Sprintf("System Container view for: %s", strings.Join(targets, ", "))
	puml := []byte{}
	wr := bytes.NewBuffer(puml)

	if err = containerTemplate.ExecuteTemplate(wr, "c4ContainerTemplate", data); err != nil {
		return "", err
	}

	return wr.String(), nil
}

//c4ContainerParse return the data to render Container diagram for given target system and clip out all others.
func c4ContainerParse(arcData model.ArcType) (C4SystemContainer, error) {

	rels := make([]C4Relation, 0)
	sys := make(map[string][]model.Container, 0)
	for _, s := range arcData.InternalSystems {
		sys[s.Name] = s.Containers
	}
	for _, r := range arcData.Relations {
		rels = append(rels, C4Relation{
			Subject:     r.Subject,
			Object:      r.Object,
			Pointer:     cleanRelation(r.Pointer),
			PointerTech: parseRelationTech(r.Pointer),
		})
	}

	return C4SystemContainer{
		Systems:   sys,
		Users:     arcData.Users,
		Relations: rels,
		Neighbors: arcData.ExternalSystems,
	}, nil
}

/* Map up all primary top path between 2 systems
by folding all relations into parent targeted systems */
func relMap(arcData model.ArcType, targets ...string) map[string][]string {
	sys := make(map[string][]string)
	tmap := make(map[string]int)
	if len(targets) > 0 {
		for _, k := range targets {
			tmap[k] = 1
		}
	}
	for _, r := range arcData.Relations {
		s := strings.Split(r.Subject, ".")
		o := strings.Split(r.Object, ".")
		if len(targets) > 0 {
			if _, oks := tmap[s[0]]; !oks {
				if _, oko := tmap[o[0]]; !oko {
					continue
				}
			}
		}
		key := fmt.Sprintf("%s+%s", s[0], o[0])
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
	return sys
}

//Utilities for C4 visualization parsing
func parseRelationTech(rel string) string {
	r := regexp.MustCompile(`\((.*?)\)`)
	matches := r.FindSubmatch([]byte(rel))
	if matches == nil {
		return ""
	}
	return string(matches[1])
}

//Utilities function for template map
func cleanRelation(rel string) string {
	r := regexp.MustCompile(`\((.*?)\)`)
	return r.ReplaceAllString(rel, "")
}

func cleanUp(s string) string {
	return strings.ReplaceAll(s, "\n", " ")
}
func cleanID(s string) string {
	return strings.ReplaceAll(s, "-", "")
}
