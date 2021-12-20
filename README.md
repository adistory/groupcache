GRPC pool for groupcache/v3
===========================

A replacement for [groupcachev2](https://github.com/mailgun/groupcache) `HTTPPool` that uses GRPC to communicate with peers.

Thanks to:
[charithe](https://github.com/charithe/gcgrpcpool)

Usage
-----

```go
import (
	...
    "github.com/xiaohuifirst/groupcache/v3"
)


server := grpc.NewServer()

p := NewGRPCPool("127.0.0.1:5000", server)
p.Set(peerAddrs...)

getter := groupcache.GetterFunc(func(ctx groupcache.Context, key string, dest groupcache.Sink) error {
	dest.SetString(...)
	return nil
})

groupcache.NewGroup("grpcPool", 1<<20, getter)
lis, err := net.Listen("tcp", "127.0.0.1:5000")
if err != nil {
	log.Fatalf("Failed to start server")
}

server.Serve(lis)
```

Use `GRPCPoolOptions` to set the GRPC client dial options such as using compression, authentication etc.

