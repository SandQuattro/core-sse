package main

import (
	"flag"
	"fmt"
	"github.com/LogDoc-org/logdoc-go-appender/common"
	logdoc "github.com/LogDoc-org/logdoc-go-appender/logrus"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"runtime"
	"sse-demo-core/internal/app/logging"
	"sse-demo-core/internal/app/utils"
	"sse-demo-core/internal/app/utils/gs"
	"sse-demo-core/internal/config"
	"sse-demo-core/internal/db"
	"sse-demo-core/internal/pkg/app"
	"syscall"
)

func main() {
	// Первым делом считаем аргументы командной строки
	confFile := flag.String("config", "conf/application.conf", "-config=<config file name>")
	port := flag.String("port", "9001", "-port=<service port>")

	flag.Parse()

	// и подгрузим конфиг
	config.MustConfig(confFile)
	conf := config.GetConfig()

	// Создаем подсистему логгирования LogDoc
	conn, err := logging.LDSubsystemInit()
	logger := logdoc.GetLogger()

	c := *conn
	if c == nil || err != nil {
		logger.Error("Error LogDoc subsystem initialization")
	} else {
		defer c.Close()
	}
	logger.Formatter = &logrus.JSONFormatter{
		TimestampFormat: "02-01-2006 15:04:05.00000",
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "@timestamp",
			logrus.FieldKeyLevel: "@level",
			logrus.FieldKeyMsg:   "@message",
			logrus.FieldKeyFunc:  "@caller",
		},
	}
	logger.Info(fmt.Sprintf("LogDoc subsystem initialized successfully@@source=%s", common.SourceNameWithLine(runtime.Caller(0))))

	if port == nil || *port == "" {
		logger.Fatal(">> Error, port is empty")
	}

	_ = utils.CreatePID()
	defer func() {
		err := os.Remove("RUNNING_PID")
		// err := os.Remove("RUNNING_PID_" + strconv.Itoa(pid))
		if err != nil {
			logger.Fatal("Error removing PID file. Exiting...")
		}
	}()

	// Коннектимся к базе
	dbPass := os.Getenv("PGPASS")
	if dbPass == "" {
		logger.Fatal("db password is empty")
	}

	d := db.Connect(conf, dbPass)
	defer func(d *sqlx.DB) {
		err := d.Close()
		if err != nil {
			logger.Fatal(err)
		}
	}(d)
	logger.Info(">> DATABASE CONNECTION SUCCESSFUL")

	// Создадим приложение
	a, err := app.New(conf, *port, d)
	if err != nil {
		logger.Fatal(err)
	}

	// ждем kill -SIGHUP <pid> и перечитываем конфиг
	sighup := make(chan os.Signal, 1)
	signal.Notify(sighup, syscall.SIGHUP)

	go func() {
		for range sighup {
			config.MustConfig(confFile)
			fmt.Println(">> CONFIG RELOADED")
		}
	}()

	go func() {
		// и запустим приложение (веб сервер)
		logger.Debug(fmt.Sprintf(">> RUNNING SERVER ON PORT: %s", *port))
		err = a.Run()
		if err != nil {
			logger.Fatal(err)
		}
	}()
	gs.GraceShutdown(a)

	close(sighup)
}
