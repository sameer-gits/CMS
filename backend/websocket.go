package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/lesismal/nbio/nbhttp/websocket"
)

type Room struct {
	subscribers map[*Subscriber]bool
	mu          sync.RWMutex
}

type RoomKey struct {
	id       uuid.UUID
	roomtype string
}

type Subscriber struct {
	conn    *websocket.Conn
	roomKey RoomKey
}

type RoomManager struct {
	rooms map[RoomKey]*Room
	mu    sync.RWMutex
}

func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[RoomKey]*Room),
	}
}

func (rm *RoomManager) publishHandler(key RoomKey, data []byte) {
	rm.mu.RLock()
	room, exists := rm.rooms[key]
	rm.mu.RUnlock()
	if !exists {
		return
	}

	room.mu.RLock()
	defer room.mu.RUnlock()
	for s := range room.subscribers {
		go func(s *Subscriber) {
			err := s.conn.WriteMessage(websocket.BinaryMessage, data)
			if err != nil {
				log.Println(s, err)
			}
		}(s)
	}
}

func (rm *RoomManager) subscribeHandler(w http.ResponseWriter, r *http.Request) {
	Id, err := uuid.Parse(r.PathValue("id"))
	rmType := r.PathValue("type")
	if err != nil || rmType == "" {
		http.Redirect(w, r, "/404", notFound)
		return
	}

	u := websocket.NewUpgrader()
	conn, err := u.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	key := RoomKey{
		id:       Id,
		roomtype: rmType,
	}
	rm.subscribe(key, conn)
}

func (rm *RoomManager) subscribe(key RoomKey, conn *websocket.Conn) {
	rm.mu.Lock()
	room, exists := rm.rooms[key]
	if !exists {
		room = &Room{
			subscribers: make(map[*Subscriber]bool),
		}
		rm.rooms[key] = room
	}
	rm.mu.Unlock()

	room.mu.Lock()
	defer room.mu.Unlock()
	subscriber := &Subscriber{
		conn:    conn,
		roomKey: key,
	}
	room.subscribers[subscriber] = true
	log.Println("room:", room)

	conn.OnClose(func(conn *websocket.Conn, err error) {
		if err != nil {
			log.Println("Connection closed with error:", err)
		}
		room.mu.Lock()
		defer room.mu.Unlock()
		delete(room.subscribers, subscriber)
	})
}
