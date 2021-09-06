# fluentdlib

A library for Fluentd's [forward](https://github.com/fluent/fluentd/wiki/Forward-Protocol-Specification-v1) protocol and tools for Fluentd / Fluent Bit

## Tools

Dump contents of Fluentd Forward messages (Forward, PackedForward, CompressedPackedForward) and Fluent Bit chunk files:

```bash
fluentlibtool dump [filepath]...
```

Run a fake Fluentd server to print all logs in JSON to file or stdout (pass "-" as filename)

```bash
fluentlibtool server -f 0 -x 0 -n 0 output.json
```

(`-f`, `-x`, and `-n` are to simulate network errors etc, use `fluentlibtool help server` to get help)

## Library

- `protocol/fluentbitchunk` can decode Fluent Bit's [internal chunk (buffer) files](https://docs.fluentbit.io/manual/administration/buffering-and-storage).
- `protocol/forwardprotocol` provides definitions of [Fluentd Forward Protocol v1](https://github.com/fluent/fluentd/wiki/Forward-Protocol-Specification-v1) in Go, as well as utility functions for handshaking and decoding.
- `server` provides a fake Fluentd server that can be used for testing

The library part is intended for verification and functions here are NOT optimized for performance.

See `server/server_test.go:TestServerBasic` for basic examples of a client

## Build

1. Install go
2. Install https://github.com/relex/gotils

```bash
make
make test
```
