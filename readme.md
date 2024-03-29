# Valkeyrie NatsKV

[![GoDoc](https://godoc.org/github.com/ozanturksever/natskv?status.png)](https://godoc.org/github.com/ozanturksever/natskv)
[![Build Status](https://github.com/ozanturksever/natskv/actions/workflows/build.yml/badge.svg)](https://github.com/ozanturksever/natskv/actions/workflows/build.yml)

[`valkeyrie`](https://github.com/kvtools/valkeyrie) provides a Go native library to store metadata using Distributed Key/Value stores (or common databases).

## Compatibility

A **storage backend** in `valkeyrie` implements (fully or partially) the [Store](https://github.com/kvtools/valkeyrie/blob/master/store/store.go#L69) interface.

| Calls                 | NatsKV |
|-----------------------|:-----:|
| Put                   |  🟢️  |
| Get                   |  🟢️  |
| Delete                |  🟢️  |
| Exists                |  🟢️  |
| Watch                 |  🟢️  |
| WatchTree             |  🟢️  |
| NewLock (Lock/Unlock) |  ️    |
| List                  |  🟢️  |
| DeleteTree            |  🟢️  |
| AtomicPut             |  🟢️  |
| AtomicDelete          |  🟢️  |

## Supported Versions

nats-server versions >= `2.10`.

## Examples

```go
package main

import (
	"context"
	"log"

	"github.com/kvtools/valkeyrie"
	"github.com/ozanturksever/natskv"
)

func main() {
	ctx := context.Background()

	config := &natskv.Config{
        Bucket: "example",
		EncodeKey: false,
	}

	kv, err := valkeyrie.NewStore(ctx, natskv.StoreName, []string{ "nats://localhost:4222"}, config)
	if err != nil {
		log.Fatal("Cannot create store")
	}

	key := "foo"

	err = kv.Put(ctx, key, []byte("bar"), nil)
	if err != nil {
		log.Fatalf("Error trying to put value at key: %v", key)
	}

	pair, err := kv.Get(ctx, key, nil)
	if err != nil {
		log.Fatalf("Error trying accessing value at key: %v", key)
	}

	log.Printf("value: %s", string(pair.Value))

	err = kv.Delete(ctx, key)
	if err != nil {
		log.Fatalf("Error trying to delete key %v", key)
	}
}
```
