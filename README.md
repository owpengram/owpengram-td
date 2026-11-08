# owpengram-td

The MTProto/TL protocol library behind the [OwpenGram](https://github.com/owpengram) server,
published as `github.com/owpengram/owpengram-td`.

It is [gotd/td](https://github.com/gotd/td) (a pure-Go MTProto 2.0 implementation) plus a
**multi-layer server runtime**: one server binary answers clients on any supported Telegram API
Layer at once. Requests are decoded and every response, update and difference is encoded for the
exact Layer the client negotiated, instead of only the newest one.

- `tg` is the canonical Go model of the newest Layer (currently **233**).
- `tlprofile` holds the exact historical profiles (**225–233**) and their generated codecs,
  admission rules and dispatcher.
- `gen` / `cmd/gotdgen` build both from the versioned schemas in
  [`_schema/layers`](_schema/layers/README.md).

## Schemas come from Telegram Desktop

Every Layer is the verbatim `api.tl` of an official Telegram Desktop revision. The manifest records
the tdesktop commit, git blob id and SHA-256 of each file, and the generator refuses to run if a
file does not match.

Adding a new Layer:

```console
go run ./cmd/layerimport -ref <tdesktop tag or commit> -canonical
# review _schema/layers/policy.json for the new Layer (dropped fields, rejected constructors)
go generate .
go test ./...
```

`layerimport` downloads the schema, checks its `// LAYER N` line, writes `layer-N.tl` and updates
`manifest.json`. The policy review stays manual on purpose: it decides what older clients may
receive, and generation fails closed until it is done.

## Origin and credits

- [gotd/td](https://github.com/gotd/td) by Aleksandr Razumov and contributors -- the protocol
  library, client and code generator this repository is forked from.
- The multi-layer runtime (`tlprofile`, the layer-aware parts of `gen`, `_schema/layers`) was
  developed in [iamxvbaba/td](https://github.com/iamxvbaba/td) and imported here at its `v1.7.0`
  (Layer 232). The generated output was verified to be identical to that release apart from the
  module path.

Both are MIT-licensed; see [LICENSE](LICENSE). From here on this repository is maintained
independently: changes from gotd/td are reviewed and replayed selectively
([`UPSTREAM_SYNC.md`](UPSTREAM_SYNC.md) keeps the historical audit log), and new Layers are
imported from Telegram Desktop directly.

## gotd/td

The rest of this README is gotd's own and describes the client side of the library, which is kept
working but is not what OwpenGram uses it for.

Telegram MTProto API client in Go for users and bots.

[![Telegram: English chat](https://api.go-faster.org/badge/telegram/gotd_en?title=EN&v=1)](https://t.me/gotd_en) [![Telegram: Russian chat](https://api.go-faster.org/badge/telegram/gotd_ru?title=RU&v=1)](https://t.me/gotd_ru) [![Telegram: Chinese chat](https://api.go-faster.org/badge/telegram/gotd_zhcn?title=CN&v=1)](https://t.me/gotd_zhcn) [![Telegram: Online count](https://api.go-faster.org/badge/telegram/online?groups=gotd_ru&groups=gotd_en&groups=gotd_zhcn)](https://t.me/gotd_en)

* [Examples](examples)
* [Security policy](.github/SECURITY.md)
* [User support and dev chat](.github/SUPPORT.md)
* [Roadmap](ROADMAP.md)
* [Contributing](CONTRIBUTING.md)
* [Architecture](ARCHITECTURE.md)
* [Multi-layer TL generation](_schema/layers/README.md)
* [Generated Go Documentation](https://ref.gotd.dev/pkg/github.com/gotd/td/tg.html)

Before using this library, read [How To Not Get Banned](.github/SUPPORT.md#how-to-not-get-banned) guide.

Due to limitations of `pkg.go.dev`, documentation for `tg` package is not shown, but there is [hosted version](https://ref.gotd.dev/pkg/github.com/gotd/td/tg.html).

## Higher level libraries

The `gotd` itself is a pretty low-level library, and you may want to use higher level libraries instead:
- [GoTGProto][gotgproto] is a helper package for gotd library, It aims to make td's raw functions easy-to-use with the help of features like using session strings, custom helper functions, storing peers and extracting chat or user ids through it etc.

[gotgproto]: https://github.com/celestix/gotgproto

## Usage

```console
go get github.com/owpengram/owpengram-td
```

```go
package main

import (
	"context"

	"github.com/owpengram/owpengram-td/telegram"
)

func main() {
	// https://core.telegram.org/api/obtaining_api_id
	client := telegram.NewClient(appID, appHash, telegram.Options{})
	if err := client.Run(context.Background(), func(ctx context.Context) error {
		// It is only valid to use client while this function is not returned
		// and ctx is not cancelled.
		api := client.API()

		// Now you can invoke MTProto RPC requests by calling the API.
		// ...

		// Return to close client connection and free up resources.
		return nil
	}); err != nil {
		panic(err)
	}
	// Client is closed.
}
```

See [examples](examples) for more info.

## Features

* Full MTProto 2.0 implementation in Golang, directly access any MTProto method with [telegram.Client.API()](https://pkg.go.dev/github.com/owpengram/owpengram-td/telegram#Client.API)
* Voice and video [calls](https://pkg.go.dev/github.com/owpengram/owpengram-td/telegram/calls) (tgcalls), both 1:1 and group calls — see the [call](examples/call) and [groupcall](examples/groupcall) examples
* Highly optimized, low memory (150kb per idle client) and CPU overhead, can handle thousands concurrent clients
* Code for Telegram types generated by `./cmd/gotdgen` (based on [gotd/tl](https://github.com/gotd/tl) parser) with embedded [official documentation](https://core.telegram.org/schema)
* Pluggable session storage
* Automatic re-connects with keepalive
* Vendored Telegram public keys that are kept up-to-date
* Rigorously tested
  * End-to-end with real Telegram server in CI
  * End-to-end with gotd Telegram server (in pure Go)
  * Lots of unit testing
  * Fuzzing
  * 24/7 canary bot in production that tests reconnects, update handling, memory leaks and performance
* No runtime reflection overhead
* [Conforms](https://github.com/gotd/td/issues/155) to [Security guidelines](https://core.telegram.org/mtproto/security_guidelines) for Telegram client software developers
  * Secure PRNG used for crypto
  * Replay attack protection
* 2FA support
* MTProxy support
* Various helpers that lighten the complexity of the Telegram API
  * [uploads](https://pkg.go.dev/github.com/owpengram/owpengram-td/telegram/uploader) for big and small files with multiple streams for single file and progress reporting
  * [downloads](https://pkg.go.dev/github.com/owpengram/owpengram-td/telegram/downloader) with CDN support, also multiple streams
  * [messages](https://pkg.go.dev/github.com/owpengram/owpengram-td/telegram/message) with various convenience builders and text styling support
  * [query](https://pkg.go.dev/github.com/owpengram/owpengram-td/telegram/query) with pagination helpers
  * [middleware](https://pkg.go.dev/github.com/owpengram/owpengram-td/telegram#Middleware) for [rate limiting](https://pkg.go.dev/github.com/gotd/contrib/middleware/ratelimit) and [FLOOD_WAIT handling](https://pkg.go.dev/github.com/gotd/contrib/middleware/floodwait)
* Connection pooling
* Automatic datacenter migration and redirects handling
* Graceful [request cancellation](https://core.telegram.org/mtproto/service_messages#cancellation-of-an-rpc-query) via context
* WebSocket transport support (works in WASM)

## Status

The goal of this project is to implement a stable, performant and safe client for Telegram in pure Go while
having a simple and convenient API and a feature parity with TDLib.

This project is fully non-commercial and not affiliated with any commercial organization
(including Telegram LLC).

## Examples

See [examples](examples) directory.

Also take a look at

* [go-faster/bot](https://github.com/go-faster/bot) as example of sophisticated telegram bot integration with GitHub
* [gotd/cli](https://github.com/gotd/cli), command line interface for subset of telegram methods.

### Auth

#### User

You can use `td/telegram/auth.Flow` to simplify user authentications.

```go
codePrompt := func(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
    // NB: Use "golang.org/x/crypto/ssh/terminal" to prompt password.
    fmt.Print("Enter code: ")
    code, err := bufio.NewReader(os.Stdin).ReadString('\n')
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(code), nil
}
// This will setup and perform authentication flow.
// If account does not require 2FA password, use telegram.CodeOnlyAuth
// instead of telegram.ConstantAuth.
if err := auth.NewFlow(
    auth.Constant(phone, password, auth.CodeAuthenticatorFunc(codePrompt)),
    auth.SendCodeOptions{},
).Run(ctx, client.Auth()); err != nil {
    panic(err)
}
```

#### Bot

Use bot token from [@BotFather](https://telegram.me/BotFather).

```go
if err := client.Auth().Bot(ctx, "token:12345"); err != nil {
    panic(err)
}
```

### Calling MTProto directly

You can use the generated `tg.Client` that allows calling any MTProto method
directly.

```go
// Grab these from https://my.telegram.org/apps.
// Never share it or hardcode!
client := telegram.NewClient(appID, appHash, telegram.Options{})
client.Run(ctx, func(ctx context.Context) error {
  // Grab token from @BotFather.
  if _, err := client.Auth().Bot(ctx, "token:12345"); err != nil {
    return err
  }
  state, err := client.API().UpdatesGetState(ctx)
  if err != nil {
    return err
  }
  // Got state: &{Pts:197 Qts:0 Date:1606855030 Seq:1 UnreadCount:106}
  // This will close client and cleanup resources.
  return nil
})
```

### Using MTProto JSON format

You can invoke methods using MTProto JSON format with `InvokeJSON`.
This accepts JSON with `@type` field in MTProto format (e.g., `"messages.sendMessage"`)
and returns JSON response in MTProto format.

```go
client := telegram.NewClient(appID, appHash, telegram.Options{})
client.Run(ctx, func(ctx context.Context) error {
  if _, err := client.Auth().Bot(ctx, "token:12345"); err != nil {
    return err
  }

  // Send message using MTProto JSON format
  jsonReq := `{
    "@type": "messages.sendMessage",
    "peer": {"@type": "inputPeerUser", "user_id": 123456789, "access_hash": 0},
    "message": "Hello from MTProto JSON!",
    "random_id": 0
  }`

  jsonResp, err := client.API().InvokeJSON(ctx, jsonReq, true)
  if err != nil {
    return err
  }
  // jsonResp contains MTProto JSON response with snake_case field names
  // {"@type":"messages.sentMessage",...} or error response
  return nil
})
```

### Generated code

The code output of `gotdgen` contains references to TL types, examples, URL to
official documentation and [extracted](https://github.com/gotd/getdoc) comments from it.

For example, the [auth.Authorization](https://core.telegram.org/type/auth.Authorization) type in `tg/tl_auth_authorization_gen.go`:

```go
// AuthAuthorizationClass represents auth.Authorization generic type.
//
// See https://core.telegram.org/type/auth.Authorization for reference.
//
// Example:
//  g, err := DecodeAuthAuthorization(buf)
//  if err != nil {
//      panic(err)
//  }
//  switch v := g.(type) {
//  case *AuthAuthorization: // auth.authorization#cd050916
//  case *AuthAuthorizationSignUpRequired: // auth.authorizationSignUpRequired#44747e9a
//  default: panic(v)
//  }
type AuthAuthorizationClass interface {
	bin.Encoder
	bin.Decoder
	construct() AuthAuthorizationClass
}
```

Also, the corresponding [auth.signIn](https://core.telegram.org/method/auth.signIn) method:

```go
// AuthSignIn invokes method auth.signIn#bcd51581 returning error if any.
// Signs in a user with a validated phone number.
//
// See https://core.telegram.org/method/auth.signIn for reference.
func (c *Client) AuthSignIn(ctx context.Context, request *AuthSignInRequest) (AuthAuthorizationClass, error) {}
```

The generated constructors contain detailed official documentation, including links:

```go
// FoldersDeleteFolderRequest represents TL type `folders.deleteFolder#1c295881`.
// Delete a peer folder¹
//
// Links:
//  1) https://core.telegram.org/api/folders#peer-folders
//
// See https://core.telegram.org/method/folders.deleteFolder for reference.
type FoldersDeleteFolderRequest struct {
    // Peer folder ID, for more info click here¹
    //
    // Links:
    //  1) https://core.telegram.org/api/folders#peer-folders
    FolderID int
}
```

## Contributions

Huge thanks to all contributors. Dealing with a project of this scale alone is impossible.

Special thanks:

* [tdakkota](https://github.com/tdakkota)
  * Two-factor authentication (SRP)
  * Proxy support
  * Update dispatcher
  * Complete transport support (abridged, padded intermediate and full)
  * Telegram server for end-to-end testing
  * Multiple major refactorings, including critical cryptographical scope reduction
  * Code generation improvements (vector support, multiple modes for pretty-print)
  * And many other cool things and performance improvements
* [shadowspore](https://github.com/shadowspore)
  * Background pings
  * Links in generated documentation
  * Message acknowledgements
  * Retries
  * RPC Engine
  * Gap (Updates) engine

## Reference

The MTProto protocol description is [hosted](https://core.telegram.org/mtproto#general-description) by Telegram.

Most important parts for client implementations:

* [Security guidelines](https://core.telegram.org/mtproto/security_guidelines) for client software developers

Current implementation [mostly conforms](https://github.com/gotd/td/issues/155) to security guidelines, but no
formal security audit were performed.

## Prior art

* [Lonami/grammers](https://github.com/Lonami/grammers) (Great Telegram client in Rust, many test vectors were used as reference)
* [sdidyk/mtproto](https://github.com/sdidyk/mtproto), [cjongseok/mtproto](https://github.com/cjongseok/mtproto), [xelaj/mtproto](https://github.com/xelaj/mtproto) (MTProto 1.0 in go)

## Who is using gotd?

- The [iyear/tdl](https://github.com/iyear/tdl), 📥 Telegram Downloader, but more than a downloader

[<img src="https://user-images.githubusercontent.com/43930873/142855897-7091ced0-4fe8-4f8d-ad43-e9db2723bacc.png" width="150">](https://telq.org)

Drop a comment [here](https://github.com/gotd/td/issues/568) to add your project.

## License

MIT License

Created by Aleksandr (ernado) Razumov

2020

## Links

- [Examples](https://github.com/gotd/td/blob/-/examples/README.md)
- [Security policy](https://github.com/gotd/td/blob/-/.github/SECURITY.md)
- [User support and dev chat](https://github.com/gotd/td/blob/-/.github/SUPPORT.md)
- [Roadmap](https://github.com/gotd/td/blob/-/ROADMAP.md)
- [Contributing](https://github.com/gotd/td/blob/-/CONTRIBUTING.md)
- [Architecture](https://github.com/gotd/td/blob/-/ARCHITECTURE.md)
- [Generated Go Documentation](https://ref.gotd.dev/pkg/github.com/gotd/td/tg.html)
