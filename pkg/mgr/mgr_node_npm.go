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

func DetectNodeNpm(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) bool {
	npmPath := path.Join(pipelineData.GitLocalPath, "package.json")
	return utils.FileExists(npmPath)
}

type mgrNodeNpm struct {
	Config       config.Interface
	PipelineData *pipeline.Data
	Client       *http.Client
}

func (m *mgrNodeNpm) Init(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) error {
	m.PipelineData = pipelineData
	m.Config = myconfig

	if client != nil {
		//primarily used for testing.
		m.Client = client
	}

	return nil
}

func (m *mgrNodeNpm) MgrValidateTools() error {
	if _, kerr := exec.LookPath("npm"); kerr != nil {
		return errors.EngineValidateToolError("npm binary is missing")
	}
	return nil
}

func (m *mgrNodeNpm) MgrPackageStep(nextMetadata interface{}) error {
	if !m.Config.GetBool("mgr_keep_lock_file") {
		os.Remove(path.Join(m.PipelineData.GitLocalPath, "npm-shrinkwrap.json"))
		os.Remove(path.Join(m.PipelineData.GitLocalPath, "package-lock.json"))
		os.Remove(path.Join(m.PipelineData.GitLocalPath, "yarn.lock"))
	}
	return nil
}
