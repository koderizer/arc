package main

import (
	"testing"

	model "github.com/koderizer/arc/src/arc/model"
)

func TestC4ContextPuml(t *testing.T) {
	var contextTests = []struct {
		in  model.ArchType
		out string
		err error
	}{
		{
			model.ArchType{
				App:   "context-test",
				Desc:  "This is a test",
				Users: []model.ArchUser{{Name: "tester", Desc: "one who test"}},
			},
			"somefailthing",
			nil,
		},
	}

	for _, tt := range contextTests {
		actual, err := C4ContextPuml(tt.in)
		if actual != tt.out || err != tt.err {
			t.Errorf("C4Context(%+v) expect %+v, actual puml is %+v, actual error is %s", tt.in, tt.out, actual, err)
		}
	}
}
