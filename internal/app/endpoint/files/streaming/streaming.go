package streaming

import (
	"context"
	"errors"
	logdoc "github.com/LogDoc-org/logdoc-go-appender/logrus"
	"github.com/gurkankaymak/hocon"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
	fileutils "sse-demo-core/internal/app/endpoint/files/utils"
	"sse-demo-core/internal/app/structs"
	"time"
)

type Endpoint struct {
	config *hocon.Config
}

func New(config *hocon.Config) *Endpoint {
	return &Endpoint{config: config}
}

func (e *Endpoint) ProcessStreamingDataHandler(sseConnections map[string]chan structs.Notification) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		logger := logdoc.GetLogger()

		logger.Info(">> ProcessStreamingDataHandler started..")

		guid := ctx.QueryParam("guid")
		stream := sseConnections[guid]

		if guid == "" {
			return echo.NewHTTPError(http.StatusBadRequest, structs.ErrorResponse{Error: "empty guid param"})
		}
		if stream == nil {
			return echo.NewHTTPError(http.StatusBadRequest, structs.ErrorResponse{Error: "empty stream"})
		}

		defer func() {
			close(stream)
			delete(sseConnections, guid)
		}()

		ctx.Response().Header().Set("Content-Type", "text/event-stream")
		ctx.Response().Header().Set("Cache-Control", "no-cache")
		ctx.Response().Header().Set("Connection", "keep-alive")
		ctx.Response().Header().Set("X-Accel-Buffering", "no")

		ctx.Response().WriteHeader(http.StatusOK)
		ctx.Response().Flush()

		c, cancel := context.WithTimeout(context.Background(), time.Duration(e.config.GetInt("upload.timeout"))*time.Second)
		defer cancel()

		for {
			select {
			case <-c.Done():
				return nil
			case msg := <-stream:
				// отправляем событие на основании данных из канала
				err := fileutils.SendSSEvent(ctx, msg.GUID, msg.UUID, msg.State, msg.FileName)
				if errors.Is(err, io.EOF) {
					break
				}

				if msg.State == "completed" {
					return nil
				}

				if err != nil {
					logger.Error("Error reading streaming data, error: ", err)
					return echo.NewHTTPError(http.StatusBadRequest, structs.ErrorResponse{Error: err.Error()})
				}
			default:
				// отправляем событие на основании данных из канала
				// _ = fileutils.SendSSEvent(ctx, "", "", "waiting for data...", "")
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}
