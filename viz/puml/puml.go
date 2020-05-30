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
	Neighbors []C4Neighbor
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

	systems := make(map[string]*model.InternalSystem, 0)
	for _, s := range arcData.InternalSystems {
		log.Println("look for:", s.Name)
		if found := strings.Contains(strings.Join(targets, "+"), s.Name); found {
			systems[s.Name] = &s
		}
	}

	funcMap := template.FuncMap{
		"CleanUp": cleanUp,
		"CleanID": cleanID,
	}
	containerTemplate, err := template.New("c4ContainerTemplate").Funcs(funcMap).Parse(c4ContainerTemplate)
	if err != nil {
		log.Println("Fail to parse tpl")
		return "", err
	}
	data, err := c4ContainerParse(arcData, systems)
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
func c4ContainerParse(arcData model.ArcType, targets map[string]*model.InternalSystem) (C4SystemContainer, error) {
	systems := make(map[string][]model.Container, 0)

	relMap := make(map[string]C4Relation, 0)
	conMap := make(map[string]bool)
	neiMap := make(map[string]bool)
	rels := make([]C4Relation, 0)
	users := make([]model.User, 0)
	neis := make([]C4Neighbor, 0)
	for _, target := range targets {

		if target == nil {
			return C4SystemContainer{}, errors.New("nil target system pointer")
		}
		/*
			Get all containers in targeted system
			Map the unique list of path to and from the containers
			Clip all neighbor elements into their system level
		*/
		for _, c := range target.Containers {
			cid := target.Name + "." + c.Name
			if _, ok := conMap[cid]; !ok {
				conMap[cid] = true
			}
		}
		for _, rel := range arcData.Relations {
			_, isO := conMap[rel.Object]
			_, isS := conMap[rel.Subject]
			if !isO && !isS {
				continue
			}

			var nIDSys string
			subjectChain := strings.Split(rel.Subject, ".")
			subjectClip := subjectChain[0]
			objectChain := strings.Split(rel.Object, ".")
			objectClip := objectChain[0]
			if _, ok := conMap[rel.Subject]; !ok {
				nIDSys = subjectClip
				if len(objectChain) >= 2 {
					objectClip = strings.Join(objectChain[0:2], ".")
				}
			} else {
				if _, ok := conMap[rel.Object]; !ok {
					nIDSys = objectClip
				} else {
					if len(objectChain) >= 2 {
						objectClip = strings.Join(objectChain[0:2], ".")
					}
				}
				if len(subjectChain) >= 2 {
					subjectClip = strings.Join(subjectChain[0:2], ".")
				}
			}

			if nIDSys != target.Name {
				if _, ok := neiMap[nIDSys]; !ok {
					neiMap[nIDSys] = true
				}
			}

			relID := subjectClip + "->" + objectClip
			if _, ok := relMap[relID]; !ok {
				relMap[relID] = C4Relation{
					Object:      objectClip,
					Subject:     subjectClip,
					Pointer:     cleanRelation(rel.Pointer),
					PointerTech: parseRelationTech(rel.Pointer),
				}
			}
		}
		systems[target.Name] = target.Containers

		for _, v := range relMap {
			rels = append(rels, v)
		}
		for k := range neiMap {
			var desc string
			for _, is := range arcData.InternalSystems {
				if k == is.Name {
					desc = is.Desc
					goto found
				}
			}
			for _, es := range arcData.ExternalSystems {
				if k == es.Name {
					desc = es.Desc
					goto found
				}
			}
			for _, us := range arcData.Users {
				if k == us.Name {
					users = append(users, us)
				}
			}
		found:
			if desc != "" {
				if _, ok := targets[k]; !ok {
					neis = append(neis, C4Neighbor{
						Name: k,
						Desc: desc,
					})
				}
			}
		}
	}
	return C4SystemContainer{
		Systems:   systems,
		Users:     users,
		Relations: rels,
		Neighbors: neis,
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
