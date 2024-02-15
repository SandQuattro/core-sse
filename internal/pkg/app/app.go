package app

import (
	"fmt"
	logdoc "github.com/LogDoc-org/logdoc-go-appender/logrus"
	"github.com/gurkankaymak/hocon"
	"github.com/jmoiron/sqlx"
	"github.com/kitabisa/teler-waf"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"sse-demo-core/internal/app/endpoint/files/streaming"
	"sse-demo-core/internal/app/endpoint/files/uploadsse"
	"sse-demo-core/internal/app/endpoint/root"
	llama2 "sse-demo-core/internal/app/integration/huggingface"
	customcors "sse-demo-core/internal/app/mv/cors"
	"sse-demo-core/internal/app/mv/multipartchecker"
	"sse-demo-core/internal/app/service/jwtservice"
	"sse-demo-core/internal/app/service/userservice"
	"sse-demo-core/internal/app/structs"
	"sse-demo-core/internal/app/utils"
	echopprof "sse-demo-core/internal/pprof"
	"strings"
)

type App struct {
	port   string
	db     *sqlx.DB
	config *hocon.Config
	Echo   *echo.Echo

	root      *root.Endpoint
	streaming *streaming.Endpoint
	files     *files.Endpoint

	u   *userservice.UserServiceImpl
	jwt *jwtservice.JwtServiceImpl
	l2  *llama2.ServiceImpl
}

var logger *logrus.Logger

func New(config *hocon.Config, port string, db *sqlx.DB) (*App, error) {
	logger = logdoc.GetLogger()

	a := App{port: port, config: config, db: db}

	// services
	a.u = userservice.New(db)
	a.jwt = jwtservice.New(config, db)
	a.l2 = llama2.New(config)

	// used to cache user data, openai thread data
	//cache := caching.NewRedisCache(config.GetString("redis.addr"))

	// controllers
	a.root = root.New()

	a.files = files.New(config, a.jwt)
	a.streaming = streaming.New(config)

	// Echo instance
	a.Echo = echo.New()

	// Global Endpoints Middleware
	// Вызов перед каждым обработчиком
	// В них может быть логгирование,
	// поверка токенов, ролей, прав и многое другое
	// TODO: !!! пофиксить на боевом корсы !!!
	a.Echo.Use(customcors.CORS())

	a.Echo.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, values middleware.RequestLoggerValues) error {
			logger.WithFields(logrus.Fields{
				"Time":   values.StartTime.Format("02-01-2006 15:04:05.00000"),
				"URI":    values.URI,
				"status": values.Status,
			}).Info("request")

			return nil
		},
	}))

	if a.config.GetBoolean("debug") {
		echopprof.Wrap(a.Echo)
	}

	a.Echo.Use(middleware.BodyLimit("10M"))

	a.Echo.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
			logger.Error("Recovery from panic, ", err.Error(), "\n", string(stack))
			return err
		},
	}))

	// Metrics middleware
	a.Echo.Use(echoprometheus.NewMiddleware("sse_demo_core"))
	a.Echo.GET("/core/metrics", echoprometheus.NewHandler())

	// Body dump mv captures the request and response payload and calls the registered handler.
	// Generally used for debugging/logging purpose.
	// Avoid using it if your request/response payload is huge e.g. file upload/download
	a.Echo.Use(middleware.BodyDumpWithConfig(middleware.BodyDumpConfig{
		Skipper: func(c echo.Context) bool {
			return strings.Compare(c.Request().RequestURI, "/upload") == 0 ||
				strings.Contains(c.Request().RequestURI, "/core/metrics") ||
				strings.Contains(c.Request().RequestURI, "/debug/") ||
				strings.Contains(c.Request().RequestURI, "/assistants/file") ||
				strings.Contains(c.Request().RequestURI, "/assistant/file")
		},
		Handler: func(c echo.Context, reqBody, resBody []byte) {
			logger.Debug(fmt.Sprintf(`>> BodyDump middleware
			request ip:%s
			request metod:%s
			request uri:%s
			request paylod (if any):%s
			response body:%s`,
				c.RealIP(),
				c.Request().Method,
				c.Request().RequestURI,
				reqBody,
				resBody))
		},
	}))

	// Teler Intrusion Detection MW
	telerMiddleware := teler.New(teler.Options{
		Whitelists: []string{
			`(Postman)/*`, // Пропускаем postman
		},
	})
	a.Echo.Use(echo.WrapMiddleware(telerMiddleware.Handler))

	// Создаем глобальный пул соединений для передачи данных между handlers
	// ключ - guid - уникальный идентификатор загрузки
	var sseConnections = make(map[string]chan structs.Notification)

	a.Echo.GET("/", a.root.RootHandler)
	a.Echo.POST("/upload", a.files.FileUploadHandler(sseConnections), multipartchecker.MultipartCountChecker(a.jwt))
	a.Echo.GET("/sse", a.streaming.ProcessStreamingDataHandler(sseConnections))
	logger.Info("Application created!")

	return &a, nil
}

func (a *App) Run() error {
	closer := utils.Tracing(a.Echo)
	defer closer.Close()

	// Start server
	err := a.Echo.Start(":" + a.port)
	if err != nil {
		logger.Info(err)
	}
	return nil
}
