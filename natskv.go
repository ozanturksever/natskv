package natskv

import (
	"context"
	"errors"
	"github.com/cristalhq/base64"
	"github.com/kvtools/valkeyrie"
	"github.com/kvtools/valkeyrie/store"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"strings"
	"time"
)

const StoreName = "natskv"

func init() {
	valkeyrie.Register(StoreName, newStore)
}

type Config struct {
	Bucket string
	Nc     *nats.Conn
}

func newStore(ctx context.Context, endpoints []string, options valkeyrie.Config) (store.Store, error) {
	cfg, ok := options.(*Config)
	if !ok && options != nil {
		return nil, &store.InvalidConfigurationError{Store: StoreName, Config: options}
	}

	return New(ctx, endpoints, cfg)
}

type Store struct {
	nc *nats.Conn
	kv jetstream.KeyValue
}

func New(_ context.Context, endpoints []string, options *Config) (store.Store, error) {
	s := &Store{}

	bucket := "kvstore"
	if options != nil {
		if options.Bucket != "" {
			bucket = options.Bucket
		}
	}

	if options != nil {
		if options.Nc != nil {
			s.nc = options.Nc

		}
	}
	if s.nc == nil {
		nc, err := nats.Connect(strings.Join(endpoints, ","))
		if err != nil {
			return nil, err
		}
		s.nc = nc
	}

	js, err := jetstream.New(s.nc)
	if err != nil {
		return nil, err
	}

	kv, err := js.CreateKeyValue(context.Background(), jetstream.KeyValueConfig{
		Bucket: bucket,
	})
	if err != nil {
		return nil, err
	}
	s.kv = kv

	return s, nil
}

func (s *Store) Get(ctx context.Context, key string, _ *store.ReadOptions) (pair *store.KVPair, err error) {
	kve, err := s.kv.Get(ctx, normalizeKey(key))
	if errors.Is(err, jetstream.ErrKeyNotFound) {
		return nil, store.ErrKeyNotFound
	}
	if err != nil {
		return nil, err
	}

	pair = &store.KVPair{
		Key:       decodeKey(kve.Key()),
		Value:     kve.Value(),
		LastIndex: kve.Revision(),
	}

	return pair, nil
}

func (s *Store) Put(ctx context.Context, key string, value []byte, opts *store.WriteOptions) error {
	_, err := s.kv.Put(ctx, normalizeKey(key), value)

	return err
}

func (s *Store) Delete(ctx context.Context, key string) error {
	err := s.kv.Delete(ctx, normalizeKey(key))
	if errors.Is(err, jetstream.ErrKeyNotFound) {
		return store.ErrKeyNotFound
	}
	return err
}

func (s *Store) Exists(ctx context.Context, key string, _ *store.ReadOptions) (bool, error) {
	_, err := s.kv.Get(context.Background(), normalizeKey(key))
	if errors.Is(err, jetstream.ErrKeyNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *Store) Watch(ctx context.Context, key string, _ *store.ReadOptions) (<-chan *store.KVPair, error) {
	w, err := s.kv.Watch(ctx, normalizeKey(key))
	if err != nil {
		return nil, err
	}

	watchCh := make(chan *store.KVPair)
	go func() {
		defer close(watchCh)
		<-time.After(500 * time.Millisecond)

		for {
			select {
			case kve := <-w.Updates():
				if kve != nil {
					pair := &store.KVPair{
						Key:       decodeKey(kve.Key()),
						Value:     kve.Value(),
						LastIndex: kve.Revision(),
					}
					watchCh <- pair
				}
			case <-ctx.Done():
				w.Stop()
				return
			}
		}
	}()

	return watchCh, nil
}
func toWildcard(key string) string {
	key = normalizeKey(key)
	if key[len(key)-1] != '.' {
		key += "."
	}
	return key + "*"
}

func (s *Store) WatchTree(ctx context.Context, directory string, opts *store.ReadOptions) (<-chan []*store.KVPair, error) {
	innerWatchCh, err := s.Watch(ctx, toWildcard(directory), opts)
	if err != nil {
		return nil, err
	}

	watchCh := make(chan []*store.KVPair)

	go func() {
		defer close(watchCh)

		for {
			select {
			case kve := <-innerWatchCh:
				watchCh <- []*store.KVPair{kve}
			case <-ctx.Done():
				// There is no way to stop ChildrenW so just quit.
				return
			}
		}
	}()

	return watchCh, nil
}

func isInDirectory(directory, key string) bool {
	if directory[len(directory)-1] != '.' {
		directory += "."
	}
	return strings.HasPrefix(key, directory)
}

func normalizeKey(key string) string {
	k := strings.ReplaceAll(key, "/", ".")
	if k[len(k)-1] == '.' {
		return k[:len(k)-1]
	}
	return base64.StdEncoding.EncodeToString([]byte(k))
}

func decodeKey(k string) string {
	d, err := base64.StdEncoding.DecodeToString([]byte(k))
	if err != nil {
		return string(k)
	}
	return d
}

func (s *Store) List(ctx context.Context, directory string, opts *store.ReadOptions) ([]*store.KVPair, error) {
	kl, err := s.kv.ListKeys(ctx)
	if err != nil {
		return nil, err

	}
	var kvs []*store.KVPair
	exists := false
	for k := range kl.Keys() {
		if strings.HasPrefix(k, normalizeKey(directory)) {
			exists = true
		}
		if !isInDirectory(normalizeKey(directory), k) {
			continue
		}
		kve, err := s.kv.Get(ctx, k)
		if err != nil {
			return nil, err
		}
		kvs = append(kvs, &store.KVPair{
			Key:       decodeKey(kve.Key()),
			Value:     kve.Value(),
			LastIndex: kve.Revision(),
		})
	}

	if !exists {
		return nil, store.ErrKeyNotFound
	}
	return kvs, nil
}

func (s *Store) DeleteTree(ctx context.Context, directory string) error {
	kl, err := s.List(ctx, directory, nil)
	if err != nil {
		return err
	}
	for _, k := range kl {
		if err := s.Delete(ctx, k.Key); err != nil {
			return err
		}
	}
	return err
}

// AtomicPut puts a value at "key" if the key has not been modified in the meantime,
// throws an error if this is the case.
func (s *Store) AtomicPut(ctx context.Context, key string, value []byte, previous *store.KVPair, opts *store.WriteOptions) (bool, *store.KVPair, error) {
	key = normalizeKey(key)
	if previous != nil {
		rev, err := s.kv.Update(ctx, key, value, previous.LastIndex)
		if err != nil {
			// Compare Failed.
			//if errors.Is(err, jetstream.Err) {
			//	return false, nil, store.ErrKeyModified
			//}
			if strings.Contains(err.Error(), "wrong last sequence") {
				return false, nil, store.ErrKeyModified
			}
			return false, nil, err
		}

		pair := &store.KVPair{
			Key:       decodeKey(key),
			Value:     value,
			LastIndex: rev,
		}

		return true, pair, nil
	}

	rev, err := s.kv.Create(ctx, key, value)
	if errors.Is(err, jetstream.ErrKeyExists) {
		return false, nil, store.ErrKeyExists
	}
	if err != nil {
		return false, nil, err
	}
	return true, &store.KVPair{
		Key:       decodeKey(key),
		Value:     value,
		LastIndex: rev,
	}, nil
}

func (s *Store) AtomicDelete(ctx context.Context, key string, previous *store.KVPair) (bool, error) {
	key = normalizeKey(key)
	if previous == nil {
		return false, store.ErrPreviousNotSpecified
	}

	exists, err := s.Exists(ctx, key, nil)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, store.ErrKeyNotFound
	}
	err = s.kv.Delete(ctx, key, jetstream.LastRevision(previous.LastIndex))
	if err != nil {
		if strings.Contains(err.Error(), "wrong last sequence") {
			return false, store.ErrKeyModified
		}
		return false, err
	}
	return true, nil
}

func (s *Store) NewLock(_ context.Context, key string, opts *store.LockOptions) (lock store.Locker, err error) {
	return nil, store.ErrCallNotSupported
}

// Close closes the client connection.
func (s *Store) Close() error {
	s.nc.Close()
	return nil
}
