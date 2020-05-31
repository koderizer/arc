package analyzer

import (
	"context"
	"log"
	"testing"

	"github.com/koderizer/arc/model"
)

var arc = model.ArcType{
	App:  "test",
	Desc: "Test archictecture",
	Users: []model.User{
		{
			Name: "u1",
			Role: "User 1",
		},
		{
			Name: "u2",
			Role: "User 2",
		},
	},
	InternalSystems: []model.InternalSystem{
		{
			Name: "s1",
			Desc: "System 1",
			Containers: []model.Container{
				{
					Name: "c1",
					Desc: "Container 1",
				},
				{
					Name: "c2",
					Desc: "Container 2",
				},
			},
		},
		{
			Name: "s2",
			Desc: "System 2",
			Containers: []model.Container{
				{
					Name: "c1",
					Desc: "Container 1",
				},
				{
					Name: "c2",
					Desc: "Container 2",
				},
			},
		},
		{
			Name: "s3",
			Desc: "System 3",
			Containers: []model.Container{
				{
					Name: "c1",
					Desc: "Container 1",
				},
				{
					Name: "c2",
					Desc: "Container 2",
				},
			},
		},
	},
	ExternalSystems: []model.ExternalSystem{
		{
			Name: "e1",
			Desc: "Extern System 1",
		},
		{
			Name: "e2",
			Desc: "Extern System 2",
		},
	},
	Relations: []model.Relation{
		{
			Subject: "u1",
			Object:  "s1",
			Pointer: "use",
		},
		{
			Subject: "u2",
			Object:  "s2",
			Pointer: "use",
		},
		{
			Subject: "s1",
			Object:  "s2",
			Pointer: "point to",
		},
		{
			Subject: "s2",
			Object:  "e2",
			Pointer: "point to",
		},
		{
			Subject: "s2",
			Object:  "e1",
			Pointer: "point to",
		},
		{
			Subject: "s1.c1",
			Object:  "s2.c1",
			Pointer: "call",
		},
	},
}

func prepData(per model.PresentationPerspective, targets []string) *model.RenderRequest {
	data, err := arc.Encode()
	if err != nil {
		log.Fatal("Assert fail")
	}
	return &model.RenderRequest{
		DataFormat:   model.ArcDataFormat_ARC,
		VisualFormat: model.ArcVisualFormat_SVG,
		Perspective:  per,
		Data:         data,
		Target:       targets,
	}
}

func prepResp(per int, targets []string) *Graph {
	g := &Graph{
		Type: "svg",
		Arc:  &arc,
		Pers: Perspective(per),
		tars: targets,
	}
	g.Init()
	if err := g.Analyse(); err != nil {
		return nil
	}
	return g
}
func TestProcess(t *testing.T) {
	testProcess := []struct {
		data *model.RenderRequest
		resp *Graph
		err  error
	}{
		{
			data: prepData(model.PresentationPerspective_CONTEXT, []string{"s1"}),
			resp: prepResp(1, []string{"s1"}),
			err:  nil,
		},
	}

	for _, test := range testProcess {
		if test.resp == nil {
			t.Error("Crash")
		}
		if test.resp.Analyse() != test.err {
			t.Error("Test fail")
		}
	}
}

func TestGetUsers(t *testing.T) {
	testUser := []struct {
		data *model.RenderRequest
		resp []model.User
		err  error
	}{
		{
			data: prepData(model.PresentationPerspective_CONTEXT, []string{}),
			resp: arc.Users,
			err:  nil,
		},
		{
			data: prepData(model.PresentationPerspective_CONTEXT, []string{"s1"}),
			resp: arc.Users[:1],
			err:  nil,
		},
		{
			data: prepData(model.PresentationPerspective_CONTEXT, []string{"s2"}),
			resp: arc.Users[1:],
			err:  nil,
		},
		{
			data: prepData(model.PresentationPerspective_CONTEXT, []string{"s1", "s2"}),
			resp: arc.Users,
			err:  nil,
		},
	}
	for i, test := range testUser {
		g, err := Process(context.Background(), test.data)
		if err != nil {
			t.Error(err)
		}
		resp, err := g.GetUsers()
		if err != test.err {
			t.Errorf("Test %d fail: mismatch error code", i)
			return
		}
		if len(resp) != len(test.resp) {
			t.Errorf("Test %d fail: mismatch length. Expect %d, get %d", i, len(test.resp), len(resp))
			return
		}
		for j, u := range resp {
			if u.Name != test.resp[j].Name {
				t.Errorf("Test %d fail: mismath value", i)
			}
		}
	}
}
func TestGetInternalSystems(t *testing.T) {
	testInternal := []struct {
		data *model.RenderRequest
		resp []model.InternalSystem
		err  error
	}{
		{
			data: prepData(model.PresentationPerspective_CONTEXT, []string{}),
			resp: arc.InternalSystems,
			err:  nil,
		},
		{
			data: prepData(model.PresentationPerspective_CONTEXT, []string{"s1"}),
			resp: arc.InternalSystems[:2],
			err:  nil,
		},
		{
			data: prepData(model.PresentationPerspective_CONTEXT, []string{"s2"}),
			resp: arc.InternalSystems[:2],
			err:  nil,
		},
		{
			data: prepData(model.PresentationPerspective_CONTEXT, []string{"s1", "s2"}),
			resp: arc.InternalSystems[:2],
			err:  nil,
		},
		{
			data: prepData(model.PresentationPerspective_CONTEXT, []string{"s1", "s2", "s3"}),
			resp: arc.InternalSystems,
			err:  nil,
		},
	}
	for i, test := range testInternal {
		g, err := Process(context.Background(), test.data)
		if err != nil {
			t.Error(err)
		}
		resp, err := g.GetInternalSystems()
		if err != test.err {
			t.Errorf("Test %d fail: Error mismatch", i)
		}
		if len(resp) != len(test.resp) {
			t.Errorf("Test %d fail: Length mismatch", i)
		}
		for j := range resp {
			if resp[j].Name != test.resp[j].Name {
				t.Errorf("Test %d fail: Wrong response", i)
				log.Println("Got: ", resp[j], "Expect:", test.resp[j])
			}
		}
	}
	return
}

func TestGetRelations(t *testing.T) {
	testRels := []struct {
		data *model.RenderRequest
		resp []model.Relation
		err  error
	}{
		{
			data: prepData(model.PresentationPerspective_CONTEXT, []string{}),
			resp: arc.Relations[:5],
			err:  nil,
		},
	}
	for i, test := range testRels {
		g, err := Process(context.Background(), test.data)
		if err != nil {
			t.Error(err)
		}
		resp, err := g.GetRelations()
		if err != test.err {
			t.Errorf("Test %d fail: mismatch error code", i)
			return
		}
		if len(resp) != len(test.resp) {
			t.Errorf("Test %d fail: mismatch length. Expect %d, get %d", i, len(test.resp), len(resp))
			return
		}
		for j, r := range resp {
			if r.Object != test.resp[j].Object {
				t.Errorf("Test %d fail: mismath value", i)
				return
			}
			if r.Subject != test.resp[j].Subject {
				t.Errorf("Test %d fail: mismath value", i)
				return
			}
		}
	}
}
