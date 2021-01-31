package engine

import (
	"fmt"
	"github.com/packagrio/go-common/errors"
	"github.com/packagrio/go-common/metadata"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/go-common/scm"
	"github.com/packagrio/releasr/pkg/config"
	releasrUtils "github.com/packagrio/releasr/pkg/utils"
	"os"
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
	var scmDomain string
	if g.Config.GetString(config.PACKAGR_SCM) == "bitbucket" {
		scmDomain = "bitbucket.org"
	} else {
		scmDomain = "github.com"
	}

	g.Config.SetDefault("engine_golang_package_path", fmt.Sprintf("%s/%s", scmDomain, strings.ToLower(g.Config.GetString("scm_repo_full_name"))))

	//TODO: figure out why setting the GOPATH workspace is causing the tools to timeout.
	// golang recommends that your in-development packages are in the GOPATH and glide requires it to do glide install.
	// the problem with this is that for somereason gometalinter (and the underlying linting tools) take alot longer
	// to run, and hit the default deadline limit ( --deadline=30s).
	// we can have multiple workspaces in the gopath by separating them with colon (:), but this timeout is nasty if not required.
	//TODO: g.GoPath root will not be deleted (its the parent of GitParentPath), figure out if we can do this automatically.
	g.PipelineData.GolangGoPath = g.PipelineData.GitParentPath
	os.Setenv("GOPATH", fmt.Sprintf("%s:%s", os.Getenv("GOPATH"), g.PipelineData.GolangGoPath))

	// A proper gopath has a bin and src directory.
	goPathBin := path.Join(g.PipelineData.GitParentPath, "bin")
	goPathSrc := path.Join(g.PipelineData.GitParentPath, "src")
	os.MkdirAll(goPathBin, 0666)
	os.MkdirAll(goPathSrc, 0666)

	//  the gopath bin directory should aslo be added to Path
	os.Setenv("PATH", fmt.Sprintf("%s:%s", os.Getenv("PATH"), goPathBin))

	packagePathPrefix := path.Dir(g.Config.GetString("engine_golang_package_path")) //strip out the repo name.
	// customize the git parent path for Golang Engine
	g.PipelineData.GitParentPath = path.Join(g.PipelineData.GitParentPath, "src", packagePathPrefix)
	os.MkdirAll(g.PipelineData.GitParentPath, 0666)

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
