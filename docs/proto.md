# Protobuf Notes

Migrated to docker based buf

```shell
docker run --volume "$(pwd):/workspace" --workdir /workspace bufbuild/buf --version
0.53.0
```

Protoc lib version
```shell
docker run --volume "$(pwd):/workspace" --workdir /workspace bufbuild/buf protoc --version
3.13.0-buf
```

## Generating grpc boilerplate

```shell
make proto-gen # utilizing too a custom regen protoc generator https://github.com/regen-network/cosmos-proto/
Generating Protobuf files
W1213 14:18:24.688516     174 services.go:38] No HttpRule found for method: Msg.CreateIssuer
```