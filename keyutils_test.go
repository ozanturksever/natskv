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
	t.Run("dir should not match in key", func(t *testing.T) {
		u := keyUtils{options: &Config{EncodeKey: true}}
		isExist := u.isInDirectory("Dashboard", "DashboardCategory.co7663n3vlts3ko6o2pg")
		require.False(t, isExist)
	})
	t.Run("dir should match in key", func(t *testing.T) {
		u := keyUtils{options: &Config{EncodeKey: true}}
		isExist := u.isInDirectory("Dashboard", "Dashboard.co7663n3vlts3ko6o2pg")
		require.True(t, isExist)
	})
}
