package valkeystore

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/valkey-io/valkey-go"
)

func TestFind(t *testing.T) {
	opt, err := valkey.ParseURL(os.Getenv("SCS_VALKEY_TEST_DSN"))
	if err != nil {
		t.Fatal(err)
	}

	client, err := valkey.NewClient(opt)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	ctx := context.Background()
	v := New(client)

	err = client.Do(ctx, client.B().Flushdb().Build()).Error()
	if err != nil {
		t.Fatal(err)
	}

	cmd := client.B().Set().Key(v.prefix + "session_token").Value("encoded_data").Build()
	err = client.Do(ctx, cmd).Error()
	if err != nil {
		t.Fatal(err)
	}

	b, found, err := v.FindCtx(ctx, "session_token")
	if err != nil {
		t.Fatal(err)
	}

	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}
	if bytes.Equal(b, []byte("encoded_data")) == false {
		t.Fatalf("got %v: expected %v", b, []byte("encoded_data"))
	}
}
