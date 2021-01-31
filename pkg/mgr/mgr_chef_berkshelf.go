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
	if !m.Config.GetBool("mgr_keep_lock_file") {
		os.Remove(path.Join(m.PipelineData.GitLocalPath, "Berksfile.lock"))
		os.Remove(path.Join(m.PipelineData.GitLocalPath, "Gemfile.lock"))
	}
	return nil
}

func (m *mgrChefBerkshelf) MgrDistStep(nextMetadata interface{}) error {
	if !m.Config.IsSet("chef_supermarket_username") || !m.Config.IsSet("chef_supermarket_key") {
		return errors.MgrDistCredentialsMissing("Cannot deploy cookbook to supermarket, credentials missing")
	}

	// knife is really sensitive to folder names. The cookbook name MUST match the folder name otherwise knife throws up
	// when doing a knife cookbook share. So we're going to make a new tmp directory, create a subdirectory with the EXACT
	// cookbook name, and then copy the cookbook contents into it. Yeah yeah, its pretty nasty, but blame Chef.
	tmpParentPath, terr := ioutil.TempDir("", "")
	if terr != nil {
		return terr
	}
	defer os.RemoveAll(tmpParentPath)

	tmpLocalPath := path.Join(tmpParentPath, nextMetadata.(*metadata.ChefMetadata).Name)
	if cerr := utils.CopyDir(m.PipelineData.GitLocalPath, tmpLocalPath); cerr != nil {
		return cerr
	}

	pemFile, _ := ioutil.TempFile("", "client.pem")
	defer os.Remove(pemFile.Name())
	knifeFile, _ := ioutil.TempFile("", "knife.rb")
	defer os.Remove(knifeFile.Name())

	// write the knife.rb config jfile.
	knifeContent := fmt.Sprintf(utils.StripIndent(
		`node_name "%s" # Replace with the login name you use to login to the Supermarket.
    		client_key "%s" # Define the path to wherever your client.pem file lives.  This is the key you generated when you signed up for a Chef account.
        	cookbook_path [ '%s' ] # Directory where the cookbook you're uploading resides.
		`),
		m.Config.GetString("chef_supermarket_username"),
		pemFile.Name(),
		tmpParentPath,
	)

	_, kerr := knifeFile.Write([]byte(knifeContent))
	if kerr != nil {
		return kerr
	}

	chefKey, berr := m.Config.GetBase64Decoded("chef_supermarket_key")
	if berr != nil {
		return berr
	}
	_, perr := pemFile.Write([]byte(chefKey))
	if perr != nil {
		return perr
	}

	cookbookDistCmd := fmt.Sprintf("knife cookbook site share %s %s -c %s",
		nextMetadata.(*metadata.ChefMetadata).Name,
		m.Config.GetString("chef_supermarket_type"),
		knifeFile.Name(),
	)

	if derr := utils.BashCmdExec(cookbookDistCmd, "", nil, ""); derr != nil {
		return errors.MgrDistPackageError("knife cookbook upload to supermarket failed")
	}
	return nil
}
