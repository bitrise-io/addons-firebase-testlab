package semver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertVersionComponents(t *testing.T, v Version, major int, minor int, patch int, pre string, b string) {
	assert.Equal(t, major, v.GetMajor())
	assert.Equal(t, minor, v.GetMinor())
	assert.Equal(t, patch, v.GetPatch())
	assert.Equal(t, pre, v.GetPrerelease())
	assert.Equal(t, b, v.GetBuild())
}

func parsingShouldSucceed(parseFn func(string) (Version, error), s string,
	major int, minor int, patch int, pre string, b string) func(t *testing.T) {
	return func(t *testing.T) {
		v, err := parseFn(s)
		require.NoError(t, err)
		assertVersionComponents(t, v, major, minor, patch, pre, b)
	}
}

func parsingShouldFail(parseFn func(string) (Version, error), s string) func(t *testing.T) {
	return func(t *testing.T) {
		v, err := parseFn(s)
		assert.Error(t, err)
		assert.Equal(t, Version{}, v)
	}
}

func TestParseStrict(t *testing.T) {
	strictParsingTests := func(t *testing.T, parseFn func(string) (Version, error)) {
		parsingTestsForAnyMode(t, parseFn)

		t.Run("patch version cannot be omitted", parsingShouldFail(parseFn, "2.3"))
		t.Run("micro and patch versions cannot be omitted", parsingShouldFail(parseFn, "2"))
	}

	t.Run("Parse(s)", func(t *testing.T) {
		strictParsingTests(t, Parse)
	})

	t.Run("ParseAs(s, ParseModeStrict)", func(t *testing.T) {
		strictParsingTests(t, func(s string) (Version, error) { return ParseAs(s, ParseModeStrict) })
	})
}

func TestParseAllowMissingMinorAndPatch(t *testing.T) {
	parseFn := func(s string) (Version, error) { return ParseAs(s, ParseModeAllowMissingMinorAndPatch) }

	parsingTestsForAnyMode(t, parseFn)

	t.Run("patch version can be omitted", parsingShouldSucceed(parseFn,
		"2.3", 2, 3, 0, "", ""))

	t.Run("patch version can be omitted with prerelease", parsingShouldSucceed(parseFn,
		"2.3-beta1", 2, 3, 0, "beta1", ""))

	t.Run("patch version can be omitted with build", parsingShouldSucceed(parseFn,
		"2.3+build1", 2, 3, 0, "", "build1"))

	t.Run("micro and patch versions can be omitted", parsingShouldSucceed(parseFn,
		"2", 2, 0, 0, "", ""))

	t.Run("micro and patch versions can be omitted with prerelease", parsingShouldSucceed(parseFn,
		"2-beta1", 2, 0, 0, "beta1", ""))

	t.Run("micro and patch versions can be omitted with build", parsingShouldSucceed(parseFn,
		"2+build1", 2, 0, 0, "", "build1"))
}

func parsingTestsForAnyMode(t *testing.T, parseFn func(string) (Version, error)) {
	t.Run("valid", func(t *testing.T) {
		t.Run("simple complete version", parsingShouldSucceed(parseFn,
			"2.3.4", 2, 3, 4, "", ""))

		t.Run("version with prerelease (single identifier)", parsingShouldSucceed(parseFn,
			"2.3.4-beta1", 2, 3, 4, "beta1", ""))

		t.Run("version with prerelease (multi-identifier)", parsingShouldSucceed(parseFn,
			"2.3.4-beta1.2", 2, 3, 4, "beta1.2", ""))

		t.Run("version with prerelease (multi-identifier with hyphens)", parsingShouldSucceed(parseFn,
			"2.3.4-beta1-final.2", 2, 3, 4, "beta1-final.2", ""))

		t.Run("version with build (single identifier)", parsingShouldSucceed(parseFn,
			"2.3.4+build2", 2, 3, 4, "", "build2"))

		t.Run("version with build (multi-identifier)", parsingShouldSucceed(parseFn,
			"2.3.4+build2.4", 2, 3, 4, "", "build2.4"))

		t.Run("version with build (multi-identifier with hyphens)", parsingShouldSucceed(parseFn,
			"2.3.4+build2-other.4", 2, 3, 4, "", "build2-other.4"))

		t.Run("version with prerelease and build", parsingShouldSucceed(parseFn,
			"2.3.4-beta1.rc2+build2.4", 2, 3, 4, "beta1.rc2", "build2.4"))

		t.Run("major version zero", parsingShouldSucceed(parseFn,
			"0.3.4", 0, 3, 4, "", ""))

		t.Run("minor version zero", parsingShouldSucceed(parseFn,
			"2.0.4", 2, 0, 4, "", ""))

		t.Run("patch version zero", parsingShouldSucceed(parseFn,
			"2.3.0", 2, 3, 0, "", ""))

		t.Run("prerelease identifier of only letters", parsingShouldSucceed(parseFn,
			"2.3.4-abc", 2, 3, 4, "abc", ""))

		t.Run("prerelease identifier of only digits", parsingShouldSucceed(parseFn,
			"2.3.4-123", 2, 3, 4, "123", ""))

		t.Run("prerelease identifier of only hyphens", parsingShouldSucceed(parseFn,
			"2.3.4----", 2, 3, 4, "---", ""))

		t.Run("alphanumeric prerelease identifier with leading zero", parsingShouldSucceed(parseFn,
			"2.3.4-beta1.0yes", 2, 3, 4, "beta1.0yes", ""))

		t.Run("build identifier of only letters", parsingShouldSucceed(parseFn,
			"2.3.4+abc", 2, 3, 4, "", "abc"))

		t.Run("build identifier of only digits", parsingShouldSucceed(parseFn,
			"2.3.4+123", 2, 3, 4, "", "123"))

		t.Run("build identifier of only hyphens", parsingShouldSucceed(parseFn,
			"2.3.4+---", 2, 3, 4, "", "---"))

		t.Run("alphanumeric build identifier with leading zero", parsingShouldSucceed(parseFn,
			"2.3.4+build1.0yes", 2, 3, 4, "", "build1.0yes"))

		t.Run("leading zero in numeric build identifier", parsingShouldSucceed(parseFn,
			"2.3.4+build1.02", 2, 3, 4, "", "build1.02"))
	})

	t.Run("invalid", func(t *testing.T) {
		t.Run("non-numeric major", parsingShouldFail(parseFn, "2x.3.4"))
		t.Run("non-numeric minor", parsingShouldFail(parseFn, "2.3x.4"))
		t.Run("non-numeric patch", parsingShouldFail(parseFn, "2.3.4x"))
		t.Run("empty major", parsingShouldFail(parseFn, ".3.4"))
		t.Run("empty minor", parsingShouldFail(parseFn, "2..4"))
		t.Run("empty patch", parsingShouldFail(parseFn, "2.3."))
		t.Run("empty prerelease", parsingShouldFail(parseFn, "2.3.4-"))
		t.Run("empty prerelease identifier", parsingShouldFail(parseFn, "2.3.4-a..b"))
		t.Run("empty build", parsingShouldFail(parseFn, "2.3.4+a..b"))
		t.Run("leading zero in major", parsingShouldFail(parseFn, "02.3.4"))
		t.Run("leading zero in minor", parsingShouldFail(parseFn, "2.03.4"))
		t.Run("leading zero in patch", parsingShouldFail(parseFn, "2.3.04"))
		t.Run("leading zero in prerelease numeric identifier", parsingShouldFail(parseFn, "2.3.4-beta1.02"))
		t.Run("non-alphanumeric prerelease identifier", parsingShouldFail(parseFn, "2.3.4-beta1!"))
		t.Run("non-alphanumeric build identifier", parsingShouldFail(parseFn, "2.3.4+build!"))
		t.Run("non-ASCII character before major", parsingShouldFail(parseFn, "🔥2.3.4"))
		t.Run("non-ASCII character before minor", parsingShouldFail(parseFn, "2.🔥3.4"))
		t.Run("non-ASCII character before patch", parsingShouldFail(parseFn, "2.3.🔥4"))
		t.Run("non-ASCII character after major", parsingShouldFail(parseFn, "2🔥.3.4"))
		t.Run("non-ASCII character after minor", parsingShouldFail(parseFn, "2.3🔥.4"))
		t.Run("non-ASCII character after patch", parsingShouldFail(parseFn, "2.3.4🔥"))
		t.Run("non-ASCII character in prerelease", parsingShouldFail(parseFn, "2.3.4-beta🔥no"))
		t.Run("non-ASCII character in build", parsingShouldFail(parseFn, "2.3.4+build🔥no"))
	})
}

func TestParseAsUnknownMode(t *testing.T) {
	v, err := ParseAs("1.2.3", ParseMode(2))
	assert.Error(t, err)
	assert.Equal(t, Version{}, v)
}
