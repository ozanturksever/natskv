package natskv

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestKey(t *testing.T) {
	u := keyUtils{}
	k := u.normalizeKey("a")
	require.Equal(t, "a", k)
	k = u.normalizeKey("a/b/c")
	require.Equal(t, "a.b.c", k)
	k = u.normalizeKey("a/b/1.1.1.1")
	require.Equal(t, "a.b.1.1.1.1", k)
	d := u.decodeKey("a.b.c")
	require.Equal(t, "a/b/c", d)
	d = u.decodeKey("a.b.1.1.1.1")
	require.Equal(t, "a/b/1/1/1/1", d)

	t.Run("Encode Keys", func(t *testing.T) {
		u := keyUtils{options: &Config{EncodeKey: true}}
		k := u.normalizeKey("a")
		require.Equal(t, "YQ==", k)
		k = u.normalizeKey("a/b/c")
		require.Equal(t, "YQ==.Yg==.Yw==", k)
		k = u.normalizeKey("a/b/1.1.1.1")
		require.Equal(t, "YQ==.Yg==.MS4xLjEuMQ==", k)
		d := u.decodeKey("YQ==.Yg==.Yw==")
		require.Equal(t, "a/b/c", d)
		d = u.decodeKey("YQ==.Yg==.MS4xLjEuMQ==")
		require.Equal(t, "a/b/1.1.1.1", d)
	})
}
