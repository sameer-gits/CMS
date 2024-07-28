package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
	"nhooyr.io/websocket"
)

type roomKey struct {
	roomID   string
	roomType string
}

type webstocketServer struct {
	roomsMu sync.RWMutex
	rooms   map[string]*chatRoom
	logf    func(f string, v ...interface{})
}

type chatRoom struct {
	subscriberMessageBuffer int
	publishLimiter          *rate.Limiter
	logf                    func(f string, v ...interface{})
	subscribersMu           sync.RWMutex
	subscribers             map[*subscriber]struct{}
}

type subscriber struct {
	msgs      chan []byte
	closeSlow func()
}

func newServer() *webstocketServer {
	return &webstocketServer{
		rooms: make(map[string]*chatRoom),
		logf:  log.Printf,
	}
}

func (s *webstocketServer) getRoom(roomKey roomKey) *chatRoom {
	s.roomsMu.Lock()
	key := fmt.Sprintf("%s_%s", roomKey.roomType, roomKey.roomID)
	defer s.roomsMu.Unlock()

	room, ok := s.rooms[key]
	if !ok {
		room = &chatRoom{
			subscriberMessageBuffer: 16,
			logf:                    s.logf,
			subscribers:             make(map[*subscriber]struct{}),
			publishLimiter:          rate.NewLimiter(rate.Every(time.Millisecond*100), 8),
		}
		s.rooms[key] = room
	}

	return room
}

func (s *webstocketServer) publishHandler(w http.ResponseWriter, r *http.Request) {
	roomKey := roomKey{
		roomID:   r.PathValue("room_id"),
		roomType: r.PathValue("room_type"),
	}

	if roomKey.roomType == "" || roomKey.roomID == "" {
		http.Error(w, "room is required", badCode)
		return
	}

	cr := s.getRoom(roomKey)
	cr.publishHandler(w, r)
}

func (s *webstocketServer) subscribeHandler(w http.ResponseWriter, r *http.Request) {
	roomKey := roomKey{
		roomID:   r.PathValue("room_id"),
		roomType: r.PathValue("room_type"),
	}

	if roomKey.roomType == "" || roomKey.roomID == "" {
		http.Error(w, "room is required", badCode)
		return
	}

	cr := s.getRoom(roomKey)
	err := cr.subscribe(r.Context(), w, r)
	if errors.Is(err, context.Canceled) {
		return
	}
	if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
		websocket.CloseStatus(err) == websocket.StatusGoingAway {
		return
	}
	if err != nil {
		s.logf("subscribe error: %v", err)
		return
	}
}

func (cr *chatRoom) publishHandler(w http.ResponseWriter, r *http.Request) {
	body := http.MaxBytesReader(w, r.Body, 8192)
	msg, err := io.ReadAll(body)
	if err != nil {
		http.Error(w, "request too large", http.StatusRequestEntityTooLarge)
		return
	}

	cr.publish(msg)

	w.WriteHeader(statusAccepted)
}

func (cr *chatRoom) publish(msg []byte) {
	cr.subscribersMu.Lock()
	defer cr.subscribersMu.Unlock()

	cr.publishLimiter.Wait(context.Background())

	for s := range cr.subscribers {
		select {
		case s.msgs <- msg:
		default:
			go s.closeSlow()
		}
	}
}

func (cr *chatRoom) subscribe(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var mu sync.Mutex
	var c *websocket.Conn
	var closed bool
	s := &subscriber{
		msgs: make(chan []byte, cr.subscriberMessageBuffer),
		closeSlow: func() {
			mu.Lock()
			defer mu.Unlock()
			closed = true
			if c != nil {
				c.Close(websocket.StatusPolicyViolation, "connection too slow to keep up with messages")
			}
		},
	}
	cr.addSubscriber(s)
	defer cr.deleteSubscriber(s)

	c2, err := websocket.Accept(w, r, nil)
	if err != nil {
		return err
	}
	mu.Lock()
	if closed {
		mu.Unlock()
		return net.ErrClosed
	}
	c = c2
	mu.Unlock()
	defer c.CloseNow()

	ctx = c.CloseRead(ctx)

	for {
		select {
		case msg := <-s.msgs:
			err := writeTimeout(ctx, time.Second*5, c, msg)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (cr *chatRoom) addSubscriber(s *subscriber) {
	cr.subscribersMu.Lock()
	cr.subscribers[s] = struct{}{}
	cr.subscribersMu.Unlock()
}

func (cr *chatRoom) deleteSubscriber(s *subscriber) {
	cr.subscribersMu.Lock()
	delete(cr.subscribers, s)
	cr.subscribersMu.Unlock()
}

func writeTimeout(ctx context.Context, timeout time.Duration, c *websocket.Conn, msg []byte) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return c.Write(ctx, websocket.MessageText, msg)
}
