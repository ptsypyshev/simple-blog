package blog

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/ptsypyshev/simple-blog/internal/models"
	"github.com/ptsypyshev/simple-blog/internal/repositories/commentrepo"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
	"strconv"
)

type commentHandlers struct {
	commentrepo commentrepo.Comments
	logger      *zap.Logger
	tracer      opentracing.Tracer
}

func NewCommentHandlers(c commentrepo.Comments, l *zap.Logger, t opentracing.Tracer) commentHandlers {
	return commentHandlers{
		commentrepo: c,
		logger:      l,
		tracer:      t,
	}
}

func (h commentHandlers) CreateComment(c *gin.Context) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(c, h.tracer,
		"commentHandlers.CreateComment")
	defer span.Finish()
	h.logger.Info("commentHandlers.CreateComment", zap.Field{Key: "method", String: c.Request.Method, Type: zapcore.StringType})
	span.SetTag("method", c.Request.Method)
	span.SetTag("params", c.Params)
	var comment models.Comment
	if err := c.BindJSON(&comment); err != nil {
		h.logger.Error(fmt.Sprintf(`bad json: %s`, err))
		span.LogFields(
			log.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	span.LogFields(
		log.String("Comment request", comment.String()),
	)
	newComment, err := h.commentrepo.Create(ctx, comment)
	if err != nil {
		msg := fmt.Sprintf(`create comment error: %s`, err)
		h.logger.Error(msg)
		span.LogFields(log.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}
	span.LogFields(
		log.String("comment result", newComment.String()),
	)
	c.JSON(http.StatusOK, newComment)
}

func (h commentHandlers) GetComment(c *gin.Context) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(c, h.tracer,
		"commentHandlers.GetComment")
	defer span.Finish()
	h.logger.Info("commentHandlers.GetComment", zap.Field{Key: "method", String: c.Request.Method, Type: zapcore.StringType})
	span.SetTag("method", c.Request.Method)
	span.SetTag("params", c.Params)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.logger.Warn(fmt.Sprintf(`bad param: %s`, err))
		span.LogFields(log.Error(err))
		c.String(http.StatusBadRequest, fmt.Sprintf(`bad id: %s`, c.Param("id")))
		return
	}
	comment, err := h.commentrepo.Read(ctx, id)
	if err != nil {
		msg := fmt.Sprintf(`get error: %s`, err)
		h.logger.Warn(msg)
		span.LogFields(log.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}
	span.LogFields(
		log.String("Successfully get comment ", fmt.Sprintf("%v", comment)),
	)
	c.JSON(http.StatusOK, comment)
}

func (h commentHandlers) UpdateComment(c *gin.Context) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(c, h.tracer,
		"commentHandlers.UpdateComment")
	defer span.Finish()
	h.logger.Info("commentHandlers.UpdateComment", zap.Field{Key: "method", String: c.Request.Method, Type: zapcore.StringType})
	span.SetTag("method", c.Request.Method)
	span.SetTag("params", c.Params)
	var comment models.Comment
	if err := c.BindJSON(&comment); err != nil {
		h.logger.Error(fmt.Sprintf(`bad json: %s`, err))
		span.LogFields(
			log.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	span.LogFields(
		log.String("Comment request", comment.String()),
	)
	updatedComment, err := h.commentrepo.Update(ctx, comment)
	if err != nil {
		msg := fmt.Sprintf(`update comment error: %s`, err)
		h.logger.Error(msg)
		span.LogFields(log.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}
	span.LogFields(
		log.String("Comment result", updatedComment.String()),
	)
	c.JSON(http.StatusOK, updatedComment)
}

func (h commentHandlers) DeleteComment(c *gin.Context) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(c, h.tracer,
		"commentHandlers.DeleteComment")
	defer span.Finish()
	h.logger.Info("commentHandlers.DeleteComment", zap.Field{Key: "method", String: c.Request.Method, Type: zapcore.StringType})
	span.SetTag("method", c.Request.Method)
	span.SetTag("params", c.Params)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.logger.Warn(fmt.Sprintf(`bad param: %s`, err))
		span.LogFields(log.Error(err))
		c.String(http.StatusBadRequest, fmt.Sprintf(`bad id: %s`, c.Param("id")))
		return
	}
	deletedComment, err := h.commentrepo.Delete(ctx, id)
	if err != nil {
		msg := fmt.Sprintf(`delete comment error: %s`, err)
		h.logger.Error(msg)
		span.LogFields(log.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}
	span.LogFields(
		log.String("Comment result", deletedComment.String()),
	)
	c.JSON(http.StatusOK, deletedComment)
}
