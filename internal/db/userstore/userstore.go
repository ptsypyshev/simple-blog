package userstore

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/ptsypyshev/simple-blog/internal/db/pgdb"
	"github.com/ptsypyshev/simple-blog/internal/models"
	"github.com/ptsypyshev/simple-blog/internal/repositories/userrepo"
	"go.uber.org/zap"
	"strconv"
)

const (
	UserCreate = `
INSERT INTO users(username, password, first_name, last_name, email, is_active)
VALUES
    ($1, crypt($2, gen_salt('bf', 8)), $3, $4, $5, $6)
RETURNING id;
`
	UserSelectByID = `SELECT * FROM users WHERE id = $1;`
	UserDeleteByID = `
DELETE FROM users WHERE id = $1;
`
)

var _ userrepo.UserStorage = &UsersDB{}

type UsersDB struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
	tracer opentracing.Tracer
}

func NewUsersDB(p *pgxpool.Pool, l *zap.Logger, t opentracing.Tracer) *UsersDB {
	return &UsersDB{
		pool:   p,
		logger: l,
		tracer: t,
	}
}

func (db *UsersDB) Create(ctx context.Context, user models.User) (int, error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, db.tracer,
		"UserStore.Create")
	defer span.Finish()
	span.LogFields(
		log.String("query", UserCreate),
		log.String("arg0", user.String()),
	)
	var id int
	res := db.pool.QueryRow(
		ctx, UserCreate, user.Username, user.Password, user.FirstName, user.LastName, user.Email, user.IsActive,
	)
	err := res.Scan(&id)
	if err != nil {
		span.LogFields(log.Error(err))
		return 0, err
	}
	span.LogFields(
		log.String("User result", user.String()),
	)
	return id, nil
}

func (db *UsersDB) Read(ctx context.Context, id int) (*models.User, error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, db.tracer,
		"UserStore.Read")
	defer span.Finish()
	span.LogFields(
		log.String("query", UserSelectByID),
		log.String("arg0", strconv.Itoa(id)),
	)
	rows, _ := db.pool.Query(ctx, UserSelectByID, id)
	var (
		user  models.User
		found bool
	)
	for rows.Next() {
		if found {
			err := fmt.Errorf("%w: user id %d", pgdb.ErrMultipleFound, id)
			span.LogFields(log.Error(err))
			return nil, err
		}
		if err := rows.Scan(&user.Id, &user.Username, &user.Password, &user.FirstName, &user.LastName, &user.Email, &user.IsActive); err != nil {
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
		err := fmt.Errorf("%w: user id %d", pgdb.ErrNotFound, id)
		span.LogFields(log.Error(err))
		return nil, err
	}
	span.LogFields(
		log.String("User result", user.String()),
	)
	return &user, nil
}

func (db *UsersDB) Update(ctx context.Context, user models.User) (*models.User, error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, db.tracer,
		"UserStore.Update")
	defer span.Finish()
	UpdateQuery, err := pgdb.UpdateQueryCompilation("users", user, models.User{})
	if err != nil {
		err = fmt.Errorf("cannot compile query: %w", err)
		span.LogFields(log.Error(err))
		return &models.User{}, err
	}
	span.LogFields(
		log.String("query", UpdateQuery),
		log.String("arg0", user.String()),
	)
	res, err := db.pool.Exec(ctx, UpdateQuery)
	if err != nil {
		span.LogFields(log.Error(err))
		return &models.User{}, err
	}

	if rowsAffected := res.RowsAffected(); rowsAffected != 1 {
		err = fmt.Errorf("update user error: %d rows affected", rowsAffected)
		span.LogFields(log.Error(err))
		return &models.User{}, err
	}
	span.LogFields(
		log.String("User result", user.String()),
	)
	return &user, nil
}

func (db *UsersDB) Delete(ctx context.Context, id int) error {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, db.tracer,
		"UserStore.Delete")
	defer span.Finish()
	span.LogFields(
		log.String("query", UserSelectByID),
		log.String("arg0", strconv.Itoa(id)),
	)
	res, err := db.pool.Exec(ctx, UserDeleteByID, id)
	if err != nil {
		span.LogFields(log.Error(err))
		return err
	}
	if rowsAffected := res.RowsAffected(); rowsAffected != 1 {
		err = fmt.Errorf("delete user error: %d rows affected", rowsAffected)
		span.LogFields(log.Error(err))
		return err
	}
	span.LogFields(
		log.String("Deleted user with id", strconv.Itoa(id)),
	)
	return nil
}

//
//func UpdateQueryCompilation(user models.User) (string, error) {
//	defaultUserMap, err := structToMap(models.User{})
//	if err != nil {
//		return "", fmt.Errorf("convert error: %w", err)
//	}
//	userMap, err := structToMap(user)
//	if err != nil {
//		return "", fmt.Errorf("convert error: %w", err)
//	}
//
//	fields := make([]string, 0, len(userMap))
//	values := make([]string, 0, len(userMap))
//	for k, v := range userMap {
//		if k == "id" {
//			continue
//		}
//		if v != defaultUserMap[k] {
//			fields = append(fields, k)
//			var vStr string
//			switch v.(type) {
//			case bool:
//				vStr = strconv.FormatBool(v.(bool))
//			case float64:
//				vStr = strconv.FormatFloat(v.(float64), 'f', 0, 64)
//			case string:
//				vStr = v.(string)
//			}
//			values = append(values, fmt.Sprintf("'%s'", vStr))
//		}
//	}
//
//	query := fmt.Sprintf(
//		"UPDATE users SET (%s) = (%s) WHERE id = %d;",
//		strings.Join(fields, ","),
//		strings.Join(values, ","),
//		user.Id,
//	)
//
//	return query, nil
//}
//
//func structToMap(s models.User) (m map[string]interface{}, err error) {
//	j, err := json.Marshal(s)
//	if err != nil {
//		return nil, fmt.Errorf("cannot parse json: %w", err)
//	}
//	err = json.Unmarshal(j, &m)
//	if err != nil {
//		return nil, fmt.Errorf("cannot unmarshal to map: %w", err)
//	}
//	return m, nil
//}
