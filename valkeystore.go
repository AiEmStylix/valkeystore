package valkeystore

import (
	"context"
	"time"

	"github.com/valkey-io/valkey-go"
)

type ValkeyStore struct {
	client valkey.Client
	prefix string
}

// New returns a new ValkeyStore instance. The client parameter should be a pointer
// to a go-redis connection.

// valkey.Client is already an interface, so we don't use a pointer for param
func New(client valkey.Client) *ValkeyStore {
	return NewWithPrefix(client, "scs:session:")
}

func NewWithPrefix(client valkey.Client, prefix string) *ValkeyStore {
	return &ValkeyStore{
		client: client,
		prefix: prefix,
	}
}

// FindCtx returns the data for a given session token from the ValkeyStore instance.
// If the session token is not found or is expired, the returned exists flag
// will be set to false.

func (v *ValkeyStore) FindCtx(ctx context.Context, token string) (b []byte, exists bool, err error) {
	cmd := v.client.B().Get().Key(v.prefix + token).Build()

	resp, err := v.client.Do(ctx, cmd).AsBytes()
	if err != nil {
		if valkey.IsValkeyNil(err) {
			return nil, false, nil
		}
	}

	return resp, true, nil
}

func (v *ValkeyStore) CommitCtx(ctx context.Context, token string, b []byte, expiry time.Time) error {
	duration := time.Until(expiry)

	cmd := v.client.B().Set().Key(v.prefix + token).Value(valkey.BinaryString(b)).Px(duration).Build()
	return v.client.Do(ctx, cmd).Error()
}

func (v *ValkeyStore) DeleteCtx(ctx context.Context, token string) error {
	cmd := v.client.B().Del().Key(v.prefix + token).Build()
	return v.client.Do(ctx, cmd).Error()
}

func (v *ValkeyStore) AllCtx(ctx context.Context) (map[string][]byte, error) {
	var cursor uint64
	sessions := make(map[string][]byte)

	for {
		cmd := v.client.B().Scan().Cursor(cursor).Match(v.prefix + "*").Build()

		entry, err := v.client.Do(ctx, cmd).AsScanEntry()
		if err != nil {
			if valkey.IsValkeyNil(err) {
				return nil, nil
			}
			return nil, err
		}

		for _, key := range entry.Elements {
			token := key[len(v.prefix):]
			data, exists, err := v.FindCtx(ctx, token)
			if err != nil {
				return nil, err
			}
			if exists {
				sessions[token] = data
			}
		}

		if entry.Cursor == 0 {
			break
		}
		cursor = entry.Cursor
	}

	return sessions, nil
}

//
// We have to add the plain Store methods here to be recognized a Store
// by the go compiler. Not using a seperate type makes any errors caught
// only at runtime instead of compile time. Oh well.

func (v *ValkeyStore) Find(token string) ([]byte, bool, error) {
	panic("missing context arg")
}

func (v *ValkeyStore) Commit(token string, b []byte, expiry time.Time) error {
	panic("missing context arg")
}

func (v *ValkeyStore) Delete(token string) error {
	panic("missing context arg")
}
