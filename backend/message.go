package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sameer-gits/CMS/database"
)

type message struct {
	Author            string
	AuthorId          uuid.UUID
	MessageId         uuid.UUID
	ReplyToIdentifier uuid.UUID
	Content           string
	CreatedAt         time.Time
	InTable           rune
	InTableId         uuid.UUID
}

func insertMessageHandler(w http.ResponseWriter, r *http.Request) {
	user, err := userInfoMiddleware(r)
	if err != nil {
		http.Redirect(w, r, "/logout", badCode)
		return
	}

	// insert message and from now only use user.Identifier for user query
	fmt.Print(user.Identifier)
}

func (msg message) insertMessage(ctx context.Context) (message, error) {
	var result message
	var insertMsg string

	if msg.ReplyToIdentifier == uuid.Nil {
		insertMsg = `INSERT INTO messages (author, author_id, content, in_table, in_table_id)
                     VALUES ($1, $2, $3, $4, $5)
                     RETURNING message_id, author, author_id, content, created_at, in_table, in_table_id`
		err := database.Dbpool.QueryRow(ctx, insertMsg, msg.Author, msg.AuthorId, msg.Content, msg.InTable, msg.InTableId).Scan(
			&result.MessageId, &result.Author, &result.AuthorId, &result.Content, &result.CreatedAt, &result.InTable, &result.InTableId)
		if err != nil {
			return message{}, err
		}
	} else {
		insertMsg = `INSERT INTO messages (author, author_id, reply_to, content, in_table, in_table_id)
                     VALUES ($1, $2, $3, $4, $5, $6)
                     RETURNING message_id, author, author_id, reply_to_identifier, content, created_at, in_table, in_table_id`
		err := database.Dbpool.QueryRow(ctx, insertMsg, msg.Author, msg.AuthorId, msg.Content, msg.InTable, msg.InTableId).Scan(
			&result.MessageId, &result.Author, &result.AuthorId, &result.ReplyToIdentifier, &result.Content, &result.CreatedAt, &result.InTable, &result.InTableId)
		if err != nil {
			return message{}, err
		}
	}

	return result, nil
}
