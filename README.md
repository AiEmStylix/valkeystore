# valkeystore

A [valkey-go](https://github.com/valkey-io/valkey-go) bases session store for [scs](https://github.com/alexedwards/scs)

## Install 

```go
go get github.com/AiEmStylix/valkeystore
```

## Setup

You should follow the instructions to [set a client](https://github.com/valkey-io/valkey-go#getting-started),
and pass the client to `valkeystore.New()` or `valkeystore.NewWithPrefix()`
to establish the session store.

## Keys

Default key is `scs:session:`, you can change it via

```go
sessionManager.Store = valkeystore.NewWithPrefix(client, "scs:session:1:")
```
