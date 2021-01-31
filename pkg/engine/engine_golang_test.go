// +build golang

package engine_test

import (
	"github.com/analogj/go-util/utils"
	"github.com/golang/mock/gomock"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/go-common/scm"
	"github.com/packagrio/releasr/pkg/config"
	"github.com/packagrio/releasr/pkg/engine"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"path"
	//"path/filepath"
	"github.com/packagrio/go-common/scm/mock"
	"github.com/packagrio/releasr/pkg/config/mock"
	"os"
	"testing"
)

func TestEngineGolang_Create(t *testing.T) {
	//setup
	testConfig, err := config.Create()
	require.NoError(t, err)

	testConfig.Set(config.PACKAGR_SCM, "github")
	testConfig.Set(config.PACKAGR_PACKAGE_TYPE, "golang")
	pipelineData := new(pipeline.Data)
	githubScm, err := scm.Create("github", pipelineData)
	require.NoError(t, err)

	//test
	golangEngine, err := engine.Create(engine.PACKAGR_ENGINE_TYPE_GOLANG, pipelineData, testConfig, githubScm)

	//assert
	require.NoError(t, err)
	require.NotNil(t, golangEngine)
}

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type EngineGolangTestSuite struct {
	suite.Suite
	MockCtrl     *gomock.Controller
	Scm          *mock_scm.MockInterface
	Config       *mock_config.MockInterface
	PipelineData *pipeline.Data
}

// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (suite *EngineGolangTestSuite) SetupTest() {
	suite.MockCtrl = gomock.NewController(suite.T())

	suite.PipelineData = new(pipeline.Data)

	suite.Config = mock_config.NewMockInterface(suite.MockCtrl)
	suite.Scm = mock_scm.NewMockInterface(suite.MockCtrl)

}

func (suite *EngineGolangTestSuite) TearDownTest() {
	suite.MockCtrl.Finish()
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestEngineGolang_TestSuite(t *testing.T) {
	suite.Run(t, new(EngineGolangTestSuite))
}

func (suite *EngineGolangTestSuite) TestEngineGolang_ValidateTools() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetString(config.PACKAGR_SCM).Return("github").MinTimes(1)
	suite.Config.EXPECT().GetString("scm_repo_full_name").Return("AnalogJ/golang_analogj_test").MinTimes(1)
	suite.Config.EXPECT().GetString("engine_golang_package_path").Return("github.com/analogj/golang_analogj_test").MinTimes(1)

	golangEngine, err := engine.Create(engine.PACKAGR_ENGINE_TYPE_GOLANG, suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := golangEngine.ValidateTools()

	//assert
	require.NoError(suite.T(), berr)
}

func (suite *EngineGolangTestSuite) TestEngineGolang_PackageStep_WithoutLockFiles() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetString("scm").Return("github").MinTimes(1)
	suite.Config.EXPECT().GetString("scm_repo_full_name").Return("AnalogJ/golang_analogj_test").MinTimes(1)
	suite.Config.EXPECT().GetString("engine_golang_package_path").Return("github.com/analogj/golang_analogj_test").MinTimes(1)
	suite.Config.EXPECT().GetString("engine_version_bump_msg").Return("Automated packaging of release by CapsuleCD").MinTimes(1)
	suite.Config.EXPECT().GetString("engine_git_author_name").Return("CapsuleCD").MinTimes(1)
	suite.Config.EXPECT().GetString("engine_git_author_email").Return("CapsuleCD@users.noreply.github.com").MinTimes(1)

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	cpath, cerr := utils.GitClone(parentPath, "golang_analogj_test", "https://github.com/AnalogJ/golang_analogj_test.git")
	require.NoError(suite.T(), cerr)
	suite.PipelineData.GitLocalPath = cpath

	golangEngine, err := engine.Create("golang", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := golangEngine.PackageStep()

	//assert
	require.NoError(suite.T(), berr)
	require.False(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, "glide.lock")))
}
