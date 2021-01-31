// +build golang

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
type MgrGolangModTestSuite struct {
	suite.Suite
	MockCtrl     *gomock.Controller
	Mgr          *mock_mgr.MockInterface
	Config       *mock_config.MockInterface
	PipelineData *pipeline.Data
}

// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (suite *MgrGolangModTestSuite) SetupTest() {
	suite.MockCtrl = gomock.NewController(suite.T())

	suite.PipelineData = new(pipeline.Data)

	suite.Config = mock_config.NewMockInterface(suite.MockCtrl)
	suite.Mgr = mock_mgr.NewMockInterface(suite.MockCtrl)

}

func (suite *MgrGolangModTestSuite) TearDownTest() {
	suite.MockCtrl.Finish()
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestMgrGolangMod_TestSuite(t *testing.T) {
	suite.Run(t, new(MgrGolangModTestSuite))
}

func (suite *MgrGolangModTestSuite) TestMgrGolangModTestSuite_MgrDistStep_WithoutCredentials() {
	//setup
	//suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	mgrGolangMod, err := mgr.Create("mod", suite.PipelineData, suite.Config, nil)
	require.NoError(suite.T(), err)
	nextVersion := new(metadata.GolangMetadata)

	//test
	berr := mgrGolangMod.MgrDistStep(nextVersion)

	//assert
	require.NoError(suite.T(), berr)
}
