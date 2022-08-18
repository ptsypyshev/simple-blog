package commentrepo

import (
	"context"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/ptsypyshev/simple-blog/internal/models"
	"go.uber.org/zap"
	"strconv"
)

type CommentCreate interface {
	Create(ctx context.Context, comment models.Comment) (int, error)
}

type CommentRead interface {
	Read(ctx context.Context, id int) (*models.Comment, error)
}

type CommentUpdate interface {
	Update(ctx context.Context, comment models.Comment) (*models.Comment, error)
}

type CommentDelete interface {
	Delete(ctx context.Context, id int) error
}

//type UserSearch interface {
//	Search()
//}

type CommentStorage interface {
	CommentCreate
	CommentRead
	CommentUpdate
	CommentDelete
	//UserSearch
}

type Comments struct {
	cs     CommentStorage
	logger *zap.Logger
	tracer opentracing.Tracer
}

func NewComments(c CommentStorage, l *zap.Logger, t opentracing.Tracer) *Comments {
	return &Comments{
		cs:     c,
		logger: l,
		tracer: t,
	}
}

func (c Comments) Create(ctx context.Context, comment models.Comment) (*models.Comment, error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, c.tracer,
		"CommentRepo.Create")
	defer span.Finish()
	span.LogFields(
		log.String("Comment request", comment.String()),
	)
	id, err := c.cs.Create(ctx, comment)
	if err != nil {
		c.logger.Error(fmt.Sprintf(`cannot read comment: %s`, err))
		span.LogFields(log.Error(err))
		return nil, fmt.Errorf("cannot create comment: %w", err)
	}
	comment.Id = id
	span.LogFields(
		log.String("Comment result", comment.String()),
	)
	return &comment, nil
}

func (c Comments) Read(ctx context.Context, id int) (*models.Comment, error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, c.tracer,
		"CommentRepo.Read")
	defer span.Finish()
	span.LogFields(
		log.String("id", strconv.Itoa(id)),
	)
	comment, err := c.cs.Read(ctx, id)
	if err != nil {
		c.logger.Error(fmt.Sprintf(`cannot read comment: %s`, err))
		span.LogFields(log.Error(err))
		return nil, fmt.Errorf("cannot read comment: %w", err)
	}
	span.LogFields(
		log.String("Comment result", comment.String()),
	)
	return comment, nil
}

func (c Comments) Update(ctx context.Context, updateComment models.Comment) (*models.Comment, error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, c.tracer,
		"CommentRepo.Update")
	defer span.Finish()
	span.LogFields(
		log.String("id", strconv.Itoa(updateComment.Id)),
		log.String("updateComment", updateComment.String()),
	)
	comment, err := c.cs.Update(ctx, updateComment)
	if err != nil {
		c.logger.Error(fmt.Sprintf(`cannot update comment: %s`, err))
		span.LogFields(log.Error(err))
		return nil, fmt.Errorf("cannot update comment: %w", err)
	}
	return comment, nil
}

func (c Comments) Delete(ctx context.Context, id int) (*models.Comment, error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, c.tracer,
		"CommentRepo.Delete")
	defer span.Finish()
	span.LogFields(
		log.String("id", strconv.Itoa(id)),
	)
	comment, err := c.cs.Read(ctx, id)
	if err != nil {
		c.logger.Error(fmt.Sprintf(`cannot read comment: %s`, err))
		span.LogFields(log.Error(err))
		return nil, fmt.Errorf("cannot read comment: %w", err)
	}
	span.LogFields(
		log.String("Comment delete", comment.String()),
	)
	return comment, c.cs.Delete(ctx, id)
}
