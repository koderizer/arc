package server

import (
	"context"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"

	"github.com/koderizer/arc/model"
	"github.com/koderizer/arc/viz/analyzer"
	"github.com/koderizer/arc/viz/puml"
)

const renderTimeoutSecond = 3
const renderKeepaliveSecond = 3
const renderResponseTime = 3
const defaultPlantUmlID = "SyfFKj2rKt3CoKnELR1Io4ZDoSa70000"

//ArcViz type is the core config of ArcViz server
type ArcViz struct {
	PumlRenderURI string
}

//NewArcViz initialized the Plant-UML rendering viz
func NewArcViz(plantUmlAddress string) *ArcViz {
	return &ArcViz{plantUmlAddress}
}

func (s *ArcViz) doPumlRender(ctx context.Context, pumlSrc string, format model.ArcVisualFormat) ([]byte, error) {

	var outputPath = s.PumlRenderURI
	switch format {
	case model.ArcVisualFormat_PNG:
		outputPath += "/png/"
	case model.ArcVisualFormat_SVG:
		outputPath += "/svg/"
	default:
		log.Printf("Requested format %+v recieved. Not supported", format)
		return nil, errors.New("Not supported")
	}

	log.Printf("Generate code:\n%s", pumlSrc)

	resp, err := http.PostForm(s.PumlRenderURI+"/form", url.Values{
		"text": {pumlSrc},
	})
	if err != nil {
		log.Printf("Internal Server error, unable to request render engine: %+v", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Internal Server error, fail to render with code: %+v", resp.StatusCode)
		return nil, errors.New("Render Failed")
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Unable to resolve the output location: %s", err)
		return nil, err
	}

	re := regexp.MustCompile(outputPath + `(.*?)"`)
	matches := re.FindSubmatch(body)
	if matches == nil {
		log.Print("Unable to get outputID")
		return nil, errors.New("Render output location missing")
	}
	outputID := string(matches[1])

	if outputID == defaultPlantUmlID {
		log.Printf("Syntax error, default image returned")
		return nil, errors.New("Render Failed")
	}
	outputPath += outputID

	return []byte(outputPath), nil
	// image, err := http.Get(outputPath)
	// if err != nil {
	// 	log.Printf("Internal server error, unable to get file from render engine: %+v", err)
	// 	return nil, errors.New("Render Failed")
	// }
	// defer image.Body.Close()

	// return ioutil.ReadAll(image.Body)
}

//Render implement the rendering through PUML
func (s *ArcViz) Render(ctx context.Context, in *model.RenderRequest) (*model.ArcPresentation, error) {

	g, err := analyzer.Process(ctx, in)
	if err != nil {
		return nil, err
	}
	arc := model.ArcType{
		App:  g.Arc.App,
		Desc: g.Arc.Desc,
	}
	arc.InternalSystems, err = g.GetInternalSystems()
	if err != nil {
		return nil, err
	}
	arc.ExternalSystems, err = g.GetExternalSystems()
	if err != nil {
		return nil, err
	}
	arc.Relations, err = g.GetRelations()
	if err != nil {
		return nil, err
	}
	arc.Users, err = g.GetUsers()
	if err != nil {
		return nil, err
	}
	pumlSrc, err := puml.C4ContextPuml(arc, in.GetTarget()...)
	if err != nil {
		return nil, err
	}
	//Send to puml rederer
	output, err := s.doPumlRender(ctx, pumlSrc, in.VisualFormat)
	if err != nil {
		log.Printf("Fail to render %+v", err)
		return nil, err
	}

	return &model.ArcPresentation{Format: in.VisualFormat, Data: output}, nil
}
