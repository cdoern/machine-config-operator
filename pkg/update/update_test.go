package update

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTruncate(t *testing.T) {
	assert.Equal(t, truncate("", 10), "")
	assert.Equal(t, truncate("", 1), "")
	assert.Equal(t, truncate("a", 1), "a")
	assert.Equal(t, truncate("abcde", 1), "a [4 more chars]")
	assert.Equal(t, truncate("abcde", 4), "abcd [1 more chars]")
	assert.Equal(t, truncate("abcde", 7), "abcde")
	assert.Equal(t, truncate("abcde", 5), "abcde")
}

func TestRunGetOut(t *testing.T) {
	o, err := runGetOut("true")
	assert.Nil(t, err)
	assert.Equal(t, len(o), 0)

	o, err = runGetOut("false")
	assert.NotNil(t, err)

	o, err = runGetOut("echo", "hello")
	assert.Nil(t, err)
	assert.Equal(t, string(o), "hello\n")

	// base64 encode "oops" so we can't match on the command arguments
	o, err = runGetOut("/bin/sh", "-c", "echo hello; echo b29wcwo= | base64 -d 1>&2; exit 1")
	assert.Error(t, err)
	errtext := err.Error()
	assert.Contains(t, errtext, "exit status 1\noops\n")

	o, err = runGetOut("/usr/bin/test-failure-to-exec-this-should-not-exist", "arg")
	assert.Error(t, err)
}
