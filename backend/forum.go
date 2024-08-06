package main

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sameer-gits/CMS/database"
)

type Forum struct {
	ID                  uuid.UUID
	Name                string
	ForumImage          []byte
	Public              bool
	CreatedAt           time.Time
	CreatedByIdentifier uuid.UUID
}

func createForumHandler(w http.ResponseWriter, r *http.Request) {
	var forum Forum
	var errs []error

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	forumName := r.FormValue("forumName")
	forumPublic := r.FormValue("public")

	user, err := userInfoMiddleware(r)
	if err != nil {
		http.Redirect(w, r, "/logout", badCode)
		return
	}

	defer func() {
		if len(errs) > 0 {
			w.WriteHeader(badCode)
			renderHtml(w, forum, errs, "user.html")
		} else if len(errs) == 0 {
			renderHtml(w, forum, errs, "user.html")
		}
	}()

	if countCharacters(forumName) > 128 {
		errs = append(errs, errors.New("forum name should be less than 128 characters"))
		return
	}

	var public bool
	if forumPublic == "true" {
		public = true
	} else {
		public = false
	}

	forum = Forum{
		Name:                forumName,
		CreatedByIdentifier: user.Identifier,
		Public:              public,
	}

	createForum, err := forum.create(ctx)
	if err != nil {
		errs = append(errs, errors.New("error creating forum, try again"))
		return
	}

	forum = createForum
}

func (forum Forum) create(ctx context.Context) (Forum, error) {
	var result Forum

	tx, err := database.Dbpool.Begin(ctx)
	if err != nil {
		return Forum{}, err
	}
	defer tx.Rollback(ctx)

	insertForum := `INSERT INTO forums (forum_name, public, created_by_identifier)
                    VALUES ($1, $2, $3)
                    RETURNING forum_id, forum_name, forum_image, public, created_at, created_by_identifier`
	err = tx.QueryRow(ctx, insertForum, forum.Name, forum.Public, forum.CreatedByIdentifier).Scan(
		&result.ID, &result.Name, &result.ForumImage, &result.Public, &result.CreatedAt, &result.CreatedByIdentifier)
	if err != nil {
		return Forum{}, err
	}

	insertForumUser := `INSERT INTO forum_users (user_identifier, forum_id) VALUES ($1, $2)`
	_, err = tx.Exec(ctx, insertForumUser, forum.CreatedByIdentifier, result.ID)
	if err != nil {
		return Forum{}, err
	}

	insertForumAdmin := `INSERT INTO forum_admins (user_identifier, forum_id) VALUES ($1, $2)`
	_, err = tx.Exec(ctx, insertForumAdmin, forum.CreatedByIdentifier, result.ID)
	if err != nil {
		return Forum{}, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return Forum{}, err
	}

	return result, nil
}

func viewForumHandler(w http.ResponseWriter, r *http.Request) {
	var errs []error
	var forum_user struct {
		forumData Forum
		User      DbUser
	}

	defer func() {
		if len(errs) > 0 {
			http.Redirect(w, r, "/404", notFound)
		} else if len(errs) == 0 {
			renderHtml(w, forum_user, errs, "forum.html")
		}
	}()

	user, err := userInfoMiddleware(r)
	if err != nil {
		http.Redirect(w, r, "/logout", badCode)
		return
	}

	Id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		errs = append(errs, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	forum, err := getForum(ctx, Id)
	if err != nil {
		errs = append(errs, errors.New("forum not found"))
		return
	}
	forum_user = struct {
		forumData Forum
		User      DbUser
	}{
		forumData: forum,
		User:      user,
	}
}

func getForum(ctx context.Context, Id uuid.UUID) (Forum, error) {
	var forum Forum

	getbyId := `
	SELECT forum_id, forum_name, forum_image, public, created_at, created_by_identifier
	FROM forums WHERE forum_id = $1;
	`
	err := database.Dbpool.QueryRow(ctx, getbyId, Id).Scan(
		&forum.ID,
		&forum.Name,
		&forum.ForumImage,
		&forum.Public,
		&forum.CreatedAt,
		&forum.CreatedByIdentifier,
	)

	if err != nil {
		return Forum{}, err
	}

	return forum, nil
}
