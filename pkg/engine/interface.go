package engine

import (
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/go-common/scm"
	"github.com/packagrio/releasr/pkg/config"
)

// Create mock using:
// mockgen -source=pkg/engine/interface.go -destination=pkg/engine/mock/mock_engine.go
type Interface interface {
	Init(pipelineData *pipeline.Data, config config.Interface, sourceScm scm.Interface) error

	PopulateNextMetadata() error
	GetNextMetadata() interface{}

	// Validate that required executables are available for the following build/test/package/etc steps
	ValidateTools() error

	// Commit any local changes and create a git tag. Nothing should be pushed to remote repository yet.
	// Make sure you remove any unnecessary files from the repo before making the commit
	// CAN NOT override
	// MUST set ReleaseCommit
	// MUST set ReleaseVersion
	// REQUIRES pipelineData.GitLocalPath
	// REQUIRES NextMetadata
	// USES mgr_keep_lock_file
	PackageStep() error
}

const PACKAGR_ENGINE_TYPE_CHEF = "chef"
const PACKAGR_ENGINE_TYPE_GENERIC = "generic"
const PACKAGR_ENGINE_TYPE_GOLANG = "golang"
const PACKAGR_ENGINE_TYPE_NODE = "node"
const PACKAGR_ENGINE_TYPE_PYTHON = "python"
const PACKAGR_ENGINE_TYPE_RUBY = "ruby"
