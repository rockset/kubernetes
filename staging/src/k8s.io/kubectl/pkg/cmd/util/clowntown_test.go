package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProdContext(t *testing.T) {
	tests := []struct {
		context  string
		expected bool
	}{
		{"infra", true},
		{"dev-usw2a1", false},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.expected, isProdContext(tc.context))
	}
}

func TestContextName(t *testing.T) {
	_, err := currentKubeConfigContext()
	assert.NoError(t, err)
}
