package semver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleASCIIScanner(t *testing.T) {
	t.Run("empty string", func(t *testing.T) {
		s := newSimpleASCIIScanner("")
		assert.True(t, s.eof())
		assert.Equal(t, scannerEOF, s.peek())
		assert.Equal(t, scannerEOF, s.next())

		substr, term := s.readUntil(noTerminator)
		assert.True(t, s.eof())
		assert.Equal(t, "", substr)
		assert.Equal(t, scannerEOF, term)
	})

	t.Run("peek/next", func(t *testing.T) {
		s := newSimpleASCIIScanner("ab")

		assert.False(t, s.eof())
		assert.Equal(t, int8('a'), s.peek())
		assert.Equal(t, int8('a'), s.peek())
		assert.Equal(t, int8('a'), s.next())

		assert.False(t, s.eof())
		assert.Equal(t, int8('b'), s.peek())
		assert.Equal(t, int8('b'), s.peek())
		assert.Equal(t, int8('b'), s.next())

		assert.True(t, s.eof())
		assert.Equal(t, scannerEOF, s.peek())
		assert.Equal(t, scannerEOF, s.peek())
		assert.Equal(t, scannerEOF, s.next())
		assert.Equal(t, scannerEOF, s.next())
	})

	t.Run("readUntil", func(t *testing.T) {
		s1 := newSimpleASCIIScanner("abcd")
		ss1, term1 := s1.readUntil(noTerminator)
		assert.Equal(t, "abcd", ss1)
		assert.Equal(t, scannerEOF, term1)
		assert.Equal(t, scannerEOF, s1.peek())

		s2 := newSimpleASCIIScanner("abcd")
		_ = s2.next()
		ss2, term2 := s2.readUntil(noTerminator)
		assert.Equal(t, "bcd", ss2)
		assert.Equal(t, scannerEOF, term2)
		assert.Equal(t, scannerEOF, s2.peek())

		s3 := newSimpleASCIIScanner("abcd")
		_ = s3.next()
		ss3, term3 := s3.readUntil(func(ch rune) bool { return ch == 'c' })
		assert.Equal(t, "b", ss3)
		assert.Equal(t, int8('c'), term3)
		assert.Equal(t, int8('d'), s3.peek())
	})

	t.Run("halts on non-ASCII character", func(t *testing.T) {
		s := newSimpleASCIIScanner("a🥦b")
		ss, term := s.readUntil(noTerminator)
		assert.Equal(t, scannerNonASCII, term)
		assert.Equal(t, "a", ss)
	})
}
