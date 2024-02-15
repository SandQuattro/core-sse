package multipartchecker

import (
	logdoc "github.com/LogDoc-org/logdoc-go-appender/logrus"
	"github.com/labstack/echo-contrib/jaegertracing"
	"github.com/labstack/echo/v4"
	"net/http"
	"sse-demo-core/internal/app/interfaces/services"
	"sse-demo-core/internal/app/structs"
)

const AUTHORIZATION = "Authorization"

func MultipartCountChecker(service services.JwtService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			span := jaegertracing.CreateChildSpan(ctx, "multipart header middleware")
			defer span.Finish()

			logger := logdoc.GetLogger()
			logger.Debug(">> Multipart Header Middleware started...")

			if ctx.Request().Method == "POST" &&
				(ctx.Request().RequestURI == "/upload" || ctx.Request().RequestURI == "/assistants/files") { // Проверяем наличие заголовка Authorization
				token := ctx.Request().Header.Get(AUTHORIZATION)
				if token == "" {
					cookie, err := ctx.Cookie("sse_demoToken")
					if cookie != nil || err == nil {
						token = cookie.Value
					}
				}
				if token != "" {
					claims, isValid, err := service.ValidateToken(token)
					if !isValid && err != nil {
						logger.Errorf("token validation error: %s", err.Error())
						return echo.NewHTTPError(http.StatusUnauthorized, structs.ErrorResponse{Error: err.Error()})
					}

					// Кладем нужные нам данные в контекст и используем далее в handlers
					ctx.Set("claims", claims)
				} else {
					return echo.NewHTTPError(http.StatusUnauthorized, structs.ErrorResponse{Error: "authorization required"})
				}

				// Получаем данные формы
				err := ctx.Request().ParseMultipartForm(32 << 20)
				if err != nil {
					return echo.NewHTTPError(http.StatusBadRequest, structs.ErrorResponse{Error: err.Error()})
				}

				// Проверяем количество файлов
				if files, ok := ctx.Request().MultipartForm.File["file"]; ok {
					if len(files) > 10 {
						return echo.NewHTTPError(http.StatusBadRequest, structs.ErrorResponse{Error: "only 10 files maximum allowed"})
					}
				} else {
					return echo.NewHTTPError(http.StatusBadRequest, structs.ErrorResponse{Error: "at least 1 file required"})
				}
			}

			logger.Debug(">> Multipart Header Middleware done")
			return next(ctx)
		}
	}
}
