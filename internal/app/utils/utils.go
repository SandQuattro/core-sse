package utils

import (
	"fmt"
	"github.com/LogDoc-org/logdoc-go-appender/common"
	logdoc "github.com/LogDoc-org/logdoc-go-appender/logrus"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/pkoukk/tiktoken-go"
	"net/http"
	"runtime"
	"sse-demo-core/internal/app/interfaces/services"
	"sse-demo-core/internal/app/structs"
	"strconv"
	"strings"
)

func Ternary(cond bool, a any, b any) any {
	if cond {
		return a
	}
	return b
}

func GetUserID(ctx echo.Context, j services.JwtService, u services.UserService) (int, *echo.HTTPError) {
	logger := logdoc.GetLogger()

	var userID int

	token := ctx.Request().Header.Get("Authorization")
	var claims jwt.MapClaims

	if token == "" {
		cookie, err := ctx.Cookie("sse_demoToken")
		if cookie != nil || err == nil {
			token = cookie.Value
		}
	}

	if token != "" {
		cl, valid, err := j.ValidateToken(token)
		if err != nil || !valid {
			logger.Error(fmt.Sprintf(">> token validation error, %v@@source=%s", err, common.SourceNameWithLine(runtime.Caller(0))))
			return 0, echo.NewHTTPError(http.StatusForbidden, "invalid token")
		}
		claims = cl
	}
	if claims == nil {
		cookie, err := ctx.Cookie("session_id")
		if err != nil {
			return 0, echo.NewHTTPError(http.StatusForbidden, "нет авторизационной информации")
		}
		if cookie != nil {
			userID, _ = strconv.Atoi(cookie.Value)
		}
	} else {
		if claims["sub"] != nil && claims["sub"] != "" {
			sub := claims["sub"].(string)
			user, err := u.FindUserBySub(sub)
			if err != nil {
				logger.Warn("error getting user by sub, trying by id")
				user, _ = u.FindUserById(int(claims["id"].(float64)))
			}
			userID = user.ID
		} else {
			userID = int(claims["id"].(float64))
		}
	}

	return userID, nil
}

func GetUserIDFromClaims(claims jwt.MapClaims, u services.UserService) (int, error) {
	logger := logdoc.GetLogger()
	var userID int

	if claims["sub"] != nil && claims["sub"] != "" {
		sub := claims["sub"].(string)
		user, err := u.FindUserBySub(sub)
		if err != nil {
			logger.Warn("error getting user by sub, trying by id")
			user, err = u.FindUserById(int(claims["id"].(float64)))
			if err != nil {
				return 0, err
			}
		}
		userID = user.ID
	} else {
		userID = int(claims["id"].(float64))
	}
	return userID, nil
}

func GetRoleFromClaims(claims jwt.MapClaims) string {
	return claims["rol"].(string)
}

func CheckError(data string) bool {
	return data == ""
}

func StringSliceToString(data []string) string {
	return strings.Join(data, ", ")
}

func AIThreadRunToolsSliceToString(data []structs.AIThreadRunTools) string {
	a := make([]string, len(data))
	for i, val := range data {
		a[i] = val.Type
	}
	return strings.Join(a, ", ")
}

func MessageContentSliceToString(data []structs.MessageContent) string {
	a := make([]string, len(data))
	for i, val := range data {
		a[i] = val.Text.Value
	}
	return strings.Join(a, ", ")
}

func CodeInterpretedLogsSliceToString(data []structs.CodeInterpreterOutput) string {
	a := make([]string, len(data))
	for i, val := range data {
		a[i] = val.Logs
	}
	return strings.Join(a, ", ")
}

func GetTokensCount(data string) []int {
	codec, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		return nil
	}
	return codec.Encode(data, nil, nil)
}

func CountMessagesByRole[T structs.UsersCounter](rolesSlice []T, role string) int {
	count := 0
	for _, item := range rolesSlice {
		if item.GetRole() == role && !item.IsHidden() && !item.IsSystem() {
			count++
		}
	}
	return count
}
