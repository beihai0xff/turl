package turl

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/beihai0xff/turl/internal/tests"
)

func TestNewServer(t *testing.T) {
	got, err := NewServer(tests.GlobalConfig)
	require.NoError(t, err)
	require.NotNil(t, got)
}
