package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/google/uuid"
)

type RoomKey struct {
	ID       uuid.UUID
	RoomType string
}

type Subscriber struct {
	Conn net.Conn
	Room *Room
}

type Room struct {
	Subscribers map[*Subscriber]bool
	Mu          sync.RWMutex
}

type WebsocketServer struct {
	Rooms map[RoomKey]*Room
	Mu    sync.RWMutex
}

func newWebsocketServer() *WebsocketServer {
	return &WebsocketServer{
		Rooms: make(map[RoomKey]*Room),
	}
}

func (s *WebsocketServer) websocketServerHandler(w http.ResponseWriter, r *http.Request) {
	var errs []error

	defer func() {
		if len(errs) > 0 {
			w.WriteHeader(badCode)
			renderHtml(w, nil, errs, "user.html")
		} else if len(errs) == 0 {
			renderHtml(w, nil, errs, "user.html")
		}
	}()

	roomID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		errs = append(errs, fmt.Errorf("error uuid: %w", err))
		return
	}

	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		errs = append(errs, fmt.Errorf("error upgrading conn: %w", err))
		return
	}
	roomKey := RoomKey{
		ID:       roomID,
		RoomType: r.PathValue("roomtype"),
	}
	err = s.subscribe(conn, roomKey)
	if err != nil {
		errs = append(errs, fmt.Errorf("conn error: %w", err))
		return
	}
}

func (s *WebsocketServer) subscribe(conn net.Conn, roomKey RoomKey) error {
	s.Mu.Lock()

	room, exists := s.Rooms[roomKey]
	if !exists {
		room = &Room{
			Subscribers: make(map[*Subscriber]bool),
		}
		s.Rooms[roomKey] = room
	}

	s.Mu.Unlock()
	subscriber := &Subscriber{
		Conn: conn,
		Room: room,
	}
	room.Mu.Lock()
	room.Subscribers[subscriber] = true
	room.Mu.Unlock()

	err := make(chan error)
	go func() {
		err <- s.listener(subscriber, roomKey)
	}()

	log.Printf("subscriber: %v, broadcast err: %v\n", subscriber, err)
	return nil
}

func (s *WebsocketServer) listener(subscriber *Subscriber, roomKey RoomKey) error {
	defer s.unSubscribe(subscriber)
	for {
		msg, op, err := wsutil.ReadClientData(subscriber.Conn)
		if err != nil {
			if err != io.EOF {
				return fmt.Errorf("listen conn error: %w", err)
			}
		}
		if op == ws.OpClose {
			return nil
		}
		s.broadcast(msg, roomKey)
	}
}

func (s *WebsocketServer) broadcast(message []byte, roomKey RoomKey) error {
	s.Mu.RLock()
	room, exists := s.Rooms[roomKey]
	s.Mu.RUnlock()

	if !exists {
		return fmt.Errorf("room not found")
	}

	room.Mu.RLock()
	defer room.Mu.RUnlock()
	for subscriber := range room.Subscribers {
		err := wsutil.WriteServerMessage(subscriber.Conn, ws.OpText, message)
		if err != nil {
			log.Printf("subscriber: %v, broadcast err: %v\n", subscriber, err)
			s.unSubscribe(subscriber)
		}
	}
	return nil
}
func (s *WebsocketServer) unSubscribe(subscriber *Subscriber) {
	subscriber.Room.Mu.Lock()
	delete(subscriber.Room.Subscribers, subscriber)
	subscriber.Room.Mu.Unlock()
	subscriber.Conn.Close()
}
