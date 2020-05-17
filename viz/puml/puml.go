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
	SystemName string
	Users      []model.User
	Containers []model.Container
	Relations  []C4Relation
	Neighbors  []C4Neighbor
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
	sys := relMap(arcData, targets...)
	relations := make([]C4Relation, 0)
	var arc model.ArcType
	arc.InternalSystems = make([]model.InternalSystem, 0)
	arc.ExternalSystems = make([]model.ExternalSystem, 0)
	arc.Users = make([]model.User, 0)
	ismap := make(map[string]model.InternalSystem, 0)
	esmap := make(map[string]model.ExternalSystem, 0)
	usmap := make(map[string]model.User)
	if len(targets) == 0 {
		arc = arcData
	} else {
		for _, t := range arcData.InternalSystems {
			ismap[t.Name] = t
		}
		for _, t := range arcData.ExternalSystems {
			esmap[t.Name] = t
		}
		for _, t := range arcData.Users {
			usmap[t.Name] = t
		}
		arc.App = arcData.App
		arc.Desc = arcData.Desc
	}

	for k, v := range sys {
		so := strings.Split(k, "+")
		if len(targets) > 0 {
			for _, k := range so {
				if e, ok := esmap[k]; ok {
					arc.ExternalSystems = append(arc.ExternalSystems, e)
				}
				if u, ok := usmap[k]; ok {
					arc.Users = append(arc.Users, u)
				}
				if i, ok := ismap[k]; ok {
					arc.InternalSystems = append(arc.InternalSystems, i)
				}
			}
		}
		//Skip self-referencing
		if so[0] == so[1] {
			continue
		}
		relations = append(relations, C4Relation{
			Subject:     so[0],
			Pointer:     cleanRelation(strings.Join(v, ",")),
			PointerTech: parseRelationTech(strings.Join(v, ",")),
			Object:      so[1],
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
		Arc:       arc,
		Relations: relations,
	}, nil
}

//C4ContainerPuml generate the C4 plantUml code from ArcType data to draw Container diagram for target Systems
func C4ContainerPuml(arcData model.ArcType, targets ...string) (string, error) {
	if len(targets) != 1 {
		return "", errors.New("Have not support all targets or multi target yet")
	}
	var system *model.InternalSystem
	for _, s := range arcData.InternalSystems {
		if targets[0] == s.Name {
			system = &s
			break
		}
	}
	if system == nil {
		return "", errors.New("Target system not found")
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
	data, err := c4ContainerParse(arcData, system)
	if err != nil {
		log.Println(err)
		return "", err
	}
	puml := []byte{}
	wr := bytes.NewBuffer(puml)

	if err = containerTemplate.ExecuteTemplate(wr, "c4ContainerTemplate", data); err != nil {
		return "", err
	}

	return wr.String(), nil
}

//c4ContainerParse return the data to render Container diagram for given target system and clip out all others.
func c4ContainerParse(arcData model.ArcType, target *model.InternalSystem) (C4SystemContainer, error) {
	if target == nil {
		return C4SystemContainer{}, errors.New("nil target system pointer")
	}
	/*
		Get all containers in targeted system
		Map the unique list of path to and from the containers
		Clip all neighbor elements into their system level
	*/
	relMap := make(map[string]C4Relation, 0)
	conMap := make(map[string]bool)
	neiMap := make(map[string]bool)
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
	rels := make([]C4Relation, 0)
	users := make([]model.User, 0)
	for _, v := range relMap {
		rels = append(rels, v)
	}
	neis := make([]C4Neighbor, 0)
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
			neis = append(neis, C4Neighbor{
				Name: k,
				Desc: desc,
			})
		}
	}
	return C4SystemContainer{
		SystemName: target.Name,
		Users:      users,
		Containers: target.Containers,
		Relations:  rels,
		Neighbors:  neis,
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
