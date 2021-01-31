package engine

import (
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/releasr/pkg/config"
)

type engineBase struct {
	Config       config.Interface
	PipelineData *pipeline.Data
}
