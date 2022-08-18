package blog

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/ptsypyshev/simple-blog/internal/models"
	"github.com/ptsypyshev/simple-blog/internal/repositories/postrepo"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
	"strconv"
)

type postHandlers struct {
	postrepo postrepo.Posts
	logger   *zap.Logger
	tracer   opentracing.Tracer
}

func NewPostHandlers(ps postrepo.Posts, l *zap.Logger, t opentracing.Tracer) postHandlers {
	return postHandlers{
		postrepo: ps,
		logger:   l,
		tracer:   t,
	}
}

func (h postHandlers) CreatePost(c *gin.Context) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(c, h.tracer,
		"postHandlers.CreatePost")
	defer span.Finish()
	h.logger.Info("postHandlers.CreatePost", zap.Field{Key: "method", String: c.Request.Method, Type: zapcore.StringType})
	span.SetTag("method", c.Request.Method)
	span.SetTag("params", c.Params)
	var post models.Post
	if err := c.BindJSON(&post); err != nil {
		h.logger.Error(fmt.Sprintf(`bad json: %s`, err))
		span.LogFields(
			log.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	span.LogFields(
		log.String("Post request", post.String()),
	)
	newPost, err := h.postrepo.Create(ctx, post)
	if err != nil {
		msg := fmt.Sprintf(`create post error: %s`, err)
		h.logger.Error(msg)
		span.LogFields(log.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}
	span.LogFields(
		log.String("Post result", newPost.String()),
	)
	c.JSON(http.StatusOK, newPost)
}

func (h postHandlers) GetPost(c *gin.Context) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(c, h.tracer,
		"postHandlers.GetPost")
	defer span.Finish()
	h.logger.Info("postHandlers.GetPost", zap.Field{Key: "method", String: c.Request.Method, Type: zapcore.StringType})
	span.SetTag("method", c.Request.Method)
	span.SetTag("params", c.Params)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.logger.Warn(fmt.Sprintf(`bad param: %s`, err))
		span.LogFields(log.Error(err))
		c.String(http.StatusBadRequest, fmt.Sprintf(`bad id: %s`, c.Param("id")))
		return
	}
	post, err := h.postrepo.Read(ctx, id)
	if err != nil {
		msg := fmt.Sprintf(`get error: %s`, err)
		h.logger.Warn(msg)
		span.LogFields(log.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}
	span.LogFields(
		log.String("Successfully get post ", fmt.Sprintf("%v", post)),
	)
	c.JSON(http.StatusOK, post)
}

func (h postHandlers) UpdatePost(c *gin.Context) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(c, h.tracer,
		"postHandlers.UpdatePost")
	defer span.Finish()
	h.logger.Info("postHandlers.UpdatePost", zap.Field{Key: "method", String: c.Request.Method, Type: zapcore.StringType})
	span.SetTag("method", c.Request.Method)
	span.SetTag("params", c.Params)
	var post models.Post
	if err := c.BindJSON(&post); err != nil {
		h.logger.Error(fmt.Sprintf(`bad json: %s`, err))
		span.LogFields(
			log.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	span.LogFields(
		log.String("Post request", post.String()),
	)
	updatedPost, err := h.postrepo.Update(ctx, post)
	if err != nil {
		msg := fmt.Sprintf(`update post error: %s`, err)
		h.logger.Error(msg)
		span.LogFields(log.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}
	span.LogFields(
		log.String("Post result", updatedPost.String()),
	)
	c.JSON(http.StatusOK, updatedPost)
}

func (h postHandlers) DeletePost(c *gin.Context) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(c, h.tracer,
		"postHandlers.DeletePost")
	defer span.Finish()
	h.logger.Info("postHandlers.DeletePost", zap.Field{Key: "method", String: c.Request.Method, Type: zapcore.StringType})
	span.SetTag("method", c.Request.Method)
	span.SetTag("params", c.Params)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.logger.Warn(fmt.Sprintf(`bad param: %s`, err))
		span.LogFields(log.Error(err))
		c.String(http.StatusBadRequest, fmt.Sprintf(`bad id: %s`, c.Param("id")))
		return
	}
	deletedPost, err := h.postrepo.Delete(ctx, id)
	if err != nil {
		msg := fmt.Sprintf(`delete post error: %s`, err)
		h.logger.Error(msg)
		span.LogFields(log.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}
	span.LogFields(
		log.String("Post result", deletedPost.String()),
	)
	c.JSON(http.StatusOK, deletedPost)
}
