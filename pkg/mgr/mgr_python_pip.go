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

func DetectPythonPip(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) bool {
	//theres no way to automatically determine if a project was created via Yarn (vs NPM)
	return false
}

type mgrPythonPip struct {
	Config       config.Interface
	PipelineData *pipeline.Data
	Client       *http.Client
}

func (m *mgrPythonPip) Init(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) error {
	m.PipelineData = pipelineData
	m.Config = myconfig

	if client != nil {
		//primarily used for testing.
		m.Client = client
	}

	return nil
}

func (m *mgrPythonPip) MgrValidateTools() error {
	if _, berr := exec.LookPath("twine"); berr != nil {
		return errors.EngineValidateToolError("twine binary is missing")
	}
	if _, berr := exec.LookPath("pip"); berr != nil {
		return errors.EngineValidateToolError("pip binary is missing")
	}
	return nil
}

func (m *mgrPythonPip) MgrPackageStep(nextMetadata interface{}) error {
	if !m.Config.GetBool("mgr_keep_lock_file") {
		os.Remove(path.Join(m.PipelineData.GitLocalPath, "requirements.txt"))
	}
	return nil
}
