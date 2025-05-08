# gRPC Example

This example demonstrates how to use GoVisual with a gRPC server.

The example actually runs two servers:
- **gRPC**: This would be the gRPC server you want to scan requests for. It is setup with interceptors that collect the request and response data to display in the dashboard.
- **HTTP**: This server simply hosts the dashboard UI since the gRPC server cant.

Both servers share the same store which allows the dashboard (on the HTTP server) to know about the requests made to the gRPC server.

## Prerequisites:
- This example has a [Taskfile](https://taskfile.dev/) which is used for some of the commands. If you would rather not install the taskfile CLI, you can instead go to [taskfile.yaml](./taskfile.yaml) and find the commands you're looking for.
- This example makes use of [go tool management](https://www.alexedwards.net/blog/how-to-manage-tool-dependencies-in-go-1.24-plus). You can simply run `go get -modfile=go.tool.mod tool` to download the go packages that this project uses.

## Running the Example

Start the HTTP server on port 8080 and the gRPC server on port 9090:

```bash
go run .
```

An initial gRPC call is made so you can easily see what it looks like in the dashboard at [http://localhost:8080/\_\_viz](http://localhost:8080/__viz).

In a new terminal session, start the [grpcui](https://github.com/fullstorydev/grpcui) docs server:

```bash
task docs
```

If you would like to use a different tool such as `grpcurl` then you can reference the proto file located at [proto/greeter/v1/greeter.proto](./proto/greeter/v1/greeter.proto).

## Features Demonstrated

- gRPC interceptor
- Dual server
- Shared stores
