package ldtime

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUnixMillisFromTime(t *testing.T) {
	tt := time.Date(1970, time.January, 1, 0, 1, 2, 0, time.UTC)
	ut := UnixMillisFromTime(tt)
	assert.Equal(t, uint64(62000), uint64(ut))
}

func TestUnixMillisNow(t *testing.T) {
	tn := time.Now()
	un := UnixMillisNow()
	assert.GreaterOrEqual(t, uint64(un), uint64(UnixMillisFromTime(tn)))
}
