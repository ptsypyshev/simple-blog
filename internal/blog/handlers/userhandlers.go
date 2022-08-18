package blog

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/ptsypyshev/simple-blog/internal/models"
	"github.com/ptsypyshev/simple-blog/internal/repositories/userrepo"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
	"strconv"
)

type userHandlers struct {
	userrepo userrepo.Users
	logger   *zap.Logger
	tracer   opentracing.Tracer
}

func NewUserHandlers(us userrepo.Users, l *zap.Logger, t opentracing.Tracer) userHandlers {
	return userHandlers{
		userrepo: us,
		logger:   l,
		tracer:   t,
	}
}

func (h userHandlers) Index(c *gin.Context) {
	span, _ := opentracing.StartSpanFromContextWithTracer(c, h.tracer,
		"userHandlers.Index")
	defer span.Finish()
	h.logger.Info("userHandlers.Index", zap.Field{Key: "method", String: c.Request.Method, Type: zapcore.StringType})
	span.SetTag("method", c.Request.Method)
	span.SetTag("params", c.Params)

	c.String(http.StatusOK, "It works!")
}

func (h userHandlers) CreateUser(c *gin.Context) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(c, h.tracer,
		"userHandlers.CreateUser")
	defer span.Finish()
	h.logger.Info("userHandlers.CreateUser", zap.Field{Key: "method", String: c.Request.Method, Type: zapcore.StringType})
	span.SetTag("method", c.Request.Method)
	span.SetTag("params", c.Params)
	var user models.User
	if err := c.BindJSON(&user); err != nil {
		h.logger.Error(fmt.Sprintf(`bad json: %s`, err))
		span.LogFields(
			log.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	span.LogFields(
		log.String("User request", user.String()),
	)
	newUser, err := h.userrepo.Create(ctx, user)
	if err != nil {
		msg := fmt.Sprintf(`create user error: %s`, err)
		h.logger.Error(msg)
		span.LogFields(log.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}
	span.LogFields(
		log.String("User result", newUser.String()),
	)
	c.JSON(http.StatusOK, newUser)
}

func (h userHandlers) GetUser(c *gin.Context) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(c, h.tracer,
		"userHandlers.GetUser")
	defer span.Finish()
	h.logger.Info("userHandlers.GetUser", zap.Field{Key: "method", String: c.Request.Method, Type: zapcore.StringType})
	span.SetTag("method", c.Request.Method)
	span.SetTag("params", c.Params)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.logger.Warn(fmt.Sprintf(`bad param: %s`, err))
		span.LogFields(log.Error(err))
		c.String(http.StatusBadRequest, fmt.Sprintf(`bad id: %s`, c.Param("id")))
		return
	}
	user, err := h.userrepo.Read(ctx, id)
	if err != nil {
		msg := fmt.Sprintf(`get error: %s`, err)
		h.logger.Warn(msg)
		span.LogFields(log.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}
	span.LogFields(
		log.String("Successfully get user ", fmt.Sprintf("%v", user)),
	)
	c.JSON(http.StatusOK, user)
}

func (h userHandlers) UpdateUser(c *gin.Context) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(c, h.tracer,
		"userHandlers.UpdateUser")
	defer span.Finish()
	h.logger.Info("userHandlers.UpdateUser", zap.Field{Key: "method", String: c.Request.Method, Type: zapcore.StringType})
	span.SetTag("method", c.Request.Method)
	span.SetTag("params", c.Params)
	var user models.User
	if err := c.BindJSON(&user); err != nil {
		h.logger.Error(fmt.Sprintf(`bad json: %s`, err))
		span.LogFields(
			log.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	span.LogFields(
		log.String("User request", user.String()),
	)
	updatedUser, err := h.userrepo.Update(ctx, user)
	if err != nil {
		msg := fmt.Sprintf(`update user error: %s`, err)
		h.logger.Error(msg)
		span.LogFields(log.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}
	span.LogFields(
		log.String("User result", updatedUser.String()),
	)
	c.JSON(http.StatusOK, updatedUser)
}

func (h userHandlers) DeleteUser(c *gin.Context) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(c, h.tracer,
		"userHandlers.DeleteUser")
	defer span.Finish()
	h.logger.Info("userHandlers.DeleteUser", zap.Field{Key: "method", String: c.Request.Method, Type: zapcore.StringType})
	span.SetTag("method", c.Request.Method)
	span.SetTag("params", c.Params)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.logger.Warn(fmt.Sprintf(`bad param: %s`, err))
		span.LogFields(log.Error(err))
		c.String(http.StatusBadRequest, fmt.Sprintf(`bad id: %s`, c.Param("id")))
		return
	}
	deletedUser, err := h.userrepo.Delete(ctx, id)
	if err != nil {
		msg := fmt.Sprintf(`delete user error: %s`, err)
		h.logger.Error(msg)
		span.LogFields(log.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}
	span.LogFields(
		log.String("User result", deletedUser.String()),
	)
	c.JSON(http.StatusOK, deletedUser)
}
