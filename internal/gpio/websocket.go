package internal

import (
    "encoding/json"
    "sync"

    "github.com/fasthttp/websocket"
    "github.com/gofiber/fiber/v2"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    wsConnections = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "active_websocket_connections",
        Help: "Number of active WebSocket connections",
    })
)

type WebSocketManager struct {
    upgrader    websocket.FastHTTPUpgrader
    gpio       *GPIOManager
    clients    map[*websocket.Conn]bool
    mu         sync.RWMutex
}

func NewWebSocketManager(gpio *GPIOManager) *WebSocketManager {
    wsm := &WebSocketManager{
        upgrader: websocket.FastHTTPUpgrader{
            ReadBufferSize:  1024,
            WriteBufferSize: 1024,
        },
        gpio:    gpio,
        clients: make(map[*websocket.Conn]bool),
    }

    gpio.RegisterCallback(wsm.broadcastPinChange)
    return wsm
}

func (wsm *WebSocketManager) HandleWebSocket(c *fiber.Ctx) error {
    return wsm.upgrader.Upgrade(c.Context(), func(conn *websocket.Conn) {
        wsm.mu.Lock()
        wsm.clients[conn] = true
        wsm.mu.Unlock()
        wsConnections.Inc()

        defer func() {
            wsm.mu.Lock()
            delete(wsm.clients, conn)
            wsm.mu.Unlock()
            wsConnections.Dec()
            conn.Close()
        }()

        for {
            messageType, message, err := conn.ReadMessage()
            if err != nil {
                return
            }

            if messageType == websocket.TextMessage {
                var req struct {
                    Action string `json:"action"`
                    Pin    int    `json:"pin"`
                    Value  bool   `json:"value,omitempty"`
                }

                if err := json.Unmarshal(message, &req); err != nil {
                    wsm.sendError(conn, "Invalid JSON format")
                    continue
                }

                switch req.Action {
                case "write":
                    if err := wsm.gpio.WritePin(req.Pin, req.Value); err != nil {
                        wsm.sendError(conn, err.Error())
                        continue
                    }
                case "read":
                    value, err := wsm.gpio.ReadPin(req.Pin)
                    if err != nil {
                        wsm.sendError(conn, err.Error())
                        continue
                    }
                    wsm.sendResponse(conn, "read", req.Pin, value)
                }
            }
        }
    })
}

func (wsm *WebSocketManager) broadcastPinChange(pin int, value bool) {
    wsm.mu.RLock()
    defer wsm.mu.RUnlock()

    for conn := range wsm.clients {
        wsm.sendResponse(conn, "pin_change", pin, value)
    }
}

func (wsm *WebSocketManager) sendError(conn *websocket.Conn, message string) {
    response := struct {
        Status  string `json:"status"`
        Error   string `json:"error"`
    }{
        Status:  "error",
        Error:   message,
    }
    conn.WriteJSON(response)
}

func (wsm *WebSocketManager) sendResponse(conn *websocket.Conn, action string, pin int, value bool) {
    response := struct {
        Status  string `json:"status"`
        Action  string `json:"action"`
        Pin     int    `json:"pin"`
        Value   bool   `json:"value"`
    }{
        Status:  "success",
        Action:  action,
        Pin:     pin,
        Value:   value,
    }
    conn.WriteJSON(response)
}
