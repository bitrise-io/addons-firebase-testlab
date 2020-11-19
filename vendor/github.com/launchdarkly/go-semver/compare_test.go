package semver

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// The test data set is based on the one defined in github.com/blang/semver (see:
// https://github.com/blang/semver/blob/master/semver_test.go), not including data for any
// non-standard parsing modes or range comparisons.

type compareTest struct {
	v1     Version
	v2     Version
	result int
}

var compareTests = []compareTest{
	{Version{1, 0, 0, "", ""}, Version{1, 0, 0, "", ""}, 0},
	{Version{2, 0, 0, "", ""}, Version{1, 0, 0, "", ""}, 1},
	{Version{0, 1, 0, "", ""}, Version{0, 1, 0, "", ""}, 0},
	{Version{0, 2, 0, "", ""}, Version{0, 1, 0, "", ""}, 1},
	{Version{0, 0, 1, "", ""}, Version{0, 0, 1, "", ""}, 0},
	{Version{0, 0, 2, "", ""}, Version{0, 0, 1, "", ""}, 1},
	{Version{1, 2, 3, "", ""}, Version{1, 2, 3, "", ""}, 0},
	{Version{2, 2, 4, "", ""}, Version{1, 2, 4, "", ""}, 1},
	{Version{1, 3, 3, "", ""}, Version{1, 2, 3, "", ""}, 1},
	{Version{1, 2, 4, "", ""}, Version{1, 2, 3, "", ""}, 1},

	{Version{1, 0, 0, "", ""}, Version{2, 0, 0, "", ""}, -1},
	{Version{2, 0, 0, "", ""}, Version{2, 1, 0, "", ""}, -1},
	{Version{2, 1, 0, "", ""}, Version{2, 1, 1, "", ""}, -1},

	{Version{1, 0, 0, "alpha", ""}, Version{1, 0, 0, "alpha", ""}, 0},
	{Version{1, 0, 0, "", ""}, Version{1, 0, 0, "alpha", ""}, 1},
	{Version{1, 0, 0, "alpha", ""}, Version{1, 0, 0, "alpha.1", ""}, -1},
	{Version{1, 0, 0, "alpha.1", ""}, Version{1, 0, 0, "alpha.beta", ""}, -1},
	{Version{1, 0, 0, "alpha.beta", ""}, Version{1, 0, 0, "beta", ""}, -1},
	{Version{1, 0, 0, "beta", ""}, Version{1, 0, 0, "beta.2", ""}, -1},
	{Version{1, 0, 0, "beta.2", ""}, Version{1, 0, 0, "beta.11", ""}, -1},
	{Version{1, 0, 0, "beta.2", ""}, Version{1, 0, 0, "rc.1", ""}, -1},
	{Version{1, 0, 0, "rc.1", ""}, Version{1, 0, 0, "", ""}, -1},

	{Version{1, 0, 0, "", "1.2.3"}, Version{1, 0, 0, "", ""}, 0},
}

func TestSemanticVersionCompare(t *testing.T) {
	for _, test := range compareTests {
		opDesc := "=="
		if test.result < 0 {
			opDesc = "<"
		} else if test.result > 0 {
			opDesc = ">"
		}
		t.Run(fmt.Sprintf("%+v should %s %+v", test.v1, opDesc, test.v2), func(t *testing.T) {
			assert.Equal(t, test.result, test.v1.ComparePrecedence(test.v2))
			if test.result != 0 {
				assert.Equal(t, -test.result, test.v2.ComparePrecedence(test.v1),
					"%+v should not %s %+v", test.v2, opDesc, test.v1)
			}
		})
	}
}
