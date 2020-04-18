package puml

import (
	"testing"

	model "github.com/koderizer/arc/model"
)

func TestC4ContextPuml(t *testing.T) {
	var contextTests = []struct {
		in  model.ArcType
		out string
		err error
	}{
		{
			model.ArcType{
				App:             "context-test",
				Desc:            "This is a test",
				Users:           []model.ArchUser{{Name: "tester", Desc: "one who test"}},
				InternalSystems: []model.ArchInternalSystem{{Name: "testsys", Role: "To test system", Desc: "system test"}},
				ExternalSystems: []model.ArchExternalSystem{{Name: "testexternsys", Role: "To test external system", Desc: "external system test"}},
			},
			`@startuml
!include C4-PlantUML/C4_Context.puml
LAYOUT_WITH_LEGEND
title System Context Diagram for context-test 

Person(tester, "", "one who test")


System(testsys, "To test system","system test")


System_Ext(testexternsys, "To test external system", "external system test")

' 
@enduml`,
			nil,
		},
	}

	for _, tt := range contextTests {
		actual, err := C4ContextPuml(tt.in)
		if actual != tt.out || err != tt.err {
			t.Errorf("C4Context(%+v) expect \n%+v\n, actual puml is\n%+v\n, actual error is %s", tt.in, tt.out, actual, err)
		}
	}
}
