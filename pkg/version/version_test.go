package version_test

import (
	"github.com/Masterminds/semver"
	"github.com/packagrio/releasr/pkg/version"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestVersion(t *testing.T) {
	t.Parallel()

	//test
	v, nerr := semver.NewVersion(version.VERSION)

	//assert
	require.NoError(t, nerr, "should be a valid semver")
	require.Equal(t, version.VERSION, v.String())
}
