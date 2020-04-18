package server

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/koderizer/arc/model"
	puml "github.com/koderizer/arc/viz/puml"
)

const RENDER_TIMEOUT_SECONDS int = 3
const RENDER_KEEPALIVE_SECONDS int = 3
const RENDER_RESPONSE_TIMEOUT int = 3

type ArcViz struct {
	PumlRenderURI string
}

//NewArcViz initialized the Plant-UML rendering viz
func NewArcViz(plantUmlAddress string) *ArcViz {
	return &ArcViz{fmt.Sprintf("%s/form", plantUmlAddress)}
}

func (s *ArcViz) doPumlRender(ctx context.Context, pumlSrc string) ([]byte, error) {
	postData := url.Values{}
	postData.Add("text", pumlSrc)
	req, _ := http.NewRequest("POST", s.PumlRenderURI, strings.NewReader(postData.Encode()))
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

	buf, err := httputil.DumpResponse(resp, true)
	if err != nil {
		log.Print("Fail to read output data")
		return nil, error.New("Fail to get result")
	}
	return buf, nil
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
		log.Fatalf("Fail to decode data: %v", err)
	}

	//conver to puml command
	pumlSrc, err := puml.C4ContextPuml(arcData)
	if err != nil {
		log.Fatalf("Fail to generate PUML script from data: %+v", err)
	}

	//Send to puml rederer

	return &model.ArcPresentation{}, nil
}
