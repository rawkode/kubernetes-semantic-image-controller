package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldResolve(t *testing.T) {
	d := &dockerHubResolver{}

	// Things that should resolve
	assert.True(t, d.ShouldResolve("nginx:>1.15"))
	assert.True(t, d.ShouldResolve("nginx:=1.15"))
	assert.True(t, d.ShouldResolve("nginx:<1.15"))
	assert.True(t, d.ShouldResolve("nginx:>1.15,<1.16"))

	// Things that shouldn't resolve
	assert.False(t, d.ShouldResolve("nginx:1.15.7"))
	assert.False(t, d.ShouldResolve("nginx"))
	assert.False(t, d.ShouldResolve("nginx:"))
	assert.False(t, d.ShouldResolve("nginx:foo"))
	assert.False(t, d.ShouldResolve(""))
}
