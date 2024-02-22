package natskv

import (
	"context"
	"testing"
	"time"

	"github.com/kvtools/valkeyrie"
	"github.com/kvtools/valkeyrie/store"
	"github.com/kvtools/valkeyrie/testsuite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testTimeout = 60 * time.Second

const enpoint = "nats://localhost:4222"

func makeClient(t *testing.T) store.Store {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	config := &Config{}

	kv, err := New(ctx, []string{enpoint}, config)
	require.NoErrorf(t, err, "cannot create store")

	return kv
}

func TestRegister(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	kv, err := valkeyrie.NewStore(ctx, StoreName, []string{enpoint}, nil)
	require.NoError(t, err)
	assert.NotNil(t, kv)

	if _, ok := kv.(*Store); !ok {
		t.Fatal("Error registering and initializing zookeeper")
	}
}

func TestNatsKvStore(t *testing.T) {
	kv := makeClient(t)

	testsuite.RunTestCommon(t, kv)
	testsuite.RunTestAtomic(t, kv)
	testsuite.RunTestWatch(t, kv)
	testsuite.RunCleanup(t, kv)
}
