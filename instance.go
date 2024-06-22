package vanity

import (
	"context"
	"fmt"
	"sync"
)

type Instance struct {
	instCtx       context.Context
	instCtxCancel context.CancelFunc
	servers       []*Server
	await         sync.WaitGroup
}

func NewInstance() *Instance {
	ctx, cancel := context.WithCancel(context.Background())
	return &Instance{
		instCtx:       ctx,
		instCtxCancel: cancel,
		await:         sync.WaitGroup{},
	}
}

func (i *Instance) Start() error {
	i.servers = append(i.servers, NewServer("test"))
	for _, server := range i.servers {
		if err := server.Init(); err != nil {
			return fmt.Errorf("server init error. %w", err)
		}
	}
	if len(i.servers) == 0 {
		return fmt.Errorf("servers is required")
	}
	return nil
}

func (i *Instance) Stop() error {
	i.instCtxCancel()
	i.await.Wait()
	return nil
}

func (i *Instance) Serve() error {
	if len(i.servers) == 0 {
		return fmt.Errorf("servers is required")
	}
	errors := make(chan error, len(i.servers))
	for _, server := range i.servers {
		i.await.Add(1)
		go func(server *Server, ctx context.Context) {
			defer i.await.Done()
			errors <- server.Serve(ctx)
		}(server, i.instCtx)
	}
	select {
	case err := <-errors:
		return err
	case <-i.instCtx.Done():
		return nil
	}
}
