package websocket

import (
	"bufio"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"flowforge-automation-backend/pkg/service/execution"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Client struct {
	ID       string
	TenantID string
	RunID    string
	Send     chan []byte
}

type Hub struct {
	mu         sync.RWMutex
	clients    map[string]*Client
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	h := &Hub{
		clients:    make(map[string]*Client),
		register:   make(chan *Client, 64),
		unregister: make(chan *Client, 64),
	}
	go h.run()
	return h
}

func (h *Hub) run() {
	for {
		select {
		case c := <-h.register:
			h.mu.Lock()
			h.clients[c.ID] = c
			h.mu.Unlock()
		case c := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[c.ID]; ok {
				close(c.Send)
				delete(h.clients, c.ID)
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) Register(c *Client) {
	h.register <- c
}

func (h *Hub) Unregister(c *Client) {
	h.unregister <- c
}

func (h *Hub) BroadcastStepUpdate(event execution.StepUpdateEvent) {
	h.broadcast("step_update", event.RunID.String(), event)
}

func (h *Hub) BroadcastRunUpdate(event execution.RunUpdateEvent) {
	h.broadcast("run_update", event.RunID.String(), event)
}

func (h *Hub) broadcast(eventType, runID string, payload interface{}) {
	msg, err := json.Marshal(map[string]interface{}{
		"type":    eventType,
		"payload": payload,
	})
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, c := range h.clients {
		if c.RunID == "" || c.RunID == runID {
			select {
			case c.Send <- msg:
			default:
			}
		}
	}
}

// sse endpoint for real-time run monitoring
func SSEHandler(hub *Hub) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		runID := ctx.Query("run_id", "")
		tenantID := ctx.Locals("tenant_id")
		tenantStr := ""
		if tenantID != nil {
			tenantStr = tenantID.(string)
		}

		ctx.Set("Content-Type", "text/event-stream")
		ctx.Set("Cache-Control", "no-cache")
		ctx.Set("Connection", "keep-alive")

		clientID := uuid.New().String()
		client := &Client{
			ID:       clientID,
			TenantID: tenantStr,
			RunID:    runID,
			Send:     make(chan []byte, 256),
		}
		hub.Register(client)

		ctx.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case msg, ok := <-client.Send:
					if !ok {
						return
					}
					fmt.Fprintf(w, "data: %s\n\n", msg)
					w.Flush()
				case <-ticker.C:
					// keepalive
					fmt.Fprintf(w, ": keepalive\n\n")
					w.Flush()
				}
			}
		})

		hub.Unregister(client)
		return nil
	}
}
