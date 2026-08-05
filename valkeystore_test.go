package valkeystore

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

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

func TestSaveNew(t *testing.T) {
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
	r := New(client)

	err = client.Do(ctx, client.B().Flushdb().Build()).Error()
	if err != nil {
		t.Fatal(err)
	}

	err = r.CommitCtx(ctx, "session_token", []byte("encoded_data"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	cmd := client.B().Get().Key(r.prefix + "session_token").Build()
	data, err := client.Do(ctx, cmd).AsBytes()
	if err != nil {
		t.Fatal(err)
	}

	if reflect.DeepEqual(data, []byte("encoded_data")) == false {
		t.Fatalf("got %v: expected %v", data, []byte("encoded_data"))
	}
}

func TestFindMissing(t *testing.T) {
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
	r := New(client)

	err = client.Do(ctx, client.B().Flushdb().Build()).Error()
	if err != nil {
		t.Fatal(err)
	}

	_, found, err := r.FindCtx(ctx, "missing_session_token")
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestSaveUpdated(t *testing.T) {
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
	r := New(client)

	err = client.Do(ctx, client.B().Flushdb().Build()).Error()
	if err != nil {
		t.Fatal(err)
	}

	cmdSet := client.B().Set().Key(r.prefix + "session_token").Value("encoded_data").Build()
	err = client.Do(ctx, cmdSet).Error()
	if err != nil {
		t.Fatal(err)
	}

	err = r.CommitCtx(ctx, "session_token", []byte("new_encoded_data"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	cmdGet := client.B().Get().Key(r.prefix + "session_token").Build()
	data, err := client.Do(ctx, cmdGet).AsBytes()
	if err != nil {
		t.Fatal(err)
	}

	if reflect.DeepEqual(data, []byte("new_encoded_data")) == false {
		t.Fatalf("got %v: expected %v", data, []byte("new_encoded_data"))
	}
}

func TestExpiry(t *testing.T) {
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
	r := New(client)

	err = client.Do(ctx, client.B().Flushdb().Build()).Error()
	if err != nil {
		t.Fatal(err)
	}

	err = r.CommitCtx(ctx, "session_token", []byte("encoded_data"), time.Now().Add(100*time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}

	_, found, _ := r.FindCtx(ctx, "session_token")
	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}

	time.Sleep(200 * time.Millisecond)
	_, found, _ = r.FindCtx(ctx, "session_token")
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestDelete(t *testing.T) {
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
	r := New(client)

	err = client.Do(ctx, client.B().Flushdb().Build()).Error()
	if err != nil {
		t.Fatal(err)
	}

	cmdSet := client.B().Set().Key(r.prefix + "session_token").Value("encoded_data").Build()
	err = client.Do(ctx, cmdSet).Error()
	if err != nil {
		t.Fatal(err)
	}

	err = r.DeleteCtx(ctx, "session_token")
	if err != nil {
		t.Fatal(err)
	}

	cmdGet := client.B().Get().Key(r.prefix + "session_token").Build()
	data, err := client.Do(ctx, cmdGet).AsBytes()

	if !valkey.IsValkeyNil(err) {
		t.Fatal(err)
	}
	if data != nil {
		t.Fatalf("got %v: expected %v", data, nil)
	}
}

func TestAll(t *testing.T) {
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
	r := New(client)

	err = client.Do(ctx, client.B().Flushdb().Build()).Error()
	if err != nil {
		t.Fatal(err)
	}

	sessions := make(map[string][]byte)
	for i := range 4 {
		key := fmt.Sprintf("token_%v", i)
		val := []byte(key)
		cmd := client.B().Set().Key(r.prefix + key).Value(key).Build()
		err = client.Do(ctx, cmd).Error()
		if err != nil {
			t.Fatal(err)
		}
		sessions[key] = val
	}

	gotSessions, err := r.AllCtx(ctx)
	if err != nil {
		t.Fatal(err)
	}

	for k := range sessions {
		err = r.DeleteCtx(ctx, k)
		if err != nil {
			t.Fatal(err)
		}
	}
	if reflect.DeepEqual(sessions, gotSessions) == false {
		t.Fatalf("got %v: expected %v", gotSessions, sessions)
	}
}
