# Internal communication protocol

This contains a grpc specification and some helper fuctions, meant for communicating internally between f2 components.

## Compiling

```
protoc --go_out=. --go_opt=paths=source_relative     --go-grpc_out=. --go-grpc_opt=paths=source_relative  forecaster.proto
```

