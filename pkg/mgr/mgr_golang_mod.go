package mgr

import (
	"github.com/analogj/go-util/utils"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/releasr/pkg/config"
	"net/http"
	"os"
	"path"
)

func DetectGolangMod(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) bool {
	gomodPath := path.Join(pipelineData.GitLocalPath, "go.mod")
	return utils.FileExists(gomodPath)
}

type mgrGolangMod struct {
	Config       config.Interface
	PipelineData *pipeline.Data
	Client       *http.Client
}

func (m *mgrGolangMod) Init(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) error {
	m.PipelineData = pipelineData
	m.Config = myconfig

	if client != nil {
		//primarily used for testing.
		m.Client = client
	}

	return nil
}

func (m *mgrGolangMod) MgrValidateTools() error {
	//if _, kerr := exec.LookPath("dep"); kerr != nil {
	//	return errors.EngineValidateToolError("dep binary is missing")
	//}
	return nil
}

func (m *mgrGolangMod) MgrPackageStep(nextMetadata interface{}) error {
	if !m.Config.GetBool(config.PACKAGR_MGR_KEEP_LOCK_FILE) {
		os.Remove(path.Join(m.PipelineData.GitLocalPath, "go.sum"))
	}
	return nil
}
