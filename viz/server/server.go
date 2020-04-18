package server

import (
	"context"

	"github.com/koderizer/arc/model"
	"google.golang.org/grpc"
)

type ArcViz struct {
}

//Render implement the rendering through PUML
func (c *ArcViz) Render(ctx context.Context, in *model.RenderRequest, opts ...grpc.CallOption) (*model.ArcPresentation, error) {

	return &model.ArcPresentation{}, nil
}
