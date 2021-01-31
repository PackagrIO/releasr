// +build ruby

package mgr_test

import (
	"github.com/golang/mock/gomock"
	"github.com/packagrio/go-common/metadata"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/releasr/pkg/config/mock"
	"github.com/packagrio/releasr/pkg/mgr"
	"github.com/packagrio/releasr/pkg/mgr/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type MgrRubyBundlerTestSuite struct {
	suite.Suite
	MockCtrl     *gomock.Controller
	Mgr          *mock_mgr.MockInterface
	Config       *mock_config.MockInterface
	PipelineData *pipeline.Data
}

// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (suite *MgrRubyBundlerTestSuite) SetupTest() {
	suite.MockCtrl = gomock.NewController(suite.T())

	suite.PipelineData = new(pipeline.Data)

	suite.Config = mock_config.NewMockInterface(suite.MockCtrl)
	suite.Mgr = mock_mgr.NewMockInterface(suite.MockCtrl)

}

func (suite *MgrRubyBundlerTestSuite) TearDownTest() {
	suite.MockCtrl.Finish()
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestMgrRubyBundler_TestSuite(t *testing.T) {
	suite.Run(t, new(MgrRubyBundlerTestSuite))
}

func (suite *MgrRubyBundlerTestSuite) TestMgrRubyBundlerTestSuite_MgrDistStep_WithoutCredentials() {
	//setup
	//suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().IsSet("rubygems_api_key").MinTimes(1).Return(false)

	mgrRubyBundler, err := mgr.Create("bundler", suite.PipelineData, suite.Config, nil)
	require.NoError(suite.T(), err)
	nextVersion := new(metadata.RubyMetadata)

	//test
	berr := mgrRubyBundler.MgrDistStep(nextVersion)

	//assert
	require.Error(suite.T(), berr)
}
