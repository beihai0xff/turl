package turl

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/beihai0xff/turl/internal/tests"
)

func TestNewServer(t *testing.T) {
	h, err := NewHandler(tests.GlobalConfig)
	require.NoError(t, err)
	got, err := NewServer(h, tests.GlobalConfig)
	require.NoError(t, err)
	require.NotNil(t, got)
}
