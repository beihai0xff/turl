package configs

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestReadFile(t *testing.T) {
	c, err := ReadFile("../internal/example/config.yaml")
	require.NoError(t, err)
	require.NotNil(t, c)

	fmt.Printf("%+v\n", c)
	require.Equal(t, "0.0.0.0", c.Listen)
	require.Equal(t, 8080, c.Port)
	require.Equal(t, 5*time.Second, c.RequestTimeout)
	require.Equal(t, "turl_rate_limit", c.GlobalRateLimitKey)
	require.Equal(t, 100, c.GlobalWriteRate)
	require.Equal(t, 200, c.GlobalWriteBurst)
	require.Equal(t, 10000, c.StandAloneReadRate)
	require.Equal(t, 500, c.StandAloneReadBurst)
}
