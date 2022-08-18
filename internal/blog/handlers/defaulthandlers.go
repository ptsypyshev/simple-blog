package blog

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/ptsypyshev/simple-blog/internal/db/pgdb"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
)

type defaultHandlers struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
	tracer opentracing.Tracer
}

func NewDefaultHandlers(p *pgxpool.Pool, l *zap.Logger, t opentracing.Tracer) defaultHandlers {
	return defaultHandlers{
		pool:   p,
		logger: l,
		tracer: t,
	}
}

func (h defaultHandlers) Index(c *gin.Context) {
	span, _ := opentracing.StartSpanFromContextWithTracer(c, h.tracer,
		"defaultHandlers.Index")
	defer span.Finish()
	h.logger.Info("defaultHandlers.Index", zap.Field{Key: "method", String: c.Request.Method, Type: zapcore.StringType})
	span.SetTag("method", c.Request.Method)
	span.SetTag("params", c.Params)

	c.HTML(http.StatusOK, "main", gin.H{
		"title":   "Simple Blog API",
		"h1_text": "Simple Blog API",
	})
}

func (h defaultHandlers) InitSchema(c *gin.Context) {
	span, _ := opentracing.StartSpanFromContextWithTracer(c, h.tracer,
		"defaultHandlers.InitSchema")
	defer span.Finish()
	h.logger.Info("defaultHandlers.InitSchema", zap.Field{Key: "method", String: c.Request.Method, Type: zapcore.StringType})
	span.SetTag("method", c.Request.Method)
	span.SetTag("params", c.Params)

	if err := pgdb.InitSchema(c, h.pool); err != nil {
		h.logger.Error(fmt.Sprintf(`cannot init schema: %s`, err))
		span.LogFields(log.Error(err))
		c.String(http.StatusInternalServerError, "DB is not initialized")
	}
	c.String(http.StatusOK, "DB Initialized")
}

func (h defaultHandlers) AddDemoData(c *gin.Context) {
	span, _ := opentracing.StartSpanFromContextWithTracer(c, h.tracer,
		"defaultHandlers.AddDemoData")
	defer span.Finish()
	h.logger.Info("defaultHandlers.AddDemoData", zap.Field{Key: "method", String: c.Request.Method, Type: zapcore.StringType})
	span.SetTag("method", c.Request.Method)
	span.SetTag("params", c.Params)

	if err := pgdb.AddDemoData(c, h.pool); err != nil {
		h.logger.Error(fmt.Sprintf(`cannot add demo data: %s`, err))
		span.LogFields(log.Error(err))
		c.String(http.StatusInternalServerError, "Demo data is not added")
	}
	c.String(http.StatusOK, "Demo data is added")
}
