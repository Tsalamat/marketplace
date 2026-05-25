package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	ws "github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"student-marketplace/internal/models"
	jwtpkg "student-marketplace/pkg/jwt"
)

// ─── MESSAGES ────────────────────────────────────────────────

type WSEventType string

const (
	EventMessage      WSEventType = "message"
	EventTyping       WSEventType = "typing"
	EventRead         WSEventType = "read"
	EventOnline       WSEventType = "online"
	EventOffline      WSEventType = "offline"
	EventNotification WSEventType = "notification"
	// WebRTC signaling
	EventRTCOffer     WSEventType = "rtc_offer"
	EventRTCAnswer    WSEventType = "rtc_answer"
	EventRTCCandidate WSEventType = "rtc_candidate"
	EventRTCHangup    WSEventType = "rtc_hangup"
	EventRTCCall      WSEventType = "rtc_call" // incoming call notification
	// Location sharing
	EventLocation     WSEventType = "location"
)

type WSEvent struct {
	Type    WSEventType `json:"type"`
	ChatID  string      `json:"chat_id,omitempty"`
	Payload any         `json:"payload"`
}

type SendMessagePayload struct {
	ChatID   string `json:"chat_id"`
	Content  string `json:"content"`
	FileURL  string `json:"file_url,omitempty"`
	FileName string `json:"file_name,omitempty"`
	FileType string `json:"file_type,omitempty"`
}

type TypingPayload struct {
	ChatID   string `json:"chat_id"`
	UserID   string `json:"user_id"`
	IsTyping bool   `json:"is_typing"`
}

// ─── CLIENT ──────────────────────────────────────────────────

type Client struct {
	UserID uuid.UUID
	Conn   *ws.Conn
	Send   chan []byte
	Hub    *Hub
}

func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(ws.CloseMessage, []byte{})
				return
			}
			c.Conn.WriteMessage(ws.TextMessage, message)
		case <-ticker.C:
			if err := c.Conn.WriteMessage(ws.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512 * 1024) // 512KB
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, data, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		var event WSEvent
		if err := json.Unmarshal(data, &event); err != nil {
			continue
		}

		c.Hub.broadcast <- &BroadcastMessage{
			SenderID: c.UserID,
			Event:    event,
			RawData:  data,
		}
	}
}

// ─── HUB ─────────────────────────────────────────────────────

type BroadcastMessage struct {
	SenderID uuid.UUID
	Event    WSEvent
	RawData  []byte
}

type Hub struct {
	clients    map[uuid.UUID]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan *BroadcastMessage
	mu         sync.RWMutex
	db         *gorm.DB
}

func NewHub(db *gorm.DB) *Hub {
	return &Hub{
		clients:    make(map[uuid.UUID]*Client),
		register:   make(chan *Client, 256),
		unregister: make(chan *Client, 256),
		broadcast:  make(chan *BroadcastMessage, 256),
		db:         db,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.UserID] = client
			h.mu.Unlock()
			// Broadcast online status
			h.broadcastPresence(client.UserID, true)
			h.db.Model(&models.Profile{}).Where("user_id = ?", client.UserID).Update("is_online", true)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; ok {
				delete(h.clients, client.UserID)
				close(client.Send)
			}
			h.mu.Unlock()
			h.broadcastPresence(client.UserID, false)
			h.db.Model(&models.Profile{}).Where("user_id = ?", client.UserID).Update("is_online", false)
			h.db.Model(&models.User{}).Where("id = ?", client.UserID).Update("last_active", time.Now())

		case msg := <-h.broadcast:
			h.handleEvent(msg)
		}
	}
}

func (h *Hub) handleEvent(msg *BroadcastMessage) {
	switch msg.Event.Type {
	case EventMessage:
		h.handleMessage(msg)
	case EventTyping:
		h.handleTyping(msg)
	case EventRead:
		h.handleRead(msg)
	case EventRTCOffer, EventRTCAnswer, EventRTCCandidate, EventRTCHangup:
		h.relayRTC(msg)
	case EventLocation:
		h.handleLocation(msg)
	}
}

// relayRTC forwards WebRTC signaling directly to target peer
func (h *Hub) relayRTC(msg *BroadcastMessage) {
	raw, _ := json.Marshal(msg.Event.Payload)
	var p struct {
		TargetID string `json:"target_id"`
	}
	if err := json.Unmarshal(raw, &p); err != nil || p.TargetID == "" {
		return
	}
	targetID, err := uuid.Parse(p.TargetID)
	if err != nil {
		return
	}
	// Attach sender_id to payload before forwarding
	var full map[string]any
	json.Unmarshal(raw, &full)
	full["sender_id"] = msg.SenderID.String()
	event := WSEvent{Type: msg.Event.Type, Payload: full}
	// Notify caller about incoming call
	if msg.Event.Type == EventRTCOffer {
		h.SendToUser(targetID, WSEvent{Type: EventRTCCall, Payload: full})
	}
	h.SendToUser(targetID, event)
}

// handleLocation updates and broadcasts user location to friends
func (h *Hub) handleLocation(msg *BroadcastMessage) {
	raw, _ := json.Marshal(msg.Event.Payload)
	var p struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return
	}
	// Persist to DB
	h.db.Model(&models.Profile{}).Where("user_id = ?", msg.SenderID).
		Updates(map[string]any{"lat": p.Lat, "lng": p.Lng})

	// Broadcast to online friends
	var friendships []models.Friendship
	h.db.Where("(requester_id=? OR addressee_id=?) AND status='accepted'", msg.SenderID, msg.SenderID).
		Find(&friendships)

	event := WSEvent{
		Type: EventLocation,
		Payload: map[string]any{
			"user_id": msg.SenderID.String(),
			"lat":     p.Lat,
			"lng":     p.Lng,
		},
	}
	data, _ := json.Marshal(event)
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, f := range friendships {
		fID := f.AddresseeID
		if f.AddresseeID == msg.SenderID {
			fID = f.RequesterID
		}
		if client, ok := h.clients[fID]; ok {
			select {
			case client.Send <- data:
			default:
			}
		}
	}
}

func (h *Hub) handleMessage(msg *BroadcastMessage) {
	raw, _ := json.Marshal(msg.Event.Payload)
	var payload SendMessagePayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return
	}

	chatID, err := uuid.Parse(payload.ChatID)
	if err != nil {
		return
	}

	// Verify sender is participant
	var participant models.ChatParticipant
	if err := h.db.Where("chat_id = ? AND user_id = ?", chatID, msg.SenderID).
		First(&participant).Error; err != nil {
		return
	}

	// Save message
	message := models.ChatMessage{
		ChatID:   chatID,
		SenderID: msg.SenderID,
		Content:  payload.Content,
		FileURL:  payload.FileURL,
		FileName: payload.FileName,
		FileType: payload.FileType,
	}
	h.db.Create(&message)
	h.db.Preload("Sender.Profile").First(&message, message.ID)

	// Update chat timestamp
	h.db.Model(&models.Chat{}).Where("id = ?", chatID).Update("updated_at", time.Now())

	// Get all participants
	var participants []models.ChatParticipant
	h.db.Where("chat_id = ? AND user_id != ?", chatID, msg.SenderID).Find(&participants)

	event := WSEvent{
		Type:    EventMessage,
		ChatID:  payload.ChatID,
		Payload: message,
	}
	data, _ := json.Marshal(event)

	// Send to all participants
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, p := range participants {
		if client, ok := h.clients[p.UserID]; ok {
			select {
			case client.Send <- data:
			default:
			}
		} else {
			// Create notification for offline user
			h.db.Create(&models.Notification{
				UserID: p.UserID,
				Type:   models.NotifMessage,
				Title:  "New Message",
				Body:   "You have a new message",
				Data:   map[string]any{"chat_id": chatID},
			})
		}
	}
	// Echo message back to sender so they see it in real-time
	if senderClient, ok := h.clients[msg.SenderID]; ok {
		select {
		case senderClient.Send <- data:
		default:
		}
	}
}

func (h *Hub) handleTyping(msg *BroadcastMessage) {
	raw, _ := json.Marshal(msg.Event.Payload)
	var payload TypingPayload
	json.Unmarshal(raw, &payload)

	chatID, err := uuid.Parse(payload.ChatID)
	if err != nil {
		return
	}

	var participants []models.ChatParticipant
	h.db.Where("chat_id = ? AND user_id != ?", chatID, msg.SenderID).Find(&participants)

	event := WSEvent{
		Type:   EventTyping,
		ChatID: payload.ChatID,
		Payload: TypingPayload{
			ChatID:   payload.ChatID,
			UserID:   msg.SenderID.String(),
			IsTyping: payload.IsTyping,
		},
	}
	data, _ := json.Marshal(event)

	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, p := range participants {
		if client, ok := h.clients[p.UserID]; ok {
			select {
			case client.Send <- data:
			default:
			}
		}
	}
}

func (h *Hub) handleRead(msg *BroadcastMessage) {
	raw, _ := json.Marshal(msg.Event.Payload)
	var payload struct {
		ChatID string `json:"chat_id"`
	}
	json.Unmarshal(raw, &payload)
	chatID, _ := uuid.Parse(payload.ChatID)
	h.db.Model(&models.ChatParticipant{}).
		Where("chat_id = ? AND user_id = ?", chatID, msg.SenderID).
		Update("last_read", time.Now())
}

func (h *Hub) broadcastPresence(userID uuid.UUID, online bool) {
	eventType := EventOnline
	if !online {
		eventType = EventOffline
	}
	event := WSEvent{
		Type:    eventType,
		Payload: map[string]string{"user_id": userID.String()},
	}
	data, _ := json.Marshal(event)

	h.mu.RLock()
	defer h.mu.RUnlock()
	for id, client := range h.clients {
		if id != userID {
			select {
			case client.Send <- data:
			default:
			}
		}
	}
}

// SendToUser sends a notification to a specific user if online
func (h *Hub) SendToUser(userID uuid.UUID, event WSEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	if client, ok := h.clients[userID]; ok {
		select {
		case client.Send <- data:
		default:
			log.Printf("failed to send to user %s: buffer full", userID)
		}
	}
}

// ─── FIBER HANDLER ───────────────────────────────────────────

func (h *Hub) Handler() fiber.Handler {
	return ws.New(func(c *ws.Conn) {
		userID := c.Locals("user_id").(uuid.UUID)

		client := &Client{
			UserID: userID,
			Conn:   c,
			Send:   make(chan []byte, 256),
			Hub:    h,
		}

		h.register <- client

		go client.writePump()
		client.readPump()
	})
}

// Upgrade middleware — validates JWT from query param (WS can't set headers)
func WSUpgradeMiddleware(c *fiber.Ctx) error {
	if !ws.IsWebSocketUpgrade(c) {
		return fiber.ErrUpgradeRequired
	}

	// WebSocket clients pass token as query param since they can't set headers
	token := c.Query("token")
	if token == "" {
		// fallback: cookie or Authorization header
		if cookie := c.Cookies("access_token"); cookie != "" {
			token = cookie
		} else if auth := c.Get("Authorization"); len(auth) > 7 && auth[:7] == "Bearer " {
			token = auth[7:]
		}
	}
	if token == "" {
		return fiber.ErrUnauthorized
	}

	claims, err := jwtpkg.ValidateAccessToken(token)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid or expired token")
	}

	c.Locals("user_id", claims.UserID)
	return c.Next()
}
