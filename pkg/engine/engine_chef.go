package engine

import (
	"fmt"
	"github.com/packagrio/go-common/errors"
	"github.com/packagrio/go-common/metadata"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/go-common/scm"
	"github.com/packagrio/releasr/pkg/config"
	"github.com/packagrio/releasr/pkg/utils"
	"os/exec"
)

type engineChef struct {
	engineBase
	NextMetadata *metadata.ChefMetadata
	Scm          scm.Interface
}

func (g *engineChef) Init(pipelineData *pipeline.Data, configData config.Interface, sourceScm scm.Interface) error {
	g.Config = configData
	g.Scm = sourceScm
	g.PipelineData = pipelineData
	g.NextMetadata = new(metadata.ChefMetadata)

	return nil
}

func (g *engineChef) GetNextMetadata() interface{} {
	return g.NextMetadata
}

func (g *engineChef) ValidateTools() error {
	if _, kerr := exec.LookPath("knife"); kerr != nil {
		return errors.EngineValidateToolError("knife binary is missing")
	}
	// TODO, check for knife spork
	return nil
}

func (g *engineChef) PackageStep() error {

	signature := utils.GitSignature(g.Config.GetString("engine_git_author_name"), g.Config.GetString("engine_git_author_email"))

	if cerr := utils.GitCommit(
		g.PipelineData.GitLocalPath,
		fmt.Sprintf("(v%s) %s", g.NextMetadata.Version,
			g.Config.GetString("engine_version_bump_msg")),
		signature); cerr != nil {
		return cerr
	}
	tagCommit, terr := utils.GitTag(g.PipelineData.GitLocalPath,
		fmt.Sprintf("v%s", g.NextMetadata.Version), g.Config.GetString("engine_version_bump_msg"),
		signature)
	if terr != nil {
		return terr
	}

	g.PipelineData.ReleaseCommit = tagCommit
	g.PipelineData.ReleaseVersion = g.NextMetadata.Version
	return nil
}
