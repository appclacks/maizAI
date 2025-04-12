package mcp

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

type Manager struct {
	clients map[string]*Client
	lock    sync.RWMutex
}

func (m *Manager) Start() {

	go func() {
		for k, client := range m.clients {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			err := client.Ping(ctx)
			cancel()
			if err != nil {
				slog.Error(fmt.Sprintf("fail to ping mcp client %s: %s", err.Error()))
				m.lock.Lock()
				err := client.Close()
				if err != nil {
					slog.Error(fmt.Sprintf("fail to close mcp client %s: %s", err.Error()))
				}
				m.lock.Unlock()
				// todo: remove from map ? reconnect somehow?
				cl
				continue
			}
		}
	}()

}
