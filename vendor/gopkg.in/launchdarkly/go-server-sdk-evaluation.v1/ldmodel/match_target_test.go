package ldmodel

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTargetMatchesUser(t *testing.T) {
	for _, withPreprocessing := range []bool{false, true} {
		t.Run(fmt.Sprintf("preprocessed: %t", withPreprocessing), func(t *testing.T) {
			target := Target{Values: []string{"a", "b", "c"}}
			if withPreprocessing {
				target.preprocessed = preprocessTarget(target)
			}
			assert.True(t, TargetContainsKey(&target, "b"))
		})
	}
}

func TestTargetDoesNotMatchUser(t *testing.T) {
	for _, withPreprocessing := range []bool{false, true} {
		t.Run(fmt.Sprintf("preprocessed: %t", withPreprocessing), func(t *testing.T) {
			target := Target{Values: []string{"a", "b", "c"}}
			if withPreprocessing {
				target.preprocessed = preprocessTarget(target)
			}
			assert.False(t, TargetContainsKey(&target, "d"))
		})
	}
}
