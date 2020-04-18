package server

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
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

const RENDER_TIMEOUT_SECONDS = 3
const RENDER_KEEPALIVE_SECONDS = 3
const RENDER_RESPONSE_TIMEOUT = 3

type ArcViz struct {
	PumlRenderURI string
}

//NewArcViz initialized the Plant-UML rendering viz
func NewArcViz(plantUmlAddress string) *ArcViz {
	return &ArcViz{plantUmlAddress}
}

func (s *ArcViz) doPumlRender(ctx context.Context, pumlSrc string, format model.ArcVisualFormat) ([]byte, error) {
	postData := url.Values{}
	postData.Add("text", pumlSrc)
	req, _ := http.NewRequest("POST", s.PumlRenderURI+"/form", strings.NewReader(postData.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlendcoded")

	var client http.RoundTripper = &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   RENDER_TIMEOUT_SECONDS * time.Second,
			KeepAlive: RENDER_KEEPALIVE_SECONDS * time.Second,
		}).Dial,
		ResponseHeaderTimeout: time.Duration(RENDER_RESPONSE_TIMEOUT * time.Second),
		DisableKeepAlives:     true,
	}

	resp, err := client.RoundTrip(req)
	if err != nil {
		log.Panicf("Internal Server error, unable to request render engine: %+v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Internal Server error, fail to render with code: %+v", resp.StatusCode)
		return nil, errors.New("Render Failed")
	}

	outputURI, err := resp.Location()
	if err != nil {
		log.Printf("Unable to resolve the output location: %s", err)
		return nil, err
	}

	outputID := strings.TrimPrefix(outputURI.Path, "/uml/")
	var outputPath string
	switch format {
	case model.ArcVisualFormat_PNG:
		outputPath = fmt.Sprintf("/png/%s", outputID)
	case model.ArcVisualFormat_PDF:
		outputPath = fmt.Sprintf("/pdf/%s", outputID)
	case model.ArcVisualFormat_SVG:
		outputPath = fmt.Sprintf("/svg/%s", outputID)
	}

	datReq, _ := http.NewRequest("GET", outputPath, nil)

	respOut, err := client.RoundTrip(datReq)
	if err != nil {
		log.Panic("Internal server error, unable to get file form render engine: %+v", err)
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

	return &model.ArcPresentation{in.VisualFormat, output}, nil
}
