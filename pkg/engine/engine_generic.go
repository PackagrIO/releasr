package engine

import (
	"bufio"
	"fmt"
	"github.com/packagrio/go-common/errors"
	"github.com/packagrio/go-common/metadata"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/go-common/scm"
	"github.com/packagrio/go-common/utils/git"
	"github.com/packagrio/releasr/pkg/config"
	"os"
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

func (g *engineGeneric) PopulateReleaseVersion() error {
	err := g.retrieveCurrentMetadata(g.PipelineData.GitLocalPath)
	if err != nil {
		return err
	}
	g.PipelineData.ReleaseVersion = g.NextMetadata.Version
	return nil
}

// Helpers
func (g *engineGeneric) retrieveCurrentMetadata(gitLocalPath string) error {
	//read VERSION file.
	filePath := path.Join(gitLocalPath, g.Config.GetString(config.PACKAGR_VERSION_METADATA_PATH))
	template := g.Config.GetString(config.PACKAGR_GENERIC_VERSION_TEMPLATE)

	// Handle if the user wants to merge the version file and not overwrite it
	if g.Config.GetBool(config.PACKAGR_GENERIC_MERGE_VERSION_FILE) {
		versionContent, err := g.matchAsSingleLine(filePath, template)
		if err != nil {
			return err
		}
		g.NextMetadata.Version = versionContent
		return nil
	}

	match, err := g.matchAsMultiLine(filePath, template)
	if err != nil {
		return err
	}
	g.NextMetadata.Version = match
	return nil
}

// Matches the template with the entire file, useful for simple version files
func (g *engineGeneric) matchAsMultiLine(filePath string, template string) (string, error) {
	versionContent, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return g.getVersionFromString(string(versionContent), template)
}

// Only matches the version for a single line, used when you have a version on a single line within a complete multiline file
func (g *engineGeneric) matchAsSingleLine(filePath string, template string) (string, error) {
	fileReader, rerr := os.Open(filePath)
	scanner := bufio.NewScanner(fileReader)
	if rerr != nil {
		return "", rerr
	}

	for scanner.Scan() {
		readLine := scanner.Text()
		version, err := g.getVersionFromString(readLine, template)
		if err != nil {
			continue
		}
		return version, nil
	}
	return "", errors.EngineUnspecifiedError(fmt.Sprintf(
		"Was unable to find a version with the format `%s` in file %s", template, filePath,
	))
}

func (g *engineGeneric) getVersionFromString(versionContent string, template string) (string, error) {
	major := 0
	minor := 0
	patch := 0
	_, err := fmt.Sscanf(strings.TrimSpace(string(versionContent)), template, &major, &minor, &patch)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d.%d.%d", major, minor, patch), nil
}
