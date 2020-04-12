package main

import (
	"testing"

	"github.com/koderizer/arc/src/arc/model"
)

func TestC4ContextPuml(t *testing.T) {
	var contextTests = []struct {
		in model.ArchType
		out struct { string error}
	} {
		{
			in: model.ArchType{
				App: "context-test", 
				Desc: "This is a test", 
				Users: []struct{string,string}{{Name: "tester", Desc: "one who test"}}
			},
			out: {
				"somefailthing",
				nil
			}
		}

	}

	for _, tt := range contextTests {
		actual := C4ContextPuml(tt)
		if actual != tt.out {
			t.Errorf("C4Context(%+v) expect %+v, actual %+v",tt.in, tt.out, actual)
		}
	}
}
