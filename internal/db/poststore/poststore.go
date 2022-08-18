package poststore

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/ptsypyshev/simple-blog/internal/db/pgdb"
	"github.com/ptsypyshev/simple-blog/internal/models"
	"github.com/ptsypyshev/simple-blog/internal/repositories/postrepo"
	"go.uber.org/zap"
	"strconv"
)

const (
	PostCreate = `
INSERT INTO posts(title, body, user_id)
VALUES
    ($1, $2, $3)
RETURNING id;
`
	PostSelectByID = `SELECT * FROM posts WHERE id = $1;`
	PostDeleteByID = `
DELETE FROM posts WHERE id = $1;
`
)

var _ postrepo.PostStorage = &PostsDB{}

type PostsDB struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
	tracer opentracing.Tracer
}

func NewPostsDB(p *pgxpool.Pool, l *zap.Logger, t opentracing.Tracer) *PostsDB {
	return &PostsDB{
		pool:   p,
		logger: l,
		tracer: t,
	}
}

func (db *PostsDB) Create(ctx context.Context, post models.Post) (int, error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, db.tracer,
		"PostStore.Create")
	defer span.Finish()
	span.LogFields(
		log.String("query", PostCreate),
		log.String("arg0", post.String()),
	)
	var id int
	res := db.pool.QueryRow(
		ctx, PostCreate, post.Title, post.Body, post.UserId,
	)
	err := res.Scan(&id)
	if err != nil {
		span.LogFields(log.Error(err))
		return 0, err
	}
	span.LogFields(
		log.String("Post result", post.String()),
	)
	return id, nil
}

func (db *PostsDB) Read(ctx context.Context, id int) (*models.Post, error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, db.tracer,
		"PostStore.Read")
	defer span.Finish()
	span.LogFields(
		log.String("query", PostSelectByID),
		log.String("arg0", strconv.Itoa(id)),
	)
	rows, _ := db.pool.Query(ctx, PostSelectByID, id)
	var (
		post  models.Post
		found bool
	)
	for rows.Next() {
		if found {
			err := fmt.Errorf("%w: post id %d", pgdb.ErrMultipleFound, id)
			span.LogFields(log.Error(err))
			return nil, err
		}
		if err := rows.Scan(&post.Id, &post.Title, &post.Body, &post.UserId); err != nil {
			span.LogFields(log.Error(err))
			return nil, err
		}
		found = true
	}
	if err := rows.Err(); err != nil {
		span.LogFields(log.Error(err))
		return nil, err
	}
	if !found {
		err := fmt.Errorf("%w: post id %d", pgdb.ErrNotFound, id)
		span.LogFields(log.Error(err))
		return nil, err
	}
	span.LogFields(
		log.String("Post result", post.String()),
	)
	return &post, nil
}

func (db *PostsDB) Update(ctx context.Context, post models.Post) (*models.Post, error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, db.tracer,
		"PostStore.Update")
	defer span.Finish()
	UpdateQuery, err := pgdb.UpdateQueryCompilation("posts", post, models.Post{})
	if err != nil {
		err = fmt.Errorf("cannot compile query: %w", err)
		span.LogFields(log.Error(err))
		return &models.Post{}, err
	}
	span.LogFields(
		log.String("query", UpdateQuery),
		log.String("arg0", post.String()),
	)
	res, err := db.pool.Exec(ctx, UpdateQuery)
	if err != nil {
		span.LogFields(log.Error(err))
		return &models.Post{}, err
	}

	if rowsAffected := res.RowsAffected(); rowsAffected != 1 {
		err = fmt.Errorf("update post error: %d rows affected", rowsAffected)
		span.LogFields(log.Error(err))
		return &models.Post{}, err
	}
	span.LogFields(
		log.String("Post result", post.String()),
	)
	return &post, nil
}

func (db *PostsDB) Delete(ctx context.Context, id int) error {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, db.tracer,
		"PostStore.Delete")
	defer span.Finish()
	span.LogFields(
		log.String("query", PostDeleteByID),
		log.String("arg0", strconv.Itoa(id)),
	)
	res, err := db.pool.Exec(ctx, PostDeleteByID, id)
	if err != nil {
		span.LogFields(log.Error(err))
		return err
	}
	if rowsAffected := res.RowsAffected(); rowsAffected != 1 {
		err = fmt.Errorf("delete post error: %d rows affected", rowsAffected)
		span.LogFields(log.Error(err))
		return err
	}
	span.LogFields(
		log.String("Deleted post with id", strconv.Itoa(id)),
	)
	return nil
}
