package pkg

import (
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/go-common/scm"
	"github.com/packagrio/releasr/pkg/config"
	"github.com/packagrio/releasr/pkg/engine"
	"github.com/packagrio/releasr/pkg/mgr"
	"log"
	"os"
	"path/filepath"
)

type Pipeline struct {
	Data           *pipeline.Data
	Config         config.Interface
	Scm            scm.Interface
	Engine         engine.Interface
	PackageManager mgr.Interface
}

func (p *Pipeline) Start(configData config.Interface) error {
	// Initialize Pipeline.
	p.Config = configData
	p.Data = new(pipeline.Data)

	//by default the current working directory is the local directory to execute in
	cwdPath, _ := os.Getwd()
	p.Data.GitLocalPath = cwdPath
	p.Data.GitParentPath = filepath.Dir(cwdPath)

	//assumes that this is a git repository, and version file has already been bumped (using Bumpr)

	//Generate a new instance of the engine
	engineImpl, eerr := engine.Create(p.Config.GetString("package_type"), p.Data, p.Config, p.Scm)
	if eerr != nil {
		return eerr
	}
	p.Engine = engineImpl

	//initialize the manager
	if p.Config.IsSet("mgr_type") {
		manager, merr := mgr.Create(p.Config.GetString("mgr_type"), p.Data, p.Config, nil)
		if merr != nil {
			return merr
		}
		p.PackageManager = manager
	} else {
		manager, merr := mgr.Detect(p.Config.GetString("package_type"), p.Data, p.Config, nil)
		if merr != nil {
			return merr
		}
		p.PackageManager = manager
	}

	log.Println("mgr_package_step")
	if err := p.PackageManager.MgrPackageStep(p.Engine.GetNextMetadata()); err != nil {
		return err
	}
	log.Println("package_step")
	if err := p.Engine.PackageStep(); err != nil {
		return err
	}

	//////func (p *Pipeline) MgrDistStep() error {
	if p.Config.GetBool("mgr_disable_dist") {
		log.Println("skipping mgr_dist_step.pre, mgr_dist_step, mgr_dist_step.post")
		return nil
	}

	log.Println("mgr_dist_step")
	if err := p.PackageManager.MgrDistStep(p.Engine.GetNextMetadata()); err != nil {
		return err
	}

	return nil
}
