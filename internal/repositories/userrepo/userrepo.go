package userrepo

import (
	"context"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/ptsypyshev/simple-blog/internal/models"
	"go.uber.org/zap"
	"strconv"
)

type UserCreate interface {
	Create(ctx context.Context, user models.User) (int, error)
}

type UserRead interface {
	Read(ctx context.Context, id int) (*models.User, error)
}

type UserUpdate interface {
	Update(ctx context.Context, user models.User) (*models.User, error)
}

type UserDelete interface {
	Delete(ctx context.Context, id int) error
}

//type UserSearch interface {
//	Search()
//}

type UserStorage interface {
	UserCreate
	UserRead
	UserUpdate
	UserDelete
	//UserSearch
}

type Users struct {
	us     UserStorage
	logger *zap.Logger
	tracer opentracing.Tracer
}

func NewUsers(u UserStorage, l *zap.Logger, t opentracing.Tracer) *Users {
	return &Users{
		us:     u,
		logger: l,
		tracer: t,
	}
}

func (u Users) Create(ctx context.Context, user models.User) (*models.User, error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, u.tracer,
		"UserRepo.Create")
	defer span.Finish()
	span.LogFields(
		log.String("User request", user.String()),
	)
	id, err := u.us.Create(ctx, user)
	if err != nil {
		u.logger.Error(fmt.Sprintf(`cannot read user: %s`, err))
		span.LogFields(log.Error(err))
		return nil, fmt.Errorf("cannot create user: %w", err)
	}
	user.Id = id
	span.LogFields(
		log.String("User result", user.String()),
	)
	return &user, nil
}

func (u Users) Read(ctx context.Context, id int) (*models.User, error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, u.tracer,
		"UserRepo.Read")
	defer span.Finish()
	span.LogFields(
		log.String("id", strconv.Itoa(id)),
	)
	user, err := u.us.Read(ctx, id)
	if err != nil {
		u.logger.Error(fmt.Sprintf(`cannot read user: %s`, err))
		span.LogFields(log.Error(err))
		return nil, fmt.Errorf("cannot read user: %w", err)
	}
	span.LogFields(
		log.String("User result", user.String()),
	)
	return user, nil
}

func (u Users) Update(ctx context.Context, updateUser models.User) (*models.User, error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, u.tracer,
		"UserRepo.Update")
	defer span.Finish()
	span.LogFields(
		log.String("id", strconv.Itoa(updateUser.Id)),
		log.String("updateUser", updateUser.String()),
	)
	user, err := u.us.Update(ctx, updateUser)
	if err != nil {
		u.logger.Error(fmt.Sprintf(`cannot update user: %s`, err))
		span.LogFields(log.Error(err))
		return nil, fmt.Errorf("cannot update user: %w", err)
	}
	return user, nil
}

func (u Users) Delete(ctx context.Context, id int) (*models.User, error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, u.tracer,
		"UserRepo.Delete")
	defer span.Finish()
	span.LogFields(
		log.String("id", strconv.Itoa(id)),
	)
	user, err := u.us.Read(ctx, id)
	if err != nil {
		u.logger.Error(fmt.Sprintf(`cannot read user: %s`, err))
		span.LogFields(log.Error(err))
		return nil, fmt.Errorf("cannot read user: %w", err)
	}
	span.LogFields(
		log.String("User delete", user.String()),
	)
	return user, u.us.Delete(ctx, id)
}
