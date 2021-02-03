package engine

import (
	"fmt"
	"github.com/packagrio/go-common/errors"
	"github.com/packagrio/go-common/metadata"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/go-common/scm"
	"github.com/packagrio/releasr/pkg/config"
	releasrUtils "github.com/packagrio/releasr/pkg/utils"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os/exec"
	"path"
	"strings"
)

type engineGolang struct {
	engineBase

	Scm          scm.Interface //Interface
	NextMetadata *metadata.GolangMetadata
}

func (g *engineGolang) Init(pipelineData *pipeline.Data, configData config.Interface, sourceScm scm.Interface) error {
	g.Scm = sourceScm
	g.Config = configData
	g.PipelineData = pipelineData
	g.NextMetadata = new(metadata.GolangMetadata)

	//set command defaults (can be overridden by repo/system configuration)
	g.Config.SetDefault(config.PACKAGR_VERSION_METADATA_PATH, "pkg/version/version.go")

	return nil
}

func (g *engineGolang) GetNextMetadata() interface{} {
	return g.NextMetadata
}

func (g *engineGolang) ValidateTools() error {
	if _, kerr := exec.LookPath("go"); kerr != nil {
		return errors.EngineValidateToolError("go binary is missing")
	}

	return nil
}

func (g *engineGolang) PackageStep() error {
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

func (g *engineGolang) PopulateReleaseVersion() error {
	err := g.retrieveCurrentMetadata(g.PipelineData.GitLocalPath)
	if err != nil {
		return err
	}
	g.PipelineData.ReleaseVersion = g.NextMetadata.Version
	return nil
}

//private Helpers

func (g *engineGolang) retrieveCurrentMetadata(gitLocalPath string) error {

	versionContent, rerr := ioutil.ReadFile(path.Join(g.PipelineData.GitLocalPath, g.Config.GetString(config.PACKAGR_VERSION_METADATA_PATH)))
	if rerr != nil {
		return rerr
	}

	//Oh.My.God.

	// Create the AST by parsing src.
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, "", string(versionContent), 0)
	if err != nil {
		return err
	}

	version, verr := g.parseGoVersion(f.Decls)
	if verr != nil {
		return verr
	}

	g.NextMetadata.Version = version
	return nil
}

func (g *engineGolang) parseGoVersion(list []ast.Decl) (string, error) {
	//find version declaration (uppercase or lowercase)
	for _, decl := range list {
		gen := decl.(*ast.GenDecl)
		if gen.Tok == token.CONST || gen.Tok == token.VAR {
			for _, spec := range gen.Specs {
				valSpec := spec.(*ast.ValueSpec)
				if strings.ToLower(valSpec.Names[0].Name) == "version" {
					//found the version variable.
					return strings.Trim(valSpec.Values[0].(*ast.BasicLit).Value, "\"'"), nil
				}
			}
		}
	}
	return "", errors.EngineBuildPackageFailed(fmt.Sprintf("Could not retrieve the version from %s", g.Config.GetString(config.PACKAGR_VERSION_METADATA_PATH)))
}
