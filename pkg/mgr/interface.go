package mgr

import (
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/releasr/pkg/config"
	"net/http"
)

// Create mock using:
// mockgen -source=pkg/mgr/interface.go -destination=pkg/mgr/mock/mock_mgr.go
type Interface interface {
	Init(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) error

	// Validate that required executables are available for the following build/test/package/etc steps
	MgrValidateTools() error

	// Commit any local changes and create a git tag. Nothing should be pushed to remote repository yet.
	// Make sure you remove any unnecessary files from the repo before making the commit
	// CAN NOT override
	// REQUIRES pipelineData.GitLocalPath
	// REQUIRES NextMetadata
	// USES mgr_keep_lock_file
	MgrPackageStep(nextMetadata interface{}) error

	// Push the release to the package repository (ie. npm, chef supermarket, rubygems)
	// Should validate any required credentials are specified.
	// CAN override
	// REQUIRES pipelineData.GitLocalPath
	// REQUIRES NextMetadata
	// USES chef_supermarket_username
	// USES chef_supermarket_key
	// USES npm_auth_token
	// USES pypi_repository
	// USES pypi_username
	// USES pypi_password
	// USES rubygems_api_key
	MgrDistStep(nextMetadata interface{}) error
}
