package main

import (
	"fmt"
	"log"
	"net/http"

	"ginblog/config"
	"ginblog/controllers"
	"ginblog/models"

	"github.com/Sirupsen/logrus"
	"github.com/claudiu/gocron"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"
)

func main() {
	initLogger() // 初始化日志记录器

	config.LoadConfig() // 加载配置文件
	//conf:=config.LoadConf() //viper格式

	//models.SetDB(config.GetConnectionString())
	log.Print(config.GetConnectionString()) // 打印数据库连接字符串

	models.SetDB(config.GetConnectionString()) // 设置数据库连接
	models.AutoMigrate()                       // 自动迁移数据库
	controllers.LoadTemplates()                // 加载模板

	//Periodic tasks
	gocron.Every(1).Day().Do(controllers.CreateXMLSitemap) // 定时任务，每天执行一次生成XML站点地图
	gocron.Start()                                         // 启动定时任务

	// Creates a gin router with default middleware:
	// logger and recovery (crash-free) middleware
	router := gin.Default()                            // 创建一个gin路由器，使用默认的中间件：日志记录和恢复中间件
	router.SetHTMLTemplate(controllers.GetTemplates()) // 设置HTML模板

	//setup sessions
	store := memstore.NewStore([]byte(config.GetConfig().SessionSecret))          // 创建一个内存存储的session存储器
	store.Options(sessions.Options{Path: "/", HttpOnly: true, MaxAge: 7 * 86400}) // 设置session选项
	router.Use(sessions.Sessions("ginblog-session", store))                       // 使用session中间件

	//setup csrf protection
	router.Use(csrf.Middleware(csrf.Options{
		Secret: config.GetConfig().SessionSecret,
		ErrorFunc: func(c *gin.Context) {
			logrus.Error("CSRF token mismatch")                                  // CSRF令牌不匹配
			controllers.ShowErrorPage(c, 400, fmt.Errorf("CSRF token mismatch")) // 显示错误页面
			c.Abort()
		},
	}))

	router.StaticFS("/public", http.Dir(config.PublicPath())) // 静态文件服务
	router.Use(controllers.ContextData())                     // 使用上下文数据中间件

	router.GET("/", controllers.HomeGet)          // 首页路由
	router.NoRoute(controllers.NotFound)          // 未找到路由
	router.NoMethod(controllers.MethodNotAllowed) // 不允许的请求方法

	if config.GetConfig().SignupEnabled { // 如果允许注册
		router.GET("/signup", controllers.SignUpGet)   // 注册页面路由
		router.POST("/signup", controllers.SignUpPost) // 注册请求路由
	}
	router.GET("/signin", controllers.SignInGet)   // 登录页面路由
	router.POST("/signin", controllers.SignInPost) // 登录请求路由
	router.GET("/logout", controllers.LogoutGet)   // 登出路由

	router.GET("/oauthgooglelogin", controllers.OauthGoogleLogin) // Google OAuth登录路由
	router.GET("/oauthcallback", controllers.OauthCallback)       // OAuth回调路由
	router.POST("/new_comment", controllers.CommentPublicCreate)  // 新建评论路由

	router.GET("/pages/:id", controllers.PageGet)                // 页面路由
	router.GET("/posts/:id", controllers.PostGet)                // 文章路由
	router.GET("/tags/:slug", controllers.TagGet)                // 标签路由
	router.GET("/archives/:year/:month", controllers.ArchiveGet) // 归档路由
	router.GET("/rss", controllers.RssGet)                       // RSS路由

	authorized := router.Group("/admin")       // 管理员路由组
	authorized.Use(controllers.AuthRequired()) // 使用身份验证中间件
	{
		authorized.GET("/", controllers.AdminGet) // 管理员首页路由

		authorized.POST("/upload", controllers.UploadPost) // 图片上传路由

		authorized.GET("/users", controllers.UserIndex)              // 用户列表路由
		authorized.GET("/new_user", controllers.UserNew)             // 新建用户路由
		authorized.POST("/new_user", controllers.UserCreate)         // 新建用户请求路由
		authorized.GET("/users/:id/edit", controllers.UserEdit)      // 编辑用户路由
		authorized.POST("/users/:id/edit", controllers.UserUpdate)   // 编辑用户请求路由
		authorized.POST("/users/:id/delete", controllers.UserDelete) // 删除用户请求路由

		authorized.GET("/pages", controllers.PageIndex)              // 页面列表路由
		authorized.GET("/new_page", controllers.PageNew)             // 新建页面路由
		authorized.POST("/new_page", controllers.PageCreate)         // 新建页面请求路由
		authorized.GET("/pages/:id/edit", controllers.PageEdit)      // 编辑页面路由
		authorized.POST("/pages/:id/edit", controllers.PageUpdate)   // 编辑页面请求路由
		authorized.POST("/pages/:id/delete", controllers.PageDelete) // 删除页面请求路由

		authorized.GET("/posts", controllers.PostIndex)              // 文章列表路由
		authorized.GET("/new_post", controllers.PostNew)             // 新建文章路由
		authorized.POST("/new_post", controllers.PostCreate)         // 新建文章请求路由
		authorized.GET("/posts/:id/edit", controllers.PostEdit)      // 编辑文章路由
		authorized.POST("/posts/:id/edit", controllers.PostUpdate)   // 编辑文章请求路由
		authorized.POST("/posts/:id/delete", controllers.PostDelete) // 删除文章请求路由

		authorized.GET("/comments", controllers.CommentIndex)              // 评论列表路由
		authorized.GET("/new_comment", controllers.CommentNew)             // 新建评论路由
		authorized.POST("/new_comment", controllers.CommentCreate)         // 新建评论请求路由
		authorized.GET("/comments/:id/edit", controllers.CommentEdit)      // 编辑评论路由
		authorized.POST("/comments/:id/edit", controllers.CommentUpdate)   // 编辑评论请求路由
		authorized.POST("/comments/:id/delete", controllers.CommentDelete) // 删除评论请求路由

		authorized.GET("/tags", controllers.TagIndex)                 // 标签列表路由
		authorized.GET("/new_tag", controllers.TagNew)                // 新建标签路由
		authorized.POST("/new_tag", controllers.TagCreate)            // 新建标签请求路由
		authorized.POST("/tags/:title/delete", controllers.TagDelete) // 删除标签请求路由
	}

	// Listen and server on 0.0.0.0:8080
	host := config.GetConfig().Host // 获取主机名
	port := config.GetConfig().Port // 获取端口号
	router.Run(host + ":" + port)   // 运行路由器
}

// initLogger initializes logrus logger with some defaults
func initLogger() {
	logrus.SetFormatter(&logrus.TextFormatter{}) // 设置日志格式
	//logrus.SetOutput(os.Stderr)
	if gin.Mode() == gin.DebugMode { // 如果是调试模式
		logrus.SetLevel(logrus.DebugLevel) // 设置日志级别为Debug
	}
}
