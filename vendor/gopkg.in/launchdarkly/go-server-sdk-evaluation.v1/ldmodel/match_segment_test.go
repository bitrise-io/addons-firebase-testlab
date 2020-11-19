package ldmodel

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSegmentIncludesUser(t *testing.T) {
	for _, withPreprocessing := range []bool{false, true} {
		t.Run(fmt.Sprintf("preprocessed: %t", withPreprocessing), func(t *testing.T) {
			segment := Segment{Included: []string{"a", "b", "c"}}
			if withPreprocessing {
				PreprocessSegment(&segment)
			}
			included, found := SegmentIncludesOrExcludesKey(&segment, "b")
			assert.True(t, included)
			assert.True(t, found)
		})
	}
}

func TestSegmentExcludesUser(t *testing.T) {
	for _, withPreprocessing := range []bool{false, true} {
		t.Run(fmt.Sprintf("preprocessed: %t", withPreprocessing), func(t *testing.T) {
			segment := Segment{Excluded: []string{"a", "b", "c"}}
			if withPreprocessing {
				PreprocessSegment(&segment)
			}
			included, found := SegmentIncludesOrExcludesKey(&segment, "b")
			assert.False(t, included)
			assert.True(t, found)
		})
	}
}

func TestSegmentBothIncludesAndExcludesUser(t *testing.T) {
	for _, withPreprocessing := range []bool{false, true} {
		t.Run(fmt.Sprintf("preprocessed: %t", withPreprocessing), func(t *testing.T) {
			segment := Segment{Included: []string{"a", "b", "c"}, Excluded: []string{"b"}}
			if withPreprocessing {
				PreprocessSegment(&segment)
			}
			included, found := SegmentIncludesOrExcludesKey(&segment, "b")
			assert.True(t, included) // include takes priority over exclude
			assert.True(t, found)
		})
	}
}

func TestSegmentNeitherIncludesNorExcludesUser(t *testing.T) {
	for _, withPreprocessing := range []bool{false, true} {
		t.Run(fmt.Sprintf("preprocessed: %t", withPreprocessing), func(t *testing.T) {
			segment := Segment{Excluded: []string{"a", "b", "c"}}
			if withPreprocessing {
				PreprocessSegment(&segment)
			}
			included, found := SegmentIncludesOrExcludesKey(&segment, "d")
			assert.False(t, included)
			assert.False(t, found)
		})
	}
}
