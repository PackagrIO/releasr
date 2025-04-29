//go:build generic
// +build generic

package engine_test

import (
	"github.com/analogj/go-util/utils"
	"github.com/golang/mock/gomock"
	"github.com/packagrio/go-common/metadata"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/go-common/scm"
	"github.com/packagrio/go-common/scm/mock"
	"github.com/packagrio/releasr/pkg/config"
	"github.com/packagrio/releasr/pkg/engine"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"testing"
)

func TestEngineGeneric_Create(t *testing.T) {
	//setup
	testConfig, err := config.Create()
	require.NoError(t, err)
	testConfig.Set(config.PACKAGR_SCM, "github")
	testConfig.Set(config.PACKAGR_PACKAGE_TYPE, "generic")
	pipelineData := new(pipeline.Data)
	githubScm, err := scm.Create("github", pipelineData, testConfig, &http.Client{})
	require.NoError(t, err)

	//test
	genericEngine, err := engine.Create(engine.PACKAGR_ENGINE_TYPE_GENERIC, pipelineData, testConfig, githubScm)

	println("genericEngine", genericEngine)

	//assert
	require.NoError(t, err)
	require.NotNil(t, genericEngine)
}

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type EngineGenericTestSuite struct {
	suite.Suite
	MockCtrl     *gomock.Controller
	Scm          *mock_scm.MockInterface
	Config       config.Interface
	PipelineData *pipeline.Data
}

// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (suite *EngineGenericTestSuite) SetupTest() {
	suite.MockCtrl = gomock.NewController(suite.T())

	suite.PipelineData = new(pipeline.Data)

	testConfig, err := config.Create()
	require.NoError(suite.T(), err)
	testConfig.Set(config.PACKAGR_SCM, "github")
	testConfig.Set(config.PACKAGR_PACKAGE_TYPE, "generic")
	suite.Config = testConfig
	suite.Scm = mock_scm.NewMockInterface(suite.MockCtrl)

}

func (suite *EngineGenericTestSuite) TearDownTest() {
	suite.MockCtrl.Finish()
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestEngineGeneric_TestSuite(t *testing.T) {
	suite.Run(t, new(EngineGenericTestSuite))
}

func (suite *EngineGenericTestSuite) TestEngineGeneric_ValidateTools() {
	//setup

	genericEngine, err := engine.Create(engine.PACKAGR_ENGINE_TYPE_GENERIC, suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := genericEngine.ValidateTools()

	//assert
	require.NoError(suite.T(), berr)
}

func (suite *EngineGenericTestSuite) TestEngineGeneric_GetVersion() {
	//copy into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "generic_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "generic", "generic_analogj_test"), suite.PipelineData.GitLocalPath)
	require.NoError(suite.T(), cerr)

	genericEngine, err := engine.Create(engine.PACKAGR_ENGINE_TYPE_GENERIC, suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := genericEngine.PopulateReleaseVersion()
	require.NoError(suite.T(), berr)

	//assert
	require.Equal(suite.T(), "0.0.1", genericEngine.GetNextMetadata().(*metadata.GenericMetadata).Version)

}

func (suite *EngineGenericTestSuite) TestEngineGeneric_GetVersion_Metadata() {
	//setup
	suite.Config.Set(config.PACKAGR_VERSION_METADATA_PATH, "version.txt")
	//copy into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "generic_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "generic", "generic_metadata_analogj_test"), suite.PipelineData.GitLocalPath)
	require.NoError(suite.T(), cerr)

	genericEngine, err := engine.Create(engine.PACKAGR_ENGINE_TYPE_GENERIC, suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := genericEngine.PopulateReleaseVersion()
	require.NoError(suite.T(), berr)

	//assert
	require.Equal(suite.T(), "0.0.1", genericEngine.GetNextMetadata().(*metadata.GenericMetadata).Version)

}

func (suite *EngineGenericTestSuite) TestEngineGeneric_GetVersion_Merge() {
	//setup
	suite.Config.Set(config.PACKAGR_GENERIC_MERGE_VERSION_FILE, "true")
	//copy into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "generic_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "generic", "generic_merge_analogj_test"), suite.PipelineData.GitLocalPath)
	require.NoError(suite.T(), cerr)

	genericEngine, err := engine.Create(engine.PACKAGR_ENGINE_TYPE_GENERIC, suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := genericEngine.PopulateReleaseVersion()
	require.NoError(suite.T(), berr)

	//assert
	require.Equal(suite.T(), "0.0.1", genericEngine.GetNextMetadata().(*metadata.GenericMetadata).Version)

}

func (suite *EngineGenericTestSuite) TestEngineGeneric_GetVersion_Template() {
	//setup
	suite.Config.Set(config.PACKAGR_GENERIC_VERSION_TEMPLATE, "%d.%d.%d")
	//copy into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "generic_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "generic", "generic_template_analogj_test"), suite.PipelineData.GitLocalPath)
	require.NoError(suite.T(), cerr)

	genericEngine, err := engine.Create(engine.PACKAGR_ENGINE_TYPE_GENERIC, suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := genericEngine.PopulateReleaseVersion()
	require.NoError(suite.T(), berr)

	//assert
	require.Equal(suite.T(), "0.0.1", genericEngine.GetNextMetadata().(*metadata.GenericMetadata).Version)

}
