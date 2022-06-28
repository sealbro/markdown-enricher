package websocket

import (
	"context"
	"github.com/google/uuid"
	"markdown-enricher/infrastructure/serializer"
	"markdown-enricher/pkg/closer"
	"sync"
)

type WsClient struct {
	key      string
	ch       chan []byte
	provider *WebSocketProvider
}

func NewClient(provider *WebSocketProvider) *WsClient {
	return &WsClient{
		ch:       make(chan []byte),
		key:      uuid.New().String(),
		provider: provider,
	}
}

func (ws *WsClient) Get() []byte {
	return <-ws.ch
}

func (ws *WsClient) Close() {
	ws.provider.RemoveClient(ws.key)
}

type WebSocketProvider struct {
	mu         sync.Mutex
	clients    map[string]*WsClient
	serializer serializer.Serializer
}

func MakeWebSocketProvider(collection *closer.CloserCollection) *WebSocketProvider {
	provider := &WebSocketProvider{
		mu:         sync.Mutex{},
		clients:    make(map[string]*WsClient),
		serializer: &serializer.JsonSerializer{},
	}

	collection.Add(provider)

	return provider
}

func (p *WebSocketProvider) Close(_ context.Context) error {
	for key := range p.clients {
		p.RemoveClient(key)
	}

	return nil
}

func (p *WebSocketProvider) Send(data any) {
	go func() {
		serialize, err := p.serializer.Serialize(data)
		if err == nil {
			p.mu.Lock()
			for _, client := range p.clients {
				client.ch <- serialize
			}
			p.mu.Unlock()
		}
	}()
}

func (p *WebSocketProvider) AddClient() *WsClient {
	client := NewClient(p)

	p.mu.Lock()
	p.clients[client.key] = client
	p.mu.Unlock()

	return client
}

func (p *WebSocketProvider) RemoveClient(clientKey string) {
	p.mu.Lock()
	if client, ok := p.clients[clientKey]; ok {
		close(client.ch)
		delete(p.clients, clientKey)
	}
	p.mu.Unlock()
}
