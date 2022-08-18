package postrepo

import (
	"context"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/ptsypyshev/simple-blog/internal/models"
	"go.uber.org/zap"
	"strconv"
)

type PostCreate interface {
	Create(ctx context.Context, post models.Post) (int, error)
}

type PostRead interface {
	Read(ctx context.Context, id int) (*models.Post, error)
}

type PostUpdate interface {
	Update(ctx context.Context, post models.Post) (*models.Post, error)
}

type PostDelete interface {
	Delete(ctx context.Context, id int) error
}

//type UserSearch interface {
//	Search()
//}

type PostStorage interface {
	PostCreate
	PostRead
	PostUpdate
	PostDelete
	//UserSearch
}

type Posts struct {
	ps     PostStorage
	logger *zap.Logger
	tracer opentracing.Tracer
}

func NewPosts(p PostStorage, l *zap.Logger, t opentracing.Tracer) *Posts {
	return &Posts{
		ps:     p,
		logger: l,
		tracer: t,
	}
}

func (p Posts) Create(ctx context.Context, post models.Post) (*models.Post, error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, p.tracer,
		"PostRepo.Create")
	defer span.Finish()
	span.LogFields(
		log.String("Post request", post.String()),
	)
	id, err := p.ps.Create(ctx, post)
	if err != nil {
		p.logger.Error(fmt.Sprintf(`cannot read post: %s`, err))
		span.LogFields(log.Error(err))
		return nil, fmt.Errorf("cannot create post: %w", err)
	}
	post.Id = id
	span.LogFields(
		log.String("Post result", post.String()),
	)
	return &post, nil
}

func (p Posts) Read(ctx context.Context, id int) (*models.Post, error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, p.tracer,
		"PostRepo.Read")
	defer span.Finish()
	span.LogFields(
		log.String("id", strconv.Itoa(id)),
	)
	post, err := p.ps.Read(ctx, id)
	if err != nil {
		p.logger.Error(fmt.Sprintf(`cannot read post: %s`, err))
		span.LogFields(log.Error(err))
		return nil, fmt.Errorf("cannot read post: %w", err)
	}
	span.LogFields(
		log.String("Post result", post.String()),
	)
	return post, nil
}

func (p Posts) Update(ctx context.Context, updatePost models.Post) (*models.Post, error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, p.tracer,
		"PostRepo.Update")
	defer span.Finish()
	span.LogFields(
		log.String("id", strconv.Itoa(updatePost.Id)),
		log.String("updatePost", updatePost.String()),
	)
	post, err := p.ps.Update(ctx, updatePost)
	if err != nil {
		p.logger.Error(fmt.Sprintf(`cannot update post: %s`, err))
		span.LogFields(log.Error(err))
		return nil, fmt.Errorf("cannot update post: %w", err)
	}
	return post, nil
}

func (p Posts) Delete(ctx context.Context, id int) (*models.Post, error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, p.tracer,
		"PostRepo.Delete")
	defer span.Finish()
	span.LogFields(
		log.String("id", strconv.Itoa(id)),
	)
	post, err := p.ps.Read(ctx, id)
	if err != nil {
		p.logger.Error(fmt.Sprintf(`cannot read post: %s`, err))
		span.LogFields(log.Error(err))
		return nil, fmt.Errorf("cannot read post: %w", err)
	}
	span.LogFields(
		log.String("Post delete", post.String()),
	)
	return post, p.ps.Delete(ctx, id)
}
