package engine

import (
	"fmt"
	"github.com/packagrio/go-common/errors"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/go-common/scm"
	"github.com/packagrio/releasr/pkg/config"
)

func Create(engineType string, pipelineData *pipeline.Data, configImpl config.Interface, sourceImpl scm.Interface) (Interface, error) {

	var eng Interface

	switch engineType {
	case PACKAGR_ENGINE_TYPE_CHEF:
		eng = new(engineChef)
	case PACKAGR_ENGINE_TYPE_GENERIC:
		eng = new(engineGeneric)
	case PACKAGR_ENGINE_TYPE_GOLANG:
		eng = new(engineGolang)
	case PACKAGR_ENGINE_TYPE_NODE:
		eng = new(engineNode)
	case PACKAGR_ENGINE_TYPE_PYTHON:
		eng = new(enginePython)
	case PACKAGR_ENGINE_TYPE_RUBY:
		eng = new(engineRuby)
	default:
		return nil, errors.EngineUnspecifiedError(fmt.Sprintf("Unknown Engine Type: %s", engineType))
	}

	if err := eng.Init(pipelineData, configImpl, sourceImpl); err != nil {
		return nil, err
	}
	return eng, nil
}
