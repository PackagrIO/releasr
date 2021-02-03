package engine

import (
	"fmt"
	"github.com/packagrio/go-common/metadata"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/go-common/scm"
	"github.com/packagrio/releasr/pkg/config"
	releasrUtils "github.com/packagrio/releasr/pkg/utils"
	"io/ioutil"
	"path"
	"strings"
)

type engineGeneric struct {
	engineBase

	Scm          scm.Interface //Interface
	NextMetadata *metadata.GenericMetadata
}

func (g *engineGeneric) Init(pipelineData *pipeline.Data, configData config.Interface, sourceScm scm.Interface) error {
	g.Scm = sourceScm
	g.Config = configData
	g.PipelineData = pipelineData
	g.NextMetadata = new(metadata.GenericMetadata)

	//set command defaults (can be overridden by repo/system configuration)
	g.Config.SetDefault(config.PACKAGR_GENERIC_VERSION_TEMPLATE, `version := "%d.%d.%d"`)
	g.Config.SetDefault(config.PACKAGR_VERSION_METADATA_PATH, "VERSION")

	return nil
}

func (g *engineGeneric) GetNextMetadata() interface{} {
	return g.NextMetadata
}

func (g *engineGeneric) ValidateTools() error {
	return nil
}

func (g *engineGeneric) PackageStep() error {

	signature := releasrUtils.GitSignature(g.Config.GetString(config.PACKAGR_GIT_AUTHOR_NAME), g.Config.GetString(config.PACKAGR_GIT_AUTHOR_EMAIL))

	if cerr := releasrUtils.GitCommit(g.PipelineData.GitLocalPath, fmt.Sprintf("(v%s) %s", g.NextMetadata.Version, g.Config.GetString(config.PACKAGR_VERSION_BUMP_MESSAGE)), signature); cerr != nil {
		return cerr
	}
	tagCommit, terr := releasrUtils.GitTag(g.PipelineData.GitLocalPath, fmt.Sprintf("v%s", g.NextMetadata.Version), g.Config.GetString(config.PACKAGR_VERSION_BUMP_MESSAGE), signature)
	if terr != nil {
		return terr
	}

	g.PipelineData.ReleaseCommit = tagCommit
	return nil
}

func (g *engineGeneric) PopulateReleaseVersion() error {
	err := g.retrieveCurrentMetadata(g.PipelineData.GitLocalPath)
	if err != nil {
		return err
	}
	g.PipelineData.ReleaseVersion = g.NextMetadata.Version
	return nil
}

//Helpers
func (g *engineGeneric) retrieveCurrentMetadata(gitLocalPath string) error {
	//read VERSION file.
	versionContent, rerr := ioutil.ReadFile(path.Join(gitLocalPath, g.Config.GetString(config.PACKAGR_VERSION_METADATA_PATH)))
	if rerr != nil {
		return rerr
	}

	major := 0
	minor := 0
	patch := 0
	template := g.Config.GetString(config.PACKAGR_GENERIC_VERSION_TEMPLATE)
	fmt.Sscanf(strings.TrimSpace(string(versionContent)), template, &major, &minor, &patch)

	g.NextMetadata.Version = fmt.Sprintf("%d.%d.%d", major, minor, patch)
	return nil
}
