package files

import (
	"context"
	logdoc "github.com/LogDoc-org/logdoc-go-appender/logrus"
	"github.com/gurkankaymak/hocon"
	"github.com/labstack/echo/v4"
	uuid "github.com/satori/go.uuid"
	"mime/multipart"
	"net/http"
	fileutils "sse-demo-core/internal/app/endpoint/files/utils"
	"sse-demo-core/internal/app/interfaces/services"
	docxxlsxprocessor "sse-demo-core/internal/app/processors/docs"
	pdfprocessor "sse-demo-core/internal/app/processors/pdf"
	csvprocessor "sse-demo-core/internal/app/processors/text"
	"sse-demo-core/internal/app/structs"
	"sync"
	"time"
)

type Endpoint struct {
	config *hocon.Config
	jwt    services.JwtService
}

type Response struct {
	FileLayers []FileLayer
}

type FileLayer struct {
	ID     int
	Name   string
	Status string
	Error  string
}

func New(config *hocon.Config, jwtSvc services.JwtService) *Endpoint {
	return &Endpoint{config: config, jwt: jwtSvc}
}

func (e *Endpoint) FileUploadHandler(sseConnections map[string]chan structs.Notification) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		logger := logdoc.GetLogger()
		logger.Info(">> FileUploadHandler started..")

		logger.Info(">> Cookies: \n")
		for _, cook := range ctx.Cookies() {
			logger.Info(cook.Name, ":", cook.Value, ":", cook.Domain)
		}

		var content string

		// Multipart form
		form, err := ctx.MultipartForm()
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}
		logger.Debug(">> multipart form done..")

		files := form.File["file"]
		if files == nil {
			return echo.NewHTTPError(http.StatusBadRequest, "files not uploaded")
		}
		logger.Info(">> getting files done..")

		var guid string

		guidForm := form.Value["guid"]
		if len(guidForm) == 0 {
			guid = uuid.NewV4().String()
		} else {
			guid = guidForm[0]
		}

		sseConnections[guid] = make(chan structs.Notification)

		logger.Info(">> started uploading with guid:", guid)

		stateCh := make(chan structs.Notification, len(files))

		// Запускаем SSE
		ctx.Response().Header().Set("Content-Type", "text/event-stream")
		ctx.Response().Header().Set("Cache-Control", "no-cache")
		ctx.Response().Header().Set("Connection", "keep-alive")
		ctx.Response().Header().Set("X-Accel-Buffering", "no")

		ctx.Response().WriteHeader(http.StatusOK)
		ctx.Response().Flush()

		logger.Debug(">> SSE started..")

		wg := sync.WaitGroup{}

		done := make(chan struct{})

		c, cancel := context.WithTimeout(context.Background(), time.Duration(e.config.GetInt("upload.timeout"))*time.Second)
		defer cancel()

		// Создаем фоновый процесс, который читает сообщения из канала, заполняемого
		// в теле обработчика процесса загрузки
		// и отправляет sse данные сразу на фронтенд
		go func() {
			for {
				// Убедимся, что мы всегда сначала попытаемся вычитать данные
				// если не получается(нет читателей, /sse не запущен),
				// идем дальше и смотрим, если мы закончили обработку файлов
				// чистим сообщения и выходим
				select {
				case <-done:
					// Отменяем все события, ждем и выходим из горутины
					cancel()
					time.Sleep(100 * time.Millisecond)
					done <- struct{}{}
					return
				case state := <-stateCh:
					_ = fileutils.SendSSEvent(ctx, guid, state.UUID, state.State, state.FileName)
				}
			}
		}()

		_ = fileutils.SendSSEvent(ctx, guid, "", "upload started", "")
		fileutils.SendSSEToConnectionsChanWithTimeout(c, &wg, guid, sseConnections, &structs.Notification{GUID: guid, UUID: "", State: "upload started", FileName: ""}, true)

		// Начали обработку файлов
		for _, file := range files {
			wg.Add(1)
			go func(file *multipart.FileHeader) {
				uid := uuid.NewV4().String()
				logger.Info(">> processing file ", file.Filename, ", uid:", uid, " with guid:", guid)

				defer func() {
					stateCh <- structs.Notification{GUID: guid, UUID: uid, State: "file_completed", FileName: file.Filename}
					fileutils.SendSSEToConnectionsChanWithTimeout(c, &wg, guid, sseConnections, &structs.Notification{GUID: guid, UUID: uid, State: "file_completed", FileName: file.Filename}, true)
					wg.Done()
				}()

				// отправляем событие создания слоя данных пользователя
				stateCh <- structs.Notification{GUID: guid, UUID: uid, State: "file_processing_started", FileName: file.Filename}
				fileutils.SendSSEToConnectionsChanWithTimeout(c, &wg, guid, sseConnections, &structs.Notification{GUID: guid, UUID: uid, State: "file_processing_started", FileName: file.Filename}, true)

				// Определяем тип файла
				fileType, err := fileutils.DetectFileType(file)
				if err != nil {
					logger.Error("error checking file:%s", file.Filename)
					return
				}

				// pre-processing file
				switch {
				case fileType == "application/zip" && (file.Header.Get("Content-Type") == "application/vnd.openxmlformats-officedocument.wordprocessingml.document" ||
					file.Header.Get("Content-Type") == "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"):
					content, err = docxxlsxprocessor.ProcessDocument(file)
				case fileType == "application/pdf":
					content, err = pdfprocessor.ProcessPdfFile(file)
				case file.Header.Get("Content-Type") == "text/csv":
					content, err = csvprocessor.ProcessCSVFile(file)
					if err != nil {
						logger.Error(">> File Processing Error, ", err)
						stateCh <- structs.Notification{GUID: guid, UUID: uid, State: "file_processing_error", FileName: file.Filename}
						fileutils.SendSSEToConnectionsChanWithTimeout(c, &wg, guid, sseConnections, &structs.Notification{GUID: guid, UUID: uid, State: "file_processing_error", FileName: file.Filename}, true)
						return
					}
				case fileType == "text/csv":
					content, err = csvprocessor.ProcessCSVHeader(file, 10)
					if err != nil {
						logger.Error(">> File Processing Error, ", err)
						stateCh <- structs.Notification{GUID: guid, UUID: uid, State: "file_processing_error", FileName: file.Filename}
						fileutils.SendSSEToConnectionsChanWithTimeout(c, &wg, guid, sseConnections, &structs.Notification{GUID: guid, UUID: uid, State: "file_processing_error", FileName: file.Filename}, true)
						return
					}
				default:
					logger.Error(">> File Processing Error, unsupported content")
					stateCh <- structs.Notification{GUID: guid, UUID: uid, State: "file_processing_error", FileName: file.Filename}
					fileutils.SendSSEToConnectionsChanWithTimeout(c, &wg, guid, sseConnections, &structs.Notification{GUID: guid, UUID: uid, State: "file_processing_error", FileName: file.Filename}, true)
					return
				}

				if err != nil {
					logger.Error(">> File Processing Error, ", err)
					stateCh <- structs.Notification{GUID: guid, UUID: uid, State: "file_processing_error", FileName: file.Filename}
					fileutils.SendSSEToConnectionsChanWithTimeout(c, &wg, guid, sseConnections, &structs.Notification{GUID: guid, UUID: uid, State: "file_processing_error", FileName: file.Filename}, true)
					return
				}

				stateCh <- structs.Notification{GUID: guid, UUID: uid, State: "file_processed", FileName: file.Filename}
				fileutils.SendSSEToConnectionsChanWithTimeout(c, &wg, guid, sseConnections, &structs.Notification{GUID: guid, UUID: uid, State: "file_processed", FileName: file.Filename}, true)

				if content == "" {
					logger.Error(">> File Processing Error, empty content")
					stateCh <- structs.Notification{GUID: guid, UUID: uid, State: "file_processing_error", FileName: file.Filename}
					fileutils.SendSSEToConnectionsChanWithTimeout(c, &wg, guid, sseConnections, &structs.Notification{GUID: guid, UUID: uid, State: "file_processing_error", FileName: file.Filename}, true)
					return
				}
			}(file)
		}

		wg.Wait()
		_ = fileutils.SendSSEvent(ctx, guid, "", "completed", "")
		fileutils.SendSSEToConnectionsChanWithTimeout(c, &wg, guid, sseConnections, &structs.Notification{GUID: guid, UUID: "", State: "completed", FileName: ""}, false)

		done <- struct{}{}
		<-done

		return nil
	}
}
