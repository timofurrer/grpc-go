# Connection Age

This example demonstrates how to use the connection age functionality in gRPC-Go to track connection duration and implement graceful connection termination before reaching maximum connection age.

## Background

Load balancers in front of gRPC servers typically have timeout parameters that limit the maximum connection age. After this timeout elapses, the load balancer aborts the connection and clients see a connection reset, which is not ideal.

To mitigate this problem, long-running server-side operations (like poll loops) should stop before reaching the maximum connection age. gRPC sends a HTTP/2 GOAWAY frame to the client when the max connection age time elapses, allowing in-flight requests to finish cleanly during the grace period.

## New gRPC-Go API

The following experimental API has been added to help with connection age management:

### grpc.ConnectionAgeContext(ctx)

Returns a context that will be cancelled when the connection is considered "old" and should no longer be used for long-running operations:

```go
ageCtx, ok := grpc.ConnectionAgeContext(ctx)
if !ok {
    // Connection age context not available, use request context
    // ageCtx will just be ctx.
    // log something ... maybe?
}

select {
case <-ageCtx.Done():
    // Connection is getting old, terminate operation
    return status.Error(codes.DeadlineExceeded, "operation stopped due to connection age")
case result := <-longOperation():
return result, nil
}
```

The timeout is automatically calculated based on the server's keepalive MaxConnectionAge parameter, taking into account grace periods and jitter to ensure operations complete before the connection is terminated by gRPC.

## Usage

This approach eliminates the need for custom stats handlers to track connection age. Server handlers can directly check the connection age from the context and make decisions about whether to continue long-running operations.

## Example

See the `server/main.go` file for a complete example of how to use connection age in a long-running RPC handler.
