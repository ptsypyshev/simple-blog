package pgdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/log/zapadapter"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"strconv"
	"strings"
)

const (
	DatabaseURL = "postgres://usr:pwd@localhost:5432/simpleblog?sslmode=disable"
	InitDBQuery = `
-- Drop All Tables and Extensions
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS posts CASCADE;
DROP TABLE IF EXISTS comments;
DROP EXTENSION pgcrypto;

-- Create some required DB settings (Only first time)
-- Set timezone to Yekaterinburg (GMT+05)
set timezone = 'Asia/Yekaterinburg';
-- Create extension to use cryptography functions in queries
CREATE EXTENSION pgcrypto;

-- Create New Tables
CREATE TABLE IF NOT EXISTS users
(
	id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
	username VARCHAR(100) NOT NULL UNIQUE,
	password VARCHAR(100) NOT NULL,
	first_name VARCHAR(100),
	last_name VARCHAR(100),
	email VARCHAR(100),
	is_active BOOL
);
CREATE TABLE IF NOT EXISTS posts
(
	id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
	title VARCHAR(255) NOT NULL UNIQUE,
	body TEXT NOT NULL,
	user_id INT,
	FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE SET NULL ON UPDATE CASCADE
);
CREATE TABLE IF NOT EXISTS comments
(
	id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
	date TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(),
	body TEXT NOT NULL,
	user_id INT,
	post_id INT,
	FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE SET NULL ON UPDATE CASCADE,
	FOREIGN KEY (post_id) REFERENCES posts (id) ON DELETE CASCADE ON UPDATE CASCADE
);
`
	InitDemoQuery = `
-- Insert Users
INSERT INTO users(username, password, first_name, last_name, email, is_active)
VALUES
	('admin', crypt('password', gen_salt('bf', 8)), 'Administrator', 'TaskSystem', 'admin@example.loc', 'true'),
	('ptsypyshev', crypt('testpass', gen_salt('bf', 8)), 'Pavel', 'Tsypyshev', 'ptsypyshev@example.loc', 'true'),
	('vpupkin', crypt('puptest', gen_salt('bf', 8)), 'Vasiliy', 'Pupkin', 'vpupkin@example.loc', 'false'),
	('iivanov', crypt('ivantest', gen_salt('bf', 8)), 'Ivan', 'Ivanov', 'iivanov@example.loc', 'true'),
	('ppetrov', crypt('petrtest', gen_salt('bf', 8)), 'Petr', 'Petrov', 'ppetrov@example.loc', 'true'),
	('ssidorov', crypt('sidrtest', gen_salt('bf', 8)), 'Sidor', 'Sidorov', 'ssidorov@example.loc', 'true');

-- Insert Posts
INSERT INTO posts(title, body, user_id)
VALUES
	('Post 1', 'Content for post 1', 2),
	('Post 2', 'Content for post 2', 3),
	('Post 3', 'Content for post 3', 4),
	('Post 4', 'Content for post 4', 5),
	('Post 5', 'Content for post 5', 6),
	('Post 6', 'Content for post 6', 6),
	('Post 7', 'Content for post 7', 5),
	('Post 8', 'Content for post 8', 4),
	('Post 9', 'Content for post 9', 3),
	('Post 10', 'Content for post 10', 2);

-- Insert Comments
INSERT INTO comments(body, user_id, post_id)
VALUES
	('Comment 1', 6, 1),
	('Comment 2', 5, 2),
	('Comment 3', 4, 3),
	('Comment 4', 3, 4),
	('Comment 5', 2, 5),
	('Comment 6', 2, 1),
	('Comment 7', 3, 2),
	('Comment 8', 4, 8),
	('Comment 9', 5, 9),
	('Comment 10',6, 1);
`
)

var (
	ErrNotFound      = errors.New("not found")
	ErrMultipleFound = errors.New("multiple found")
)

func InitDB(ctx context.Context, logger *zap.Logger, tracer opentracing.Tracer) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse conn string (%s): %w", DatabaseURL, err)
	}
	config.ConnConfig.LogLevel = pgx.LogLevelDebug
	config.ConnConfig.Logger = zapadapter.NewLogger(logger) // логгер запросов в БД
	pool, err := pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}
	return pool, nil
}

func InitSchema(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, InitDBQuery)
	return err
}

func AddDemoData(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, InitDemoQuery)
	return err
}

func UpdateQueryCompilation(dbTable string, obj interface{}, defaultObj interface{}) (string, error) {
	objMap, err := structToMap(obj)
	if err != nil {
		return "", fmt.Errorf("convert error: %w", err)
	}
	id, ok := objMap["id"]
	if !ok {
		return "", fmt.Errorf("no id specified: %w", err)
	}
	defaultObjMap, err := structToMap(defaultObj)
	if err != nil {
		return "", fmt.Errorf("convert error: %w", err)
	}

	fields := make([]string, 0, len(objMap))
	values := make([]string, 0, len(objMap))
	for k, v := range objMap {
		if k == "id" {
			continue
		}
		if v != defaultObjMap[k] {
			fields = append(fields, k)
			var vStr string
			switch v.(type) {
			case bool:
				vStr = strconv.FormatBool(v.(bool))
			case float64:
				vStr = strconv.FormatFloat(v.(float64), 'f', 0, 64)
			case string:
				vStr = v.(string)
			}
			values = append(values, fmt.Sprintf("'%s'", vStr))
		}
	}
	var fmtStr string
	if len(values) < 2 {
		fmtStr = "UPDATE %s SET %s = (%s) WHERE id = %.0f;"
	} else {
		fmtStr = "UPDATE %s SET (%s) = (%s) WHERE id = %.0f;"
	}

	query := fmt.Sprintf(
		fmtStr,
		dbTable,
		strings.Join(fields, ","),
		strings.Join(values, ","),
		id,
	)

	return query, nil
}

func structToMap(s interface{}) (m map[string]interface{}, err error) {
	j, err := json.Marshal(s)
	if err != nil {
		return nil, fmt.Errorf("cannot parse json: %w", err)
	}
	err = json.Unmarshal(j, &m)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal to map: %w", err)
	}
	return m, nil
}
