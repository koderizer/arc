package server

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/koderizer/arc/model"
	puml "github.com/koderizer/arc/viz/puml"
)

const renderTimeoutSecond = 3
const renderKeepaliveSecond = 3
const renderResponseTime = 3

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

	//Create new graph from given source and extract the generated URI id
	postData := url.Values{}
	postData.Add("text", pumlSrc)
	req, _ := http.NewRequest("POST", s.PumlRenderURI+"/form", strings.NewReader(postData.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlendcoded")

	log.Printf("Generate code:\n%s", pumlSrc)
	var client http.RoundTripper = &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   renderTimeoutSecond * time.Second,
			KeepAlive: renderKeepaliveSecond * time.Second,
		}).Dial,
		ResponseHeaderTimeout: time.Duration(renderResponseTime * time.Second),
		DisableKeepAlives:     true,
	}

	resp, err := client.RoundTrip(req)
	if err != nil {
		log.Printf("Internal Server error, unable to request render engine: %+v", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusFound {
		log.Printf("Internal Server error, fail to render with code: %+v", resp.StatusCode)
		return nil, errors.New("Render Failed")
	}

	//Download data content
	outputURI, err := resp.Location()
	if err != nil {
		log.Printf("Unable to resolve the output location: %s", err)
		return nil, err
	}
	resp.Body.Close()

	outputID := strings.TrimPrefix(outputURI.Path, "/uml/")
	outputPath += outputID

	log.Printf("File generated: %s", outputPath)
	datReq, _ := http.NewRequest("GET", outputPath, nil)

	respOut, err := client.RoundTrip(datReq)
	if err != nil {
		log.Printf("Internal server error, unable to get file from render engine: %+v", err)
		return nil, errors.New("Render Failed")
	}
	defer respOut.Body.Close()

	return ioutil.ReadAll(respOut.Body)
}

//Render implement the rendering through PUML
func (s *ArcViz) Render(ctx context.Context, in *model.RenderRequest) (*model.ArcPresentation, error) {

	if in.DataFormat != model.ArcDataFormat_ARC {
		return nil, errors.New("Unsupported data format, this server only support ARC data format")
	}

	//decode the data
	var arcData model.ArcType
	dec := gob.NewDecoder(bytes.NewBuffer(in.Data))
	if err := dec.Decode(&arcData); err != nil {
		log.Printf("Fail to decode data: %v", err)
		return nil, err
	}

	//conver to puml command
	pumlSrc, err := puml.C4ContextPuml(arcData)
	if err != nil {
		log.Printf("Fail to generate PUML script from data: %+v", err)
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
