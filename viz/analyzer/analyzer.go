//Package analyzer provide utilities to analyze a render requests
package analyzer

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/koderizer/arc/model"
	"github.com/yourbasic/graph"
)

//Perspective type
const (
	Landscape                = 0
	Context                  = 1
	Container                = 2
	Component                = 3
	Code                     = 4
	DefaultDependencyPointer = "Use:"
)

//Perspective are type supported by viz
type Perspective int

//Graph data type hold all information to render the architecture info
type Graph struct {
	Pers     Perspective
	Type     string
	Arc      *model.ArcType
	tars     []string
	tarMap   map[string]int
	graph    *graph.Mutable
	vids     map[string]int
	eids     map[string]int64
	edges    map[int64]edge
	vertices map[int]Vertice
}

//VerticeType constants
const (
	VerticeTypeUser           = 0
	VerticeTypeInternalSystem = 1
	VerticeTypeExternalSystem = 2
	VerticeTypeContainer      = 3
	VerticeTypeComponent      = 4
)

//VerticeType map to the model abstraction
type VerticeType int

//Vertice type
type Vertice struct {
	Entity interface{}
	Kind   VerticeType
}

type edge struct {
	relation model.Relation
	views    map[Perspective]bool
}

//GetUsers return all users that is concern
// func (g *Graph) GetUsers() ([]model.User, error) {
// 	if g.Arc == nil {
// 		return nil, errors.New("Empty graph")
// 	}
// 	users := make([]model.User, 0)
// 	if len(g.tars) > 0 {
// 		for _, tar := range g.tars {
// 			vid, _ := g.vids[tar]
// 			for _, u := range g.Arc.Users {
// 				uid, _ := g.vids[u.Name]
// 				if g.graph.Edge(vid, uid) {
// 					users = append(users, u)
// 				}
// 			}
// 		}
// 		return users, nil
// 	}
// 	return g.Arc.Users, nil
// }

//GetInternalSystems return relevant internal systems
// func (g *Graph) GetInternalSystems() ([]model.InternalSystem, error) {
// 	if g.Arc == nil {
// 		return nil, errors.New("Empty graph")
// 	}
// 	lup := make(map[int]model.InternalSystem, 0)
// 	systems := make([]model.InternalSystem, 0)
// 	if len(g.tars) > 0 {
// 		for _, tar := range g.tars {
// 			vid, _ := g.vids[tar]
// 			for _, s := range g.Arc.InternalSystems {
// 				sid, _ := g.vids[s.Name]
// 				if sid == vid {
// 					if _, ok := lup[sid]; !ok {
// 						lup[sid] = s
// 					}
// 					continue
// 				}
// 				if _, ok := lup[sid]; !ok {
// 					if g.graph.Edge(vid, sid) {
// 						lup[sid] = s
// 					}
// 				}
// 			}
// 		}
// 		for _, s := range lup {
// 			systems = append(systems, s)
// 		}
// 		return systems, nil
// 	}
// 	return g.Arc.InternalSystems, nil
// }

// GetUsers return relevant internal systems
func (g *Graph) GetUsers() ([]model.User, error) {
	if g.Arc == nil {
		return nil, errors.New("Empty graph")
	}
	users := make([]model.User, 0)
	if len(g.tars) > 0 {
		for _, tid := range g.tarMap {
			if g.Pers == Context || g.Pers == Landscape {
				for _, vid := range g.walkTarget(tid, VerticeTypeUser) {
					users = append(users, g.vertices[vid].Entity.(model.User))
				}
			}
		}
		return users, nil
	}
	return g.Arc.Users, nil
}

//GetInternalSystems return relevant internal systems
func (g *Graph) GetInternalSystems() ([]model.InternalSystem, error) {
	if g.Arc == nil {
		return nil, errors.New("Empty graph")
	}
	systems := make([]model.InternalSystem, 0)
	if len(g.tars) > 0 {
		for _, tid := range g.tarMap {
			systems = append(systems, g.vertices[tid].Entity.(model.InternalSystem))
			// for _, vid := range g.walkTarget(tid, VerticeTypeInternalSystem) {
			// 	systems = append(systems, g.vertices[vid].Entity.(model.InternalSystem))
			// }
			// for _, container := range g.vertices[tid].Entity.(model.InternalSystem).Containers {
			// 	for _, vid := range g.walkTarget(g.vids[tar+"."+container.Name], VerticeTypeInternalSystem) {
			// 		systems = append(systems, g.vertices[vid].Entity.(model.InternalSystem))
			// 	}
			// }
		}
		return systems, nil
	}
	return g.Arc.InternalSystems, nil
}

//GetExternalSystems return relevant external systems or internal systems if the view is targeted
func (g *Graph) GetExternalSystems() ([]model.ExternalSystem, error) {
	if g.Arc == nil {
		return nil, errors.New("Empty graph")
	}
	systems := make([]model.ExternalSystem, 0)
	if len(g.tars) > 0 {
		for tar, tid := range g.tarMap {
			if g.Pers == Context || g.Pers == Landscape {
				for _, vid := range g.walkTarget(tid, VerticeTypeExternalSystem) {
					systems = append(systems, g.vertices[vid].Entity.(model.ExternalSystem))
				}
			}
			for _, container := range g.vertices[tid].Entity.(model.InternalSystem).Containers {
				for _, vid := range g.walkTarget(g.vids[tar+"."+container.Name], VerticeTypeExternalSystem) {
					systems = append(systems, g.vertices[vid].Entity.(model.ExternalSystem))
				}
				for _, vid := range g.walkTarget(g.vids[tar+"."+container.Name], VerticeTypeInternalSystem) {
					internalExtern := g.vertices[vid].Entity.(model.InternalSystem)
					systems = append(systems, model.ExternalSystem{
						Name: internalExtern.Name,
						Desc: internalExtern.Desc,
					})
				}
			}
		}
		return systems, nil
	}
	return g.Arc.ExternalSystems, nil
}

func (g *Graph) walkTarget(vid int, kind VerticeType) []int {
	results := make([]int, 0)
	g.graph.Visit(vid, func(w int, c int64) bool {
		if g.vertices[w].Kind == kind {
			results = append(results, w)
		}
		return false
	})
	return results
}

//GetRelations return the list of relevant relations
func (g *Graph) GetRelations() ([]model.Relation, error) {
	if g.Arc == nil {
		return nil, errors.New("Empty graph")
	}
	relations := make([]model.Relation, 0)
	relationIDs := make(map[int64]int, 0)
	if len(g.tars) > 0 {
		for _, vid := range g.tarMap {
			g.graph.Visit(vid, func(w int, c int64) bool {
				relationIDs[c] = w
				return false
			})
			sys := g.vertices[vid].Entity.(model.InternalSystem)
			for _, container := range sys.Containers {
				vid, _ := g.vids[sys.Name+"."+container.Name]
				g.graph.Visit(vid, func(w int, c int64) bool {
					relationIDs[c] = w
					return false
				})
			}
		}
		for eid := range relationIDs {
			if show, ok := g.edges[eid].views[g.Pers]; ok && show {
				relations = append(relations, g.edges[eid].relation)
			}
		}
	} else {
		for _, edge := range g.edges {
			if show, ok := edge.views[g.Pers]; ok && show {
				relations = append(relations, edge.relation)
			}
		}
	}
	return relations, nil
}

//Init the graph will generate a list of local ids and return total number of nodes
func (g *Graph) Init() int {
	if g.Arc == nil {
		log.Println("Initialized a empty graph!")
		return 0
	}
	g.tarMap = make(map[string]int, 0)
	g.vids = make(map[string]int, 0)
	g.vertices = make(map[int]Vertice, 0)
	if len(g.tars) > 0 {
		for _, tar := range g.tars {
			g.tarMap[tar] = 0
		}
	}
	//Form a local list of ids by simply iterate and incrementally index
	for _, user := range g.Arc.Users {
		if _, ok := g.vids[user.Name]; !ok {
			vid := len(g.vids) + 1
			g.vids[user.Name] = vid
			g.vertices[vid] = Vertice{
				Entity: user,
				Kind:   VerticeTypeUser,
			}
		}
	}
	for _, isys := range g.Arc.InternalSystems {
		if _, ok := g.vids[isys.Name]; !ok {
			vid := len(g.vids) + 1
			g.vids[isys.Name] = vid
			g.vertices[vid] = Vertice{
				Entity: isys,
				Kind:   VerticeTypeInternalSystem,
			}
			if _, found := g.tarMap[isys.Name]; len(g.tars) > 0 && found {
				g.tarMap[isys.Name] = vid
			}
		}
		for _, container := range isys.Containers {
			cname := fmt.Sprintf("%s.%s", isys.Name, container.Name)
			if _, ok := g.vids[cname]; !ok {
				vid := len(g.vids) + 1
				g.vids[cname] = vid
				g.vertices[vid] = Vertice{
					Entity: container,
					Kind:   VerticeTypeContainer,
				}
			}
			for _, component := range container.Components {
				comName := fmt.Sprintf("%s.%s", cname, component.Name)
				if _, ok := g.vids[comName]; !ok {
					vid := len(g.vids) + 1
					g.vids[comName] = vid
					g.vertices[vid] = Vertice{
						Entity: component,
						Kind:   VerticeTypeComponent,
					}
				}
			}
		}
	}
	for _, esys := range g.Arc.ExternalSystems {
		if _, ok := g.vids[esys.Name]; !ok {
			vid := len(g.vids) + 1
			g.vids[esys.Name] = vid
			g.vertices[vid] = Vertice{
				Entity: esys,
				Kind:   VerticeTypeExternalSystem,
			}
		}
	}
	g.graph = graph.New(len(g.vids) + 1)
	g.eids = make(map[string]int64, 0)
	g.edges = make(map[int64]edge, 0)
	return len(g.vids)
}

//Analyse attempt to form a graph that is relevant to the render targets
func (g *Graph) Analyse() error {
	if g.graph == nil {
		return errors.New("Empty or un-initialized graph")
	}
	for _, relation := range g.Arc.Relations {
		subjectChain := strings.Split(relation.Subject, ".")
		objectChain := strings.Split(relation.Object, ".")
		sid, ok := g.vids[relation.Subject]
		if !ok {
			return errors.New("Invalid Subject id found in relation")
		}
		oid, ok := g.vids[relation.Object]
		if !ok {
			return errors.New("Invalid Object id found in relation")
		}
		ename := fmt.Sprintf("%s&%s", relation.Subject, relation.Object)

		//Decide which views this path should be shown
		views := make(map[Perspective]bool, 0)
		switch len(subjectChain) + len(objectChain) {
		case 2:
			views[Landscape] = true
			views[Context] = true
		case 3:
		case 4:
			views[Container] = true
		default:
			views[Component] = true
		}

		id, ok := g.eids[ename]
		if !ok {
			edgeID := int64(len(g.eids) + 1)
			g.eids[ename] = edgeID
			g.edges[edgeID] = edge{relation, views}
			g.graph.AddBothCost(sid, oid, g.eids[ename])
		} else {
			if strings.Contains(g.edges[id].relation.Pointer, DefaultDependencyPointer) {
				g.edges[id] = edge{relation, views}
			}
		}

		//Add parent dependency if not exists

		if len(subjectChain) > 1 || len(objectChain) > 1 {
			parentSubjectID := subjectChain[0]
			parentObjectID := objectChain[0]
			if parentObjectID == parentSubjectID {
				continue
			}
			parentEname := fmt.Sprintf("%s&%s", parentSubjectID, parentObjectID)
			_, ok := g.eids[parentEname]
			if !ok {
				pEdgeID := int64(len(g.eids) + 1)
				g.eids[parentEname] = pEdgeID
				v := make(map[Perspective]bool, 2)
				v[Context] = true
				v[Landscape] = true
				g.edges[pEdgeID] = edge{
					relation: model.Relation{
						Subject: parentSubjectID,
						Object:  parentObjectID,
						Pointer: fmt.Sprintf("%s:%s", DefaultDependencyPointer, relation.Pointer)},
					views: v,
				}
				g.graph.AddBothCost(g.vids[parentSubjectID], g.vids[parentObjectID], pEdgeID)
			}
		}
	}
	return nil
}

//Process the render request to build a Graph to visualize
func Process(ctx context.Context, req *model.RenderRequest) (*Graph, error) {
	res := &Graph{}
	switch req.GetPerspective() {
	case model.PresentationPerspective_LANDSCAPE:
		res.Pers = Landscape
	case model.PresentationPerspective_CONTEXT:
		res.Pers = Context
	case model.PresentationPerspective_CONTAINER:
		res.Pers = Container
	case model.PresentationPerspective_COMPONENT:
		res.Pers = Component
	case model.PresentationPerspective_CODE:
		res.Pers = Code
	default:
		return nil, errors.New("Invalid perspective")
	}
	switch req.GetDataFormat() {
	case model.ArcDataFormat_ARC:
		dec := gob.NewDecoder(bytes.NewBuffer(req.GetData()))
		if err := dec.Decode(&res.Arc); err != nil {
			log.Printf("Fail to decode data: %v", err)
			return nil, err
		}
	case model.ArcDataFormat_JSON:
	case model.ArcDataFormat_PUML:
		return nil, errors.New("Json and Puml direct render request are not supported for now")
	default:
		return nil, errors.New("Unsupported data format")
	}

	res.tars = req.GetTarget()

	switch req.GetVisualFormat() {
	case model.ArcVisualFormat_SVG:
		res.Type = "svg"
	case model.ArcVisualFormat_PNG:
		res.Type = "png"
	case model.ArcVisualFormat_PDF:
		return nil, errors.New("PDF is not supported for now")
	default:
		return nil, errors.New("Unsupported visual output type")
	}

	if res.Init() == 0 {
		return res, errors.New("Empty element")
	}

	if err := res.Analyse(); err != nil {
		return res, err
	}

	return res, nil
}
