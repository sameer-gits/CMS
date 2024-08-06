package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sameer-gits/CMS/database"
)

type Message struct {
	AuthorUsername    string    `json:"author_username"`
	AuthorIdentifier  uuid.UUID `json:"author_identifier"`
	MessageId         uuid.UUID `json:"message_id"`
	ReplyToIdentifier uuid.UUID `json:"reply_to_identifier"`
	Content           string    `json:"content"`
	CreatedAt         time.Time `json:"created_at"`
	InTable           rune      `json:"in_table"`
	InTableIdentifier uuid.UUID `json:"in_table_identifier"`
}

func insertMessageHandler(w http.ResponseWriter, r *http.Request) {
	var errs []error
	var msg Message
	var inTableRune rune
	var replyToIdentifierUUID uuid.UUID

	messageContent := r.FormValue("content")
	inTable := r.FormValue("inTable")
	replyToIdentifierForm := r.FormValue("replyToIdentifier")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	defer func() {
		if len(errs) > 0 {
			w.WriteHeader(badCode)
			renderHtml(w, msg, errs, "user.html")
		} else if len(errs) == 0 {
			renderHtml(w, msg, errs, "user.html")
		}
	}()

	user, err := userInfoMiddleware(r)
	if err != nil {
		http.Redirect(w, r, "/logout", badCode)
		return
	}

	if replyToIdentifierForm == "" {
		replyToIdentifierUUID = uuid.Nil
	} else {
		replyToIdentifierUUID, err = uuid.Parse(replyToIdentifierForm)
		if err != nil {
			errs = append(errs, errors.New("something went wrong try again"))
			return
		}

		exists, err := getUserIdentifier(ctx, replyToIdentifierUUID)
		if err != nil {
			errs = append(errs, errors.New("something went wrong in server try again"))
			return
		}

		if !exists {
			errs = append(errs, errors.New("something went wrong user replying to does not exists"))
			return
		}
	}

	if msg.Content == "" {
		errs = append(errs, errors.New("message is empty try again"))
		return
	}

	switch inTable {
	case "forums":
		inTableRune = 'F'
	case "articles":
		inTableRune = 'A'
	case "polls":
		inTableRune = 'P'
	default:
		errs = append(errs, errors.New("something went wrong table type does not exists"))
		return
	}

	InTableIdentifier, err := uuid.Parse(r.FormValue("InTableIdentifier"))
	if err != nil {
		errs = append(errs, errors.New("something went wrong try again"))
		return
	}

	exists, err := getInTableID(ctx, InTableIdentifier, inTable)
	if err != nil {
		errs = append(errs, errors.New("something went wrong in server try again"))
		return
	}

	if !exists {
		errs = append(errs, errors.New("something went wrong table does not exists"))
		return
	}

	msg = Message{
		AuthorUsername:    user.Username,
		AuthorIdentifier:  user.Identifier,
		ReplyToIdentifier: replyToIdentifierUUID,
		Content:           messageContent,
		InTable:           inTableRune,
		InTableIdentifier: InTableIdentifier,
	}

	msg.insertMessage(ctx)

	msgByte, err := json.Marshal(msg)
	if err != nil {
		errs = append(errs, errors.New("something went wrong try refreshing"))
		return
	}

	roomKey := RoomKey{id: InTableIdentifier, roomtype: inTable}

	// add websocket here or if needed add redis for cache
	rmSrv.publishHandler(roomKey, msgByte)
}

func (msg Message) insertMessage(ctx context.Context) (Message, error) {
	var result Message
	var insertMsg string

	if msg.ReplyToIdentifier == uuid.Nil {
		insertMsg = `INSERT INTO messages (author, author_id, content, in_table, in_table_id)
                     VALUES ($1, $2, $3, $4, $5)
                     RETURNING message_id, author, author_id, content, created_at, in_table, in_table_id`
		err := database.Dbpool.QueryRow(ctx, insertMsg, msg.AuthorUsername, msg.AuthorIdentifier, msg.Content, msg.InTable, msg.InTableIdentifier).Scan(
			&result.MessageId, &result.AuthorUsername, &result.AuthorIdentifier, &result.Content, &result.CreatedAt, &result.InTable, &result.InTableIdentifier)
		if err != nil {
			return Message{}, err
		}
	} else {
		insertMsg = `INSERT INTO messages (author, author_id, reply_to, content, in_table, in_table_id)
                     VALUES ($1, $2, $3, $4, $5, $6)
                     RETURNING message_id, author, author_id, reply_to_identifier, content, created_at, in_table, in_table_id`
		err := database.Dbpool.QueryRow(ctx, insertMsg, msg.AuthorUsername, msg.AuthorIdentifier, msg.ReplyToIdentifier, msg.Content, msg.InTable, msg.InTableIdentifier).Scan(
			&result.MessageId, &result.AuthorUsername, &result.AuthorIdentifier, &result.ReplyToIdentifier, &result.Content, &result.CreatedAt, &result.InTable, &result.InTableIdentifier)
		if err != nil {
			return Message{}, err
		}
	}

	return result, nil
}
