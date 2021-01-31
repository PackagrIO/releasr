package engine

import (
	"fmt"
	"github.com/analogj/go-util/utils"
	"github.com/packagrio/go-common/errors"
	"github.com/packagrio/go-common/metadata"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/go-common/scm"
	"github.com/packagrio/releasr/pkg/config"
	releasrUtils "github.com/packagrio/releasr/pkg/utils"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
)

type rubyGemspec struct {
	Name    string `json:"name"`
	Version struct {
		Version string `json:"name"`
	} `json:"version"`
}

type engineRuby struct {
	engineBase

	Scm          scm.Interface //Interface
	NextMetadata *metadata.RubyMetadata
	GemspecPath  string
}

func (g *engineRuby) Init(pipelineData *pipeline.Data, config config.Interface, sourceScm scm.Interface) error {
	g.Scm = sourceScm
	g.Config = config
	g.PipelineData = pipelineData
	g.NextMetadata = new(metadata.RubyMetadata)

	//set command defaults (can be overridden by repo/system configuration)
	return g.retrieveCurrentMetadata(g.PipelineData.GitLocalPath)
}

func (g *engineRuby) GetNextMetadata() interface{} {
	return g.NextMetadata
}

func (g *engineRuby) ValidateTools() error {
	if _, kerr := exec.LookPath("ruby"); kerr != nil {
		return errors.EngineValidateToolError("ruby binary is missing")
	}

	return nil
}

func (g *engineRuby) PackageStep() error {
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
func (g *engineRuby) retrieveCurrentMetadata(gitLocalPath string) error {
	//read Gemspec file.
	gemspecFiles, gerr := filepath.Glob(path.Join(gitLocalPath, "/*.gemspec"))
	if gerr != nil {
		return errors.EngineBuildPackageInvalid("*.gemspec file is required to process Ruby gem")
	} else if len(gemspecFiles) == 0 {
		return errors.EngineBuildPackageInvalid("*.gemspec file is required to process Ruby gem")
	}

	g.GemspecPath = gemspecFiles[0]

	gemspecJsonFile, _ := ioutil.TempFile("", "gemspec.json")
	defer os.Remove(gemspecJsonFile.Name())

	//generate a JSON-style YAML file containing the Gemspec data. (still not straight valid JSON).
	//
	gemspecJsonCmd := fmt.Sprintf("ruby -e \"require('yaml'); File.write('%s', YAML::to_json(Gem::Specification::load('%s')))\"",
		gemspecJsonFile.Name(),
		g.GemspecPath,
	)
	if cerr := utils.BashCmdExec(gemspecJsonCmd, "", nil, ""); cerr != nil {
		return errors.EngineBuildPackageFailed(fmt.Sprintf("Command (%s) failed. Check log for more details.", gemspecJsonCmd))
	}

	//Load gemspec JSON file and parse it.
	gemspecJsonContent, rerr := ioutil.ReadFile(gemspecJsonFile.Name())
	if rerr != nil {
		return rerr
	}

	gemspecObj := new(rubyGemspec)
	if uerr := yaml.Unmarshal(gemspecJsonContent, gemspecObj); uerr != nil {
		fmt.Println(string(gemspecJsonContent))
		return uerr
	}

	g.NextMetadata.Name = gemspecObj.Name
	g.NextMetadata.Version = gemspecObj.Version.Version

	//ensure that there is a lib/GEMNAME/version.rb file.
	versionrbPath := path.Join("lib", gemspecObj.Name, "version.rb")
	if !utils.FileExists(path.Join(g.PipelineData.GitLocalPath, versionrbPath)) {
		return errors.EngineBuildPackageInvalid(
			fmt.Sprintf("version.rb file (%s) is required to process Ruby gem", versionrbPath))
	}
	return nil
}
