package main

import (
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/liteoj/liteoj/backend/internal/cache"
	"github.com/liteoj/liteoj/backend/internal/config"
	"github.com/liteoj/liteoj/backend/internal/db"
	"github.com/liteoj/liteoj/backend/internal/events"
	"github.com/liteoj/liteoj/backend/internal/handlers"
	"github.com/liteoj/liteoj/backend/internal/middleware"
	"github.com/liteoj/liteoj/backend/internal/seed"
	"github.com/liteoj/liteoj/backend/internal/services/ai"
	"github.com/liteoj/liteoj/backend/internal/services/judge"
	"github.com/liteoj/liteoj/backend/internal/web"
)

func main() {
	startedAt := time.Now()
	cfg := config.Load()

	gdb, err := db.Open(cfg)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	if err := db.Migrate(gdb); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	if err := seed.EnsureAdmin(gdb, cfg); err != nil {
		log.Fatalf("seed admin: %v", err)
	}
	if err := seed.EnsureTestData(gdb); err != nil {
		log.Fatalf("seed testdata: %v", err)
	}

	if cfg.AppMode == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: false,
	}))

	appCache := cache.New()
	broker := events.NewBroker()
	judgeClient := judge.NewClient(cfg.JudgeBaseURL)
	runner := judge.NewRunner(judgeClient)
	queue := judge.NewQueue(gdb, runner, broker, cfg.JudgeQueueWorkers, cfg.JudgeQueueCap,
		time.Duration(cfg.JudgeMaxWaitSeconds)*time.Second)

	aiClient := ai.NewFromConfig(cfg)
	aiPrompts := ai.NewPrompts(cfg, aiClient)
	aiQueue := ai.NewQueue(gdb, broker)
	aiRunner := ai.NewRunner(gdb, aiQueue, aiPrompts, 2, 32,
		time.Duration(cfg.AIMaxWaitSeconds)*time.Second)

	authH := &handlers.AuthHandler{DB: gdb, C: cfg}
	probH := &handlers.ProblemHandler{DB: gdb, C: cfg, Cache: appCache}
	setH := &handlers.ProblemSetHandler{DB: gdb, C: cfg, Broker: broker}
	subH := &handlers.SubmissionHandler{DB: gdb, C: cfg, Queue: queue, Broker: broker}
	adminH := &handlers.AdminHandler{DB: gdb, C: cfg, Cache: appCache, Broker: broker}
	tagH := &handlers.TagHandler{DB: gdb, Cache: appCache}
	aiH := &handlers.AIHandler{DB: gdb, Queue: aiQueue, Runner: aiRunner}
	statsH := &handlers.StatsHandler{DB: gdb}
	rankH := &handlers.RankingHandler{DB: gdb}
	adminStatsH := &handlers.AdminStatsHandler{DB: gdb}
	adminStatusH := &handlers.AdminStatusHandler{
		StartedAt: startedAt, JudgeBase: cfg.JudgeBaseURL,
		JudgeClient: judgeClient, Queue: queue,
	}
	eventsH := &handlers.EventsHandler{Broker: broker}
	homeH := &handlers.HomeHandler{DB: gdb, Broker: broker}

	api := r.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
		api.GET("/meta", func(c *gin.Context) {
			c.JSON(200, gin.H{"languages": cfg.JudgeLangs})
		})
		// Server-sent events stream. Unauthenticated by design — see
		// handlers.EventsHandler for the rationale.
		api.GET("/events/stream", eventsH.Stream)
		// Public home page markdown — 未登录首页渲染源。
		api.GET("/home", homeH.Get)
		// Public browse: anonymous visitors can see the problem catalog and tag
		// dictionary so the landing page (/problems) renders without login.
		// Anything that reveals user-specific state (my_status, submissions,
		// problemsets) stays behind Auth below.
		// 用 OptionalAuth 是为了让"已登录用户浏览 /problems"时 my_status 能
		// 正确填充（没有 OptionalAuth 时 token 被丢弃，CurrentUserID=0 → 状态全空）。
		api.GET("/problems", middleware.OptionalAuth(cfg, gdb), probH.List)
		api.GET("/tags", tagH.List)
		api.POST("/auth/login", authH.Login)

		authed := api.Group("")
		authed.Use(middleware.Auth(cfg, gdb))
		{
			authed.GET("/me", authH.Me)
			authed.POST("/me/password", authH.ChangePassword)
			authed.GET("/me/stats", statsH.Stats)
			authed.GET("/me/contribution", statsH.Contribution)

			authed.GET("/problems/:id", probH.Detail)
			authed.POST("/problems/:id/submit", subH.Submit)

			authed.GET("/problemsets", setH.List)
			authed.GET("/problemsets/:id", setH.Detail)
			authed.GET("/problemsets/:id/ranking", rankH.Problemset)
			authed.POST("/problemsets/:id/join", setH.Join)

			authed.GET("/submissions", subH.List)
			authed.GET("/submissions/:id", subH.Detail)
			authed.GET("/submissions/:id/diff/:other", subH.Diff)
			authed.POST("/submissions/:id/analyze", aiH.Analyze)
			authed.POST("/submissions/:id/optimize", aiH.Optimize)

			authed.GET("/ranking", rankH.Global)

			admin := authed.Group("/admin")
			admin.Use(middleware.AdminOnly())
			{
				admin.GET("/stats", adminStatsH.Overview)
				admin.GET("/online", adminStatsH.OnlineUsers)
				admin.GET("/system/status", adminStatusH.Status)
				admin.GET("/ai/tasks", aiH.ListTasks)
				admin.GET("/ai/tasks/:id", aiH.GetTask)

				admin.GET("/users", adminH.ListUsers)
				admin.POST("/users", adminH.CreateUser)
				admin.PUT("/users/:id", adminH.UpdateUser)
				admin.DELETE("/users/:id", adminH.DeleteUser)
				admin.POST("/users/bulk", adminH.BulkCreateUsers)
				admin.GET("/users/:id/profile", adminH.UserProfile)
				admin.PUT("/home", adminH.UpdateHome)

				admin.POST("/problems", adminH.CreateProblem)
				admin.PUT("/problems/:id", adminH.UpdateProblem)
				admin.DELETE("/problems/:id", adminH.DeleteProblem)
				admin.POST("/problems/:id/ai-tag", aiH.AITag)
				admin.POST("/problems/:id/ai-gen-title", aiH.AIGenTitle)
				admin.POST("/problems/:id/ai-gen-desc", aiH.AIGenDesc)
				admin.POST("/problems/:id/ai-gen-idea", aiH.AIGenIdea)
				admin.POST("/problems/:id/ai-gen-explain", aiH.AIGenExplain)
				admin.POST("/problems/:id/ai-gen-all", aiH.AIGenAll)
				admin.GET("/problems/:id/ai-running", aiH.RunningForProblem)

				admin.GET("/problems/:id/testcases", adminH.ListTestcases)
				admin.POST("/problems/:id/testcases", adminH.CreateTestcase)
				admin.PUT("/problems/:id/testcases/:tcid", adminH.UpdateTestcase)
				admin.DELETE("/problems/:id/testcases/:tcid", adminH.DeleteTestcase)

				admin.POST("/problemsets", adminH.CreateProblemSet)
				admin.PUT("/problemsets/:id", adminH.UpdateProblemSet)
				admin.POST("/problemsets/:id/visibility", adminH.ToggleProblemSetVisibility)
				admin.DELETE("/problemsets/:id", adminH.DeleteProblemSet)
				admin.POST("/problemsets/:id/copy", adminH.CopyProblemSet)
				admin.PUT("/problemsets/:id/problems", adminH.SetProblemSetItems)
				admin.GET("/problemsets/:id/members", adminH.ListProblemSetMembers)
				admin.DELETE("/problemsets/:id/members/:uid", adminH.RemoveProblemSetMember)
				admin.GET("/problemsets/:id/bans", adminH.ListProblemSetBans)
				admin.DELETE("/problemsets/:id/bans/:uid", adminH.UnbanProblemSetMember)

				admin.POST("/taggroups", tagH.CreateGroup)
				admin.PUT("/taggroups/:id", tagH.UpdateGroup)
				admin.DELETE("/taggroups/:id", tagH.DeleteGroup)
				admin.POST("/tags", tagH.CreateTag)
				admin.PUT("/tags/:id", tagH.UpdateTag)
				admin.DELETE("/tags/:id", tagH.DeleteTag)
			}
		}
	}

	web.Register(r)

	addr := ":" + cfg.AppPort
	log.Printf("LiteOJ listening on %s (mode=%s, db=%s)", addr, cfg.AppMode, cfg.DBDriver)
	if err := r.Run(addr); err != nil {
		log.Fatalf("listen: %v", err)
	}
}
