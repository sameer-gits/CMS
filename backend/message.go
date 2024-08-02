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
	AuthorUsername    string
	AuthorIdentifier  uuid.UUID
	MessageId         uuid.UUID
	ReplyToIdentifier uuid.UUID
	Content           string
	CreatedAt         time.Time
	InTable           rune
	InTableID         uuid.UUID
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

	if len([]rune(inTable)) == 1 {
		inTableRune = []rune(inTable)[0]
	} else {
		errs = append(errs, errors.New("something went wrong try again"))
		return
	}

	var tableType string
	switch inTableRune {
	case 'F':
		tableType = "forums"
	case 'A':
		tableType = "articles"
	case 'P':
		tableType = "polls"
	default:
		errs = append(errs, errors.New("something went wrong table type does not exists"))
		return
	}

	inTableID, err := uuid.Parse(r.PathValue("inTableID"))
	if err != nil {
		errs = append(errs, errors.New("something went wrong try again"))
		return
	}

	exists, err := getInTableID(ctx, inTableID, tableType)
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
		InTableID:         inTableID,
	}

	msg.insertMessage(ctx)

	msgByte, err := json.Marshal(msg)
	if err != nil {
		errs = append(errs, errors.New("something went wrong try refreshing"))
		return
	}

	roomKey := RoomKey{ID: inTableID, RoomType: tableType}

	// add websocket here or if needed add redis for cache
	wsSrv.broadcast(msgByte, roomKey)
}

func (msg Message) insertMessage(ctx context.Context) (Message, error) {
	var result Message
	var insertMsg string

	if msg.ReplyToIdentifier == uuid.Nil {
		insertMsg = `INSERT INTO messages (author, author_id, content, in_table, in_table_id)
                     VALUES ($1, $2, $3, $4, $5)
                     RETURNING message_id, author, author_id, content, created_at, in_table, in_table_id`
		err := database.Dbpool.QueryRow(ctx, insertMsg, msg.AuthorUsername, msg.AuthorIdentifier, msg.Content, msg.InTable, msg.InTableID).Scan(
			&result.MessageId, &result.AuthorUsername, &result.AuthorIdentifier, &result.Content, &result.CreatedAt, &result.InTable, &result.InTableID)
		if err != nil {
			return Message{}, err
		}
	} else {
		insertMsg = `INSERT INTO messages (author, author_id, reply_to, content, in_table, in_table_id)
                     VALUES ($1, $2, $3, $4, $5, $6)
                     RETURNING message_id, author, author_id, reply_to_identifier, content, created_at, in_table, in_table_id`
		err := database.Dbpool.QueryRow(ctx, insertMsg, msg.AuthorUsername, msg.AuthorIdentifier, msg.ReplyToIdentifier, msg.Content, msg.InTable, msg.InTableID).Scan(
			&result.MessageId, &result.AuthorUsername, &result.AuthorIdentifier, &result.ReplyToIdentifier, &result.Content, &result.CreatedAt, &result.InTable, &result.InTableID)
		if err != nil {
			return Message{}, err
		}
	}

	return result, nil
}
