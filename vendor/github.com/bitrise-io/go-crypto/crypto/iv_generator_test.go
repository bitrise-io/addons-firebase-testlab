package crypto_test

import (
	"testing"

	"github.com/bitrise-team/bitrise-api/utils"
	"github.com/stretchr/testify/require"
)

func Test_GenerateIV(t *testing.T) {
	t.Log("test")
	{
		generatedIV, err := utils.GenerateIV()
		require.NoError(t, err)
		require.Equal(t, int(12), len(generatedIV))
	}
}
