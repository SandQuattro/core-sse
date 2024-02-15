package fileutils

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/LogDoc-org/logdoc-go-appender/common"
	logdoc "github.com/LogDoc-org/logdoc-go-appender/logrus"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"mime/multipart"
	"net/http"
	"runtime"
	"sse-demo-core/internal/app/interfaces/services"
	"sse-demo-core/internal/app/structs"
	"sse-demo-core/internal/app/utils"
	"strconv"
	"sync"
)

type SSEvent struct {
	GUID  string      `json:"guid"`
	UUID  string      `json:"uuid"`
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

func ProcessAuth(j services.JwtService, users services.UserService, ctx echo.Context) (int, bool, *echo.HTTPError) {
	logger := logdoc.GetLogger()

	userID, err := utils.GetUserIDFromClaims(ctx.Get("claims").(jwt.MapClaims), users)
	if err != nil {
		return -1, false, echo.NewHTTPError(http.StatusBadRequest, err)
	}

	token := ctx.Request().Header.Get("Authorization")

	if token != "" {
		claims, valid, err := j.ValidateToken(token)
		if err != nil || !valid {
			logger.Error(fmt.Sprintf("token validation error, %v@@source=%s", err, common.SourceNameWithLine(runtime.Caller(0))))

			return -1, false, echo.NewHTTPError(http.StatusForbidden, "token validation error")
		}

		var sub string

		switch s := claims["sub"].(type) {
		case float64:
			sub = fmt.Sprintf("%d", int64(s))
		case string:
			sub = s
		default:
			sub = "ERROR, unknown sub type"
		}

		// Ищем пользователя в БД по sub, если не нашли, то по id
		u, err := users.FindUserBySub(sub)
		if err == nil {
			logger.Info("User received successfully bu sub, userId:", userID)
			return u.ID, false, nil
		}

		logger.Error(fmt.Errorf("user not found with sub:%s, error:%w", sub, err))
		// если не нашли, то ищем по id
		userID = int(claims["id"].(float64))
		u, err = users.FindUserById(userID)
		if err != nil {
			logger.Error(fmt.Errorf("user not found with sub:%s and id:%d, error:%w", sub, userID, err))
			return -1, false, echo.NewHTTPError(http.StatusBadRequest, "user not found")
		}
		logger.Info("User received successfully by id, userId:", userID)
		return u.ID, false, nil
	}

	// токен пустой, проверяем куку
	sessionID, _ := ctx.Cookie("session_id")
	if sessionID == nil && token == "" {
		logger.Error(fmt.Sprintf("both session_id and Authorization header are empty, @@source=%s", common.SourceNameWithLine(runtime.Caller(0))))

		return -1, true, echo.NewHTTPError(http.StatusBadRequest, "request with no authorization information")
	}
	userID, _ = strconv.Atoi(sessionID.Value)

	return userID, true, nil
}

func SendSSEvent(ctx echo.Context, guid string, uid string, eventName string, fileName string) error {
	logger := logdoc.GetLogger()

	event := SSEvent{
		GUID:  guid,
		UUID:  uid,
		Event: eventName,
		Data:  fileName,
	}
	data, err := json.Marshal(event)
	if err != nil {
		logger.Error("error marshal event:", event, " to json")
		return err
	}
	_, err = ctx.Response().Write([]byte("data: " + string(data) + "\n\n"))
	ctx.Response().Flush()

	if err != nil {
		logger.Error("error pushing server side event of processing file")
		return err
	}
	return nil
}

func SendSSEEvent(mu sync.Locker, ctx echo.Context, uid string, eventName string, content any, error bool) error {
	logger := logdoc.GetLogger()

	event := structs.SSEvent{
		UUID:  uid,
		Event: eventName,
		Error: error,
		Data:  content,
	}
	data, err := json.Marshal(event)
	if err != nil {
		logger.Error("error marshal event:", event, " to json")
		return err
	}
	mu.Lock()
	_, err = ctx.Response().Write([]byte("data: " + string(data) + "\n\n"))
	ctx.Response().Flush()
	mu.Unlock()

	if err != nil {
		logger.Error("error pushing server side event of processing file, ", err)
		return err
	}
	return nil
}

func SendSSEToConnectionsChanWithTimeout(ctx context.Context, wg *sync.WaitGroup, guid string, sseConnections map[string]chan structs.Notification, data *structs.Notification, async bool) {
	if async {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sendSSEDataToChannel(ctx, guid, sseConnections, data)
		}()
	} else {
		sendSSEDataToChannel(ctx, guid, sseConnections, data)
	}
}

func sendSSEDataToChannel(ctx context.Context, guid string, sseConnections map[string]chan structs.Notification, data *structs.Notification) {
	logger := logdoc.GetLogger()
	defer func() {
		err := recover()
		if err != nil {
			logger.Error("sendSSEDataToChannel panic, ", err)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			logger.Warn(">> nobody reading sse, SendSSEToConnectionsChanWithTimeout done with timeout, event ", *data, " dropped")
			return
		case sseConnections[guid] <- *data:
			return
		}
	}
}

func DetectFileType(file *multipart.FileHeader) (string, error) {
	fileInfo, _ := file.Open()
	defer fileInfo.Close()

	buffer := make([]byte, 512)
	if _, err := fileInfo.Read(buffer); err != nil {
		return "", err
	}
	fileType := http.DetectContentType(buffer)
	_ = fileType

	return fileType, nil
}
