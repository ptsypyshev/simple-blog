package blog

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/opentracing/opentracing-go"
	"github.com/ptsypyshev/simple-blog/internal/blog/handlers"
	"github.com/ptsypyshev/simple-blog/internal/db/commentstore"
	"github.com/ptsypyshev/simple-blog/internal/db/pgdb"
	"github.com/ptsypyshev/simple-blog/internal/db/poststore"
	"github.com/ptsypyshev/simple-blog/internal/db/userstore"
	"github.com/ptsypyshev/simple-blog/internal/repositories/commentrepo"
	"github.com/ptsypyshev/simple-blog/internal/repositories/postrepo"
	"github.com/ptsypyshev/simple-blog/internal/repositories/userrepo"
	"log"

	//nice "github.com/ekyoung/gin-nice-recovery"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
)

type App struct {
	db       *pgxpool.Pool
	users    userrepo.Users
	posts    postrepo.Posts
	comments commentrepo.Comments
	logger   *zap.Logger
	tracer   opentracing.Tracer
}

func (a *App) Init() (io.Closer, error) {
	ctx := context.Background()
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	defer func() { _ = logger.Sync() }()
	tracer, closer := InitJaeger("App", "localhost:6831", logger)

	db, err := pgdb.InitDB(ctx, logger, tracer)
	if err != nil {
		log.Fatalf("cannot init DB: %s", err)
	}

	ustore := userstore.NewUsersDB(db, logger, tracer)
	pstore := poststore.NewPostsDB(db, logger, tracer)
	cstore := commentstore.NewCommentsDB(db, logger, tracer)

	a.logger = logger
	a.tracer = tracer
	a.db = db
	a.users = *userrepo.NewUsers(ustore, logger, tracer)
	a.posts = *postrepo.NewPosts(pstore, logger, tracer)
	a.comments = *commentrepo.NewComments(cstore, logger, tracer)

	return closer, nil
}

func (a *App) Serve() error {
	////Initialize Handlers
	userHandlers := blog.NewUserHandlers(a.users, a.logger, a.tracer)
	postHandlers := blog.NewPostHandlers(a.posts, a.logger, a.tracer)
	commentHandlers := blog.NewCommentHandlers(a.comments, a.logger, a.tracer)
	defaultHandlers := blog.NewDefaultHandlers(a.db, a.logger, a.tracer)
	//panicHandler := handler.NewPanicHandler(a.logger, a.tracer)

	//Initialize Router and add Middleware
	router := gin.Default()
	router.Static("/assets", "./assets")
	//router.Use(nice.Recovery(panicHandler.RecoveryHandler))
	router.LoadHTMLGlob("assets/**/*")

	//Routes

	router.GET("/", defaultHandlers.Index)
	router.GET("/dbinit/", defaultHandlers.InitSchema)
	router.GET("/demodb/", defaultHandlers.AddDemoData)

	router.GET("/users/:id", userHandlers.GetUser)
	router.POST("/users/", userHandlers.CreateUser)
	router.PUT("/users/", userHandlers.UpdateUser)
	router.DELETE("/users/:id", userHandlers.DeleteUser)

	router.GET("/posts/:id", postHandlers.GetPost)
	router.POST("/posts/", postHandlers.CreatePost)
	router.PUT("/posts/", postHandlers.UpdatePost)
	router.DELETE("/posts/:id", postHandlers.DeletePost)

	router.GET("/comments/:id", commentHandlers.GetComment)
	router.POST("/comments/", commentHandlers.CreateComment)
	router.PUT("/comments/", commentHandlers.UpdateComment)
	router.DELETE("/comments/:id", commentHandlers.DeleteComment)

	// Start serving the application
	return router.Run()
}
