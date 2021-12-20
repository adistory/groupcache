/*
Adapted from https://github.com/golang/groupcache/blob/master/http_test.go
*/

package groupcache

import (
	"context"
	"errors"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc"
)

func TestGRPCPool(t *testing.T) {
	if *peerChild {
		beChildForTestGRPCPool()
		os.Exit(0)
	}

	const (
		nChild = 5
		nGets  = 5000
	)

	var childAddr []string
	for i := 0; i < nChild; i++ {
		childAddr = append(childAddr, pickFreeAddr(t))
	}

	var cmds []*exec.Cmd
	var wg sync.WaitGroup
	for i := 0; i < nChild; i++ {
		cmd := exec.Command(os.Args[0],
			"--test.run=TestGRPCPool",
			"--test_peer_child",
			"--test_peer_addrs="+strings.Join(childAddr, ","),
			"--test_peer_index="+strconv.Itoa(i),
		)
		cmds = append(cmds, cmd)
		wg.Add(1)
		if err := cmd.Start(); err != nil {
			t.Fatal("failed to start child process: ", err)
		}
		go awaitAddrReady(t, childAddr[i], &wg)
	}
	defer func() {
		for i := 0; i < nChild; i++ {
			if cmds[i].Process != nil {
				cmds[i].Process.Kill()
			}
		}
	}()
	wg.Wait()

	// Use a dummy self address so that we don't handle gets in-process.
	p := NewGRPCPool("should-be-ignored", grpc.NewServer())
	p.Set(childAddr...)

	// Dummy getter function. Gets should go to children only.
	// The only time this process will handle a get is when the
	// children can't be contacted for some reason.
	getter := GetterFunc(func(ctx context.Context, key string, dest Sink) error {
		return errors.New("parent getter called; something's wrong")
	})
	g := NewGroup("grpcPoolTest", 1<<20, getter)

	for _, key := range testKeys(nGets) {
		var value string
		if err := g.Get(nil, key, StringSink(&value)); err != nil {
			t.Fatal(err)
		}
		if suffix := ":" + key; !strings.HasSuffix(value, suffix) {
			t.Errorf("Get(%q) = %q, want value ending in %q", key, value, suffix)
		}
		t.Logf("Get key=%q, value=%q (peer:key)", key, value)
	}
}

func beChildForTestGRPCPool() {
	addrs := strings.Split(*peerAddrs, ",")
	server := grpc.NewServer()

	p := NewGRPCPool(addrs[*peerIndex], server)
	p.Set(addrs...)

	getter := GetterFunc(func(ctx context.Context, key string, dest Sink) error {
		dest.SetString(strconv.Itoa(*peerIndex)+":"+key, time.Time{})
		return nil
	})
	NewGroup("grpcPoolTest", 1<<20, getter)
	lis, err := net.Listen("tcp", addrs[*peerIndex])
	if err != nil {
		log.Fatalf("Failed to listen on %s", addrs[*peerIndex])
	}

	server.Serve(lis)
}
