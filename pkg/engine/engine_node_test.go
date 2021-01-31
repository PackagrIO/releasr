// +build node

package engine_test

import (
	"github.com/analogj/go-util/utils"
	"github.com/golang/mock/gomock"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/go-common/scm"
	"github.com/packagrio/releasr/pkg/config"
	"github.com/packagrio/releasr/pkg/engine"
	releasrUtils "github.com/packagrio/releasr/pkg/utils"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"path"
	//"path/filepath"
	mock_scm "github.com/packagrio/go-common/scm/mock"
	"github.com/packagrio/releasr/pkg/config/mock"
	"os"
	"testing"
)

func TestEngineNode_Create(t *testing.T) {
	//setup
	testConfig, err := config.Create()
	require.NoError(t, err)

	testConfig.Set(config.PACKAGR_SCM, "github")
	testConfig.Set(config.PACKAGR_PACKAGE_TYPE, "node")
	pipelineData := new(pipeline.Data)
	githubScm, err := scm.Create("github", pipelineData)
	require.NoError(t, err)

	//test
	nodeEngine, err := engine.Create(engine.PACKAGR_ENGINE_TYPE_NODE, pipelineData, testConfig, githubScm)

	//assert
	require.NoError(t, err)
	require.NotNil(t, nodeEngine)
}

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type EngineNodeTestSuite struct {
	suite.Suite
	MockCtrl     *gomock.Controller
	Scm          *mock_scm.MockInterface
	Config       *mock_config.MockInterface
	PipelineData *pipeline.Data
}

// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (suite *EngineNodeTestSuite) SetupTest() {
	suite.MockCtrl = gomock.NewController(suite.T())

	suite.PipelineData = new(pipeline.Data)

	suite.Config = mock_config.NewMockInterface(suite.MockCtrl)
	suite.Scm = mock_scm.NewMockInterface(suite.MockCtrl)

}

func (suite *EngineNodeTestSuite) TearDownTest() {
	suite.MockCtrl.Finish()
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestEngineNode_TestSuite(t *testing.T) {
	suite.Run(t, new(EngineNodeTestSuite))
}

func (suite *EngineNodeTestSuite) TestEngineNode_ValidateTools() {
	//setup
	//suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	//suite.Config.EXPECT().GetBool("engine_disable_security_check").Return(true).MinTimes(1)
	nodeEngine, err := engine.Create(engine.PACKAGR_ENGINE_TYPE_NODE, suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := nodeEngine.ValidateTools()

	//assert
	require.NoError(suite.T(), berr)
}

func (suite *EngineNodeTestSuite) TestEngineNode_PackageStep_WithoutLockFiles() {
	//setup
	//suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	//suite.Config.EXPECT().GetBool("mgr_keep_lock_file").Return(false).MinTimes(1)
	suite.Config.EXPECT().GetString(config.PACKAGR_VERSION_BUMP_MESSAGE).Return("Automated packaging of release by CapsuleCD").MinTimes(1)
	suite.Config.EXPECT().GetString(config.PACKAGR_GIT_AUTHOR_NAME).Return("CapsuleCD").MinTimes(1)
	suite.Config.EXPECT().GetString(config.PACKAGR_GIT_AUTHOR_EMAIL).Return("CapsuleCD@users.noreply.github.com").MinTimes(1)

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	cpath, cerr := releasrUtils.GitClone(parentPath, "npm_analogj_test", "https://github.com/AnalogJ/npm_analogj_test.git")
	require.NoError(suite.T(), cerr)
	suite.PipelineData.GitLocalPath = cpath

	nodeEngine, err := engine.Create("node", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := nodeEngine.PackageStep()

	//assert
	require.NoError(suite.T(), berr)
	require.False(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, "npm-shrinkwrap.json")))
}
