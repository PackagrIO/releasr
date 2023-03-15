package engine

import (
	"encoding/json"
	"fmt"
	"github.com/packagrio/go-common/errors"
	"github.com/packagrio/go-common/metadata"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/go-common/scm"
	"github.com/packagrio/go-common/utils/git"
	"github.com/packagrio/releasr/pkg/config"
	"io/ioutil"
	"os/exec"
	"path"
)

type engineNode struct {
	engineBase

	Scm          scm.Interface //Interface
	NextMetadata *metadata.NodeMetadata
}

func (g *engineNode) Init(pipelineData *pipeline.Data, config config.Interface, sourceScm scm.Interface) error {
	g.Scm = sourceScm
	g.Config = config
	g.PipelineData = pipelineData
	g.NextMetadata = new(metadata.NodeMetadata)

	return nil
}

func (g *engineNode) GetNextMetadata() interface{} {
	return g.NextMetadata
}

func (g *engineNode) ValidateTools() error {

	if _, kerr := exec.LookPath("node"); kerr != nil {
		return errors.EngineValidateToolError("node binary is missing")
	}

	return nil
}

func (g *engineNode) PackageStep() error {
	signature := git.GitSignature(g.Config.GetString(config.PACKAGR_GIT_AUTHOR_NAME), g.Config.GetString(config.PACKAGR_GIT_AUTHOR_EMAIL))

	if cerr := git.GitCommit(g.PipelineData.GitLocalPath, fmt.Sprintf("(v%s) %s", g.NextMetadata.Version, g.Config.GetString(config.PACKAGR_VERSION_BUMP_MESSAGE)), signature); cerr != nil {
		return cerr
	}

	tagCommit, terr := git.GitTag(g.PipelineData.GitLocalPath, fmt.Sprintf("v%s", g.NextMetadata.Version), g.Config.GetString(config.PACKAGR_VERSION_BUMP_MESSAGE), signature)
	if terr != nil {
		return terr
	}

	g.PipelineData.ReleaseCommit = tagCommit
	return nil
}

func (g *engineNode) PopulateReleaseVersion() error {
	err := g.retrieveCurrentMetadata(g.PipelineData.GitLocalPath)
	if err != nil {
		return err
	}
	g.PipelineData.ReleaseVersion = g.NextMetadata.Version
	return nil
}

//private Helpers

func (g *engineNode) retrieveCurrentMetadata(gitLocalPath string) error {
	//read package.json file.
	packageContent, rerr := ioutil.ReadFile(path.Join(gitLocalPath, "package.json"))
	if rerr != nil {
		return rerr
	}

	if uerr := json.Unmarshal(packageContent, g.NextMetadata); uerr != nil {
		return uerr
	}

	return nil
}
