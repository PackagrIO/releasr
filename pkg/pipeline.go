package pkg

import (
	"errors"
	"github.com/analogj/go-util/utils"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/go-common/scm"
	"github.com/packagrio/releasr/pkg/config"
	"github.com/packagrio/releasr/pkg/engine"
	"github.com/packagrio/releasr/pkg/mgr"
	"log"
	"os"
	"path"
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

	if err := p.PipelineInitStep(); err != nil {
		return err
	}

	//Parse Repo config if present.
	if err := p.ParseRepoConfig(); err != nil {
		return err
	}

	if err := p.Engine.PopulateNextMetadata(); err != nil {
		return err
	}

	//initialize the manager
	if err := p.MgrInitStep(); err != nil {
		return err
	}

	// validate tools
	if err := p.ValidateTools(); err != nil {
		return err
	}
	if err := p.MgrValidateTools(); err != nil {
		return err
	}

	//package repository
	if err := p.PackageStep(); err != nil {
		return err
	}
	return nil
}

func (p *Pipeline) PipelineInitStep() error {
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

	return nil
}

func (p *Pipeline) ParseRepoConfig() error {
	log.Println("parse_repo_config")
	// update the config with repo config file options
	repoConfig := path.Join(p.Data.GitLocalPath, p.Config.GetString("engine_repo_config_path"))
	if utils.FileExists(repoConfig) {
		if err := p.Config.ReadConfig(repoConfig); err != nil {
			return errors.New("An error occured while parsing repository capsule.yml file")
		}
	} else {
		log.Println("No repo capsule.yml file found, using existing config.")
	}

	if p.Config.IsSet("scm_release_assets") {
		//unmarshall config data.
		parsedAssets := new([]pipeline.ScmReleaseAsset)
		if err := p.Config.UnmarshalKey("scm_release_assets", parsedAssets); err != nil {
			return err
		}

		//append the parsed Assets to the current ReleaseAssets storage (incase assets were defined in system yml)
		p.Data.ReleaseAssets = append(p.Data.ReleaseAssets, (*parsedAssets)...)
	}
	return nil
}

func (p *Pipeline) MgrInitStep() error {
	log.Println("mgr_init_step")
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
	return nil
}

func (p *Pipeline) ValidateTools() error {
	log.Println("validate_tools")
	return p.Engine.ValidateTools()
}

func (p *Pipeline) MgrValidateTools() error {
	log.Println("mgr_validate_tools")
	return p.PackageManager.MgrValidateTools()
}

// this step should commit any local changes and create a git tag. It should also generate the releaser artifacts. Nothing should be pushed to remote repository
func (p *Pipeline) PackageStep() error {

	if p.Config.IsSet("package_step.override") {
		log.Println("Cannot override the package_step, ignoring.")
	}
	log.Println("mgr_package_step")
	if err := p.PackageManager.MgrPackageStep(p.Engine.GetNextMetadata()); err != nil {
		return err
	}
	log.Println("package_step")
	if err := p.Engine.PackageStep(); err != nil {
		return err
	}

	return nil
}
