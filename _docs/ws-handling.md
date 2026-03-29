# WebSocket Infrastructure and Patterns

## Overview

This document specifies the common infrastructure and patterns for WebSocket-based data streaming across all providers in `bits`. The goal is to provide a robust, stateful, and idiomatic interface that handles the lifecycle of message loops, authentication, and command passing.

## Core Infrastructure: `ws.Manager`

The `ws.Manager` is a stateful component that wraps the connection logic and provides a high-level API for commands and data.

### Stateful Lifecycle

Each provider instance maintains a `ws.Manager`. The manager is responsible for:
- Establishing and maintaining the connection.
- Handling the authentication handshake (if required).
- Implementing the ping/pong keep-alive logic.
- Processing incoming messages and routing them to the appropriate channels.
- Managing subscriptions via an input command channel.

### Specifications and Interfaces

#### Input Command Channel

The manager exposes an input channel for commands:

```go
type CommandKind string

const (
    CommandStop        CommandKind = "stop"
    CommandSubscribe   CommandKind = "subscribe"
    CommandUnsubscribe CommandKind = "unsubscribe"
)

type Command struct {
    Kind    CommandKind
    Params  any         // e.g., []string for symbols
    Context context.Context
}
```

#### Output Data Channel

Data is streamed through a channel of `model.Response[T]`, ensuring consistency with the project's response policy.

```go
type StreamResponse[T any] struct {
    Response model.Response[T]
    Error    error
}
```

#### Interface

```go
type StreamProvider interface {
    Stream(ctx context.Context, cmdChan <-chan Command) (<-chan StreamResponse[any], error)
}
```

## Patterns and Tools

### Authentication

Providers that require authentication should implement an `Authorizer` interface or provide a callback to the manager.

### Keep-Alive (Ping/Pong)

The manager should automatically send ping messages based on a configurable interval and handle pong responses to detect stale connections.

### Message Passing

The manager uses a unified message loop that:
1. Listens for commands from the `cmdChan`.
2. Reads messages from the WebSocket.
3. Decodes messages into provider-specific structs.
4. Maps provider structs to `model.Response`.
5. Pushes the response to the output channel.

## Implementation Details

### Reconnection Logic

Uses the existing `ws.BaseClient` backoff strategy. The manager should ensure that subscriptions are restored after a successful reconnection.

### Error Handling

All errors (connection, parsing, etc.) are wrapped in `StreamResponse` and sent to the output channel.

## Usage Example

```go
mgr := ws.NewManager(url, handler)
dataChan, err := mgr.Start(ctx)

// Subscribe to new symbols
mgr.Commands() <- ws.Command{
    Kind: ws.CommandSubscribe,
    Params: []string{"BTC_USDT"},
}
```
