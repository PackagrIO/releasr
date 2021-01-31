package mgr

import (
	"github.com/analogj/go-util/utils"
	"github.com/packagrio/go-common/errors"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/releasr/pkg/config"
	"net/http"
	"os"
	"os/exec"
	"path"
)

func DetectChefBerkshelf(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) bool {
	berksfilePath := path.Join(pipelineData.GitLocalPath, "Berksfile")
	return utils.FileExists(berksfilePath)
}

type mgrChefBerkshelf struct {
	Config       config.Interface
	PipelineData *pipeline.Data
	Client       *http.Client
}

func (m *mgrChefBerkshelf) Init(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) error {
	m.PipelineData = pipelineData
	m.Config = myconfig

	if client != nil {
		//primarily used for testing.
		m.Client = client
	}

	return nil
}

func (m *mgrChefBerkshelf) MgrValidateTools() error {
	//a chef/berkshelf like environment needs to be available for this Engine
	if _, kerr := exec.LookPath("knife"); kerr != nil {
		return errors.EngineValidateToolError("knife binary is missing")
	}

	if _, berr := exec.LookPath("berks"); berr != nil {
		return errors.EngineValidateToolError("berkshelf binary is missing")
	}

	//TODO: figure out how to validate that "bundle audit" command exists.
	if _, berr := exec.LookPath("bundle"); berr != nil {
		return errors.EngineValidateToolError("bundler binary is missing")
	}
	return nil
}

func (m *mgrChefBerkshelf) MgrPackageStep(nextMetadata interface{}) error {
	if !m.Config.GetBool(config.PACKAGR_MGR_KEEP_LOCK_FILE) {
		os.Remove(path.Join(m.PipelineData.GitLocalPath, "Berksfile.lock"))
		os.Remove(path.Join(m.PipelineData.GitLocalPath, "Gemfile.lock"))
	}
	return nil
}
