package engine

import (
	"fmt"
	"github.com/packagrio/go-common/errors"
	"github.com/packagrio/go-common/metadata"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/go-common/scm"
	"github.com/packagrio/releasr/pkg/config"
	releasrUtils "github.com/packagrio/releasr/pkg/utils"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
)

type enginePython struct {
	engineBase

	Scm          scm.Interface //Interface
	NextMetadata *metadata.PythonMetadata
}

func (g *enginePython) Init(pipelineData *pipeline.Data, configData config.Interface, sourceScm scm.Interface) error {
	g.Scm = sourceScm
	g.Config = configData
	g.PipelineData = pipelineData
	g.NextMetadata = new(metadata.PythonMetadata)

	//set command defaults (can be overridden by repo/system configuration)
	g.Config.SetDefault(config.PACKAGR_VERSION_METADATA_PATH, "VERSION")
	return g.retrieveCurrentMetadata(g.PipelineData.GitLocalPath)
}

func (g *enginePython) GetNextMetadata() interface{} {
	return g.NextMetadata
}

func (g *enginePython) ValidateTools() error {
	if _, berr := exec.LookPath("python"); berr != nil {
		return errors.EngineValidateToolError("python binary is missing")
	}

	return nil
}

func (g *enginePython) PackageStep() error {
	os.RemoveAll(path.Join(g.PipelineData.GitLocalPath, ".tox")) //remove .tox folder.

	signature := releasrUtils.GitSignature(g.Config.GetString("engine_git_author_name"), g.Config.GetString("engine_git_author_email"))

	if cerr := releasrUtils.GitCommit(g.PipelineData.GitLocalPath, fmt.Sprintf("(v%s) %s", g.NextMetadata.Version, g.Config.GetString("engine_version_bump_msg")), signature); cerr != nil {
		return cerr
	}
	tagCommit, terr := releasrUtils.GitTag(g.PipelineData.GitLocalPath, fmt.Sprintf("v%s", g.NextMetadata.Version), g.Config.GetString("engine_version_bump_msg"), signature)
	if terr != nil {
		return terr
	}

	g.PipelineData.ReleaseCommit = tagCommit
	g.PipelineData.ReleaseVersion = g.NextMetadata.Version
	return nil
}

//private Helpers

func (g *enginePython) retrieveCurrentMetadata(gitLocalPath string) error {
	//read metadata.json file.
	versionContent, rerr := ioutil.ReadFile(path.Join(gitLocalPath, g.Config.GetString(config.PACKAGR_VERSION_METADATA_PATH)))
	if rerr != nil {
		return rerr
	}
	g.NextMetadata.Version = strings.TrimSpace(string(versionContent))
	return nil
}
