package mgr

import (
	"github.com/packagrio/go-common/errors"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/releasr/pkg/config"
	"net/http"
	"os"
	"os/exec"
	"path"
)

func DetectNodeYarn(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) bool {
	//theres no way to automatically determine if a project was created via Yarn (vs NPM)
	return false
}

type mgrNodeYarn struct {
	Config       config.Interface
	PipelineData *pipeline.Data
	Client       *http.Client
}

func (m *mgrNodeYarn) Init(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) error {
	m.PipelineData = pipelineData
	m.Config = myconfig

	if client != nil {
		//primarily used for testing.
		m.Client = client
	}

	return nil
}

func (m *mgrNodeYarn) MgrValidateTools() error {
	if _, kerr := exec.LookPath("yarn"); kerr != nil {
		return errors.EngineValidateToolError("yarn binary is missing")
	}
	return nil
}

func (m *mgrNodeYarn) MgrPackageStep(nextMetadata interface{}) error {
	if !m.Config.GetBool(config.PACKAGR_MGR_KEEP_LOCK_FILE) {
		os.Remove(path.Join(m.PipelineData.GitLocalPath, "npm-shrinkwrap.json"))
		os.Remove(path.Join(m.PipelineData.GitLocalPath, "package-lock.json"))
		os.Remove(path.Join(m.PipelineData.GitLocalPath, "yarn.lock"))
	}
	return nil
}
