package commentstore

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/ptsypyshev/simple-blog/internal/db/pgdb"
	"github.com/ptsypyshev/simple-blog/internal/models"
	"github.com/ptsypyshev/simple-blog/internal/repositories/commentrepo"
	"go.uber.org/zap"
	"strconv"
)

const (
	CommentCreate = `
INSERT INTO comments(body, user_id, post_id)
VALUES
    ($1, $2, $3)
RETURNING id;
`
	CommentSelectByID = `SELECT * FROM comments WHERE id = $1;`
	CommentDeleteByID = `
DELETE FROM comments WHERE id = $1;
`
)

var _ commentrepo.CommentStorage = &CommentsDB{}

type CommentsDB struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
	tracer opentracing.Tracer
}

func NewCommentsDB(p *pgxpool.Pool, l *zap.Logger, t opentracing.Tracer) *CommentsDB {
	return &CommentsDB{
		pool:   p,
		logger: l,
		tracer: t,
	}
}

func (db *CommentsDB) Create(ctx context.Context, comment models.Comment) (int, error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, db.tracer,
		"CommentStore.Create")
	defer span.Finish()
	span.LogFields(
		log.String("query", CommentCreate),
		log.String("arg0", comment.String()),
	)
	var id int
	res := db.pool.QueryRow(
		ctx, CommentCreate, comment.Body, comment.UserId, comment.PostId,
	)
	err := res.Scan(&id)
	if err != nil {
		span.LogFields(log.Error(err))
		return 0, err
	}
	span.LogFields(
		log.String("Comment result", comment.String()),
	)
	return id, nil
}

func (db *CommentsDB) Read(ctx context.Context, id int) (*models.Comment, error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, db.tracer,
		"CommentStore.Read")
	defer span.Finish()
	span.LogFields(
		log.String("query", CommentSelectByID),
		log.String("arg0", strconv.Itoa(id)),
	)
	rows, _ := db.pool.Query(ctx, CommentSelectByID, id)
	var (
		comment models.Comment
		found   bool
	)
	for rows.Next() {
		if found {
			err := fmt.Errorf("%w: comment id %d", pgdb.ErrMultipleFound, id)
			span.LogFields(log.Error(err))
			return nil, err
		}
		if err := rows.Scan(&comment.Id, &comment.Date, &comment.Body, &comment.UserId, &comment.PostId); err != nil {
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
		err := fmt.Errorf("%w: comment id %d", pgdb.ErrNotFound, id)
		span.LogFields(log.Error(err))
		return nil, err
	}
	span.LogFields(
		log.String("Comment result", comment.String()),
	)
	return &comment, nil
}

func (db *CommentsDB) Update(ctx context.Context, comment models.Comment) (*models.Comment, error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, db.tracer,
		"CommentStore.Update")
	defer span.Finish()
	UpdateQuery, err := pgdb.UpdateQueryCompilation("comments", comment, models.Comment{})
	if err != nil {
		err = fmt.Errorf("cannot compile query: %w", err)
		span.LogFields(log.Error(err))
		return &models.Comment{}, err
	}
	span.LogFields(
		log.String("query", UpdateQuery),
		log.String("arg0", comment.String()),
	)
	res, err := db.pool.Exec(ctx, UpdateQuery)
	if err != nil {
		span.LogFields(log.Error(err))
		return &models.Comment{}, err
	}

	if rowsAffected := res.RowsAffected(); rowsAffected != 1 {
		err = fmt.Errorf("update comment error: %d rows affected", rowsAffected)
		span.LogFields(log.Error(err))
		return &models.Comment{}, err
	}
	span.LogFields(
		log.String("Comment result", comment.String()),
	)
	return &comment, nil
}

func (db *CommentsDB) Delete(ctx context.Context, id int) error {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, db.tracer,
		"CommentStore.Delete")
	defer span.Finish()
	span.LogFields(
		log.String("query", CommentDeleteByID),
		log.String("arg0", strconv.Itoa(id)),
	)
	res, err := db.pool.Exec(ctx, CommentDeleteByID, id)
	if err != nil {
		span.LogFields(log.Error(err))
		return err
	}
	if rowsAffected := res.RowsAffected(); rowsAffected != 1 {
		err = fmt.Errorf("delete comment error: %d rows affected", rowsAffected)
		span.LogFields(log.Error(err))
		return err
	}
	span.LogFields(
		log.String("Deleted comment with id", strconv.Itoa(id)),
	)
	return nil
}
