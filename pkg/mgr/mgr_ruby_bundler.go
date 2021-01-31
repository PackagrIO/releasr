package mgr

import (
	"fmt"
	"github.com/analogj/go-util/utils"
	"github.com/packagrio/go-common/errors"
	"github.com/packagrio/go-common/metadata"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/releasr/pkg/config"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
)

func DetectRubyBundler(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) bool {
	//theres no way to automatically determine if a project was created via Yarn (vs NPM)
	return false
}

type mgrRubyBundler struct {
	Config       config.Interface
	PipelineData *pipeline.Data
	Client       *http.Client
}

func (m *mgrRubyBundler) Init(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) error {
	m.PipelineData = pipelineData
	m.Config = myconfig

	if client != nil {
		//primarily used for testing.
		m.Client = client
	}

	return nil
}

func (m *mgrRubyBundler) MgrValidateTools() error {
	if _, berr := exec.LookPath("gem"); berr != nil {
		return errors.EngineValidateToolError("gem binary is missing")
	}

	if _, berr := exec.LookPath("bundle"); berr != nil {
		return errors.EngineValidateToolError("bundle binary is missing")
	}
	return nil
}

func (m *mgrRubyBundler) MgrPackageStep(nextMetadata interface{}) error {
	if !m.Config.GetBool("mgr_keep_lock_file") {
		os.Remove(path.Join(m.PipelineData.GitLocalPath, "Gemfile.lock"))
	}
	return nil
}

func (m *mgrRubyBundler) MgrDistStep(nextMetadata interface{}) error {
	if !m.Config.IsSet("rubygems_api_key") {
		return errors.MgrDistCredentialsMissing("Cannot deploy package to rubygems, credentials missing")
	}

	credFile, _ := ioutil.TempFile("", "gem_credentials")
	defer os.Remove(credFile.Name())

	// write the .gem/credentials config jfile.

	credContent := fmt.Sprintf(utils.StripIndent(
		`---
		:rubygems_api_key: %s
		`),
		m.Config.GetString("rubygems_api_key"),
	)

	if _, perr := credFile.Write([]byte(credContent)); perr != nil {
		return perr
	}

	pushCmd := fmt.Sprintf("gem push %s --config-file %s",
		fmt.Sprintf("%s-%s.gem", nextMetadata.(*metadata.RubyMetadata).Name, nextMetadata.(*metadata.RubyMetadata).Version),
		credFile.Name(),
	)
	if derr := utils.BashCmdExec(pushCmd, m.PipelineData.GitLocalPath, nil, ""); derr != nil {
		return errors.MgrDistPackageError("Pushing gem to RubyGems.org using `gem push` failed. Check log for exact error")
	}

	return nil
}
