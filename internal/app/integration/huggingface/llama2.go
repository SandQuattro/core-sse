package llama2

import (
	"fmt"
	logdoc "github.com/LogDoc-org/logdoc-go-appender/logrus"
	"github.com/go-resty/resty/v2"
	"github.com/gurkankaymak/hocon"
	"sse-demo-core/internal/app/structs"
	"sse-demo-core/internal/app/utils"
	"strings"
	"time"
)

type ServiceImpl struct {
	config *hocon.Config
	client *resty.Client
}

func New(config *hocon.Config) *ServiceImpl {
	return &ServiceImpl{
		config: config,
		client: resty.New(),
	}
}

func (o *ServiceImpl) ExecuteAnalyze(layers *[]structs.UserLayer, prefix string) (*structs.HuggingFaceResponse, error) {
	logger := logdoc.GetLogger()

	logger.Debug("Executing hugging face llama2 integration...")
	var totalData strings.Builder

	totalData.WriteString(prefix)
	for _, val := range *layers {
		totalData.WriteString(utils.Ternary(val.OptimizedData == "", val.SourceData, val.OptimizedData).(string))
	}

	result, err := o.AskLlama2Batch(totalData.String())
	if err != nil {
		logger.Error("Executing hugging face llama2 integration error, ", err)
		return nil, err
	}

	return result, nil
}

func (o *ServiceImpl) AskLlama2Batch(message string) (*structs.HuggingFaceResponse, error) {
	logger := logdoc.GetLogger()

	logger.Debug("Executing openai integration...")

	huggingFaceReq := structs.HuggingFaceRequest{
		Inputs:  message,
		Stream:  false,
		Options: structs.HuggingFaceOptions{UseCache: false},
	}

	var data structs.HuggingFaceResponse

	_, err := retryer(func() (*resty.Response, error) {
		response, err := o.client.SetPreRequestHook(utils.CurlLogger).R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Authorization", o.config.GetString("integration.huggingface.token")).
			SetBody(huggingFaceReq).
			SetResult(&data).
			Post(o.config.GetString("integration.huggingface.proto") + "://" +
				o.config.GetString("integration.huggingface.host") +
				o.config.GetString("integration.huggingface.uri.models") +
				o.config.GetString("integration.huggingface.uri.chat"))
		if err != nil {
			logger.Error("Ошибка формирования запроса в openai, response:", string(response.Body()), ", err:", err)
			return nil, err
		}
		return response, nil
	}, 3, 1*time.Minute)

	if err != nil {
		logger.Error("Ошибка выполнения запроса в hugging face llama2, ", err)
		return nil, err
	}

	logger.Debug("Got openai response successfully, response:", data[0].GeneratedText)
	return &data, nil
}

// timeout and retry pattern implementation
func retryer(fn func() (*resty.Response, error), maxRetries int, initialRetryInterval time.Duration) (*resty.Response, error) {
	logger := logdoc.GetLogger()

	retryInterval := initialRetryInterval
	for i := 0; i < maxRetries; i++ {
		resp, err := fn()
		if err == nil {
			return resp, nil
		}
		logger.Error(fmt.Errorf("unable to call an openai service, attempt %d ", i+1))
		time.Sleep(retryInterval)
		retryInterval *= 2 // увеличиваем время ожидания в 2 раза с каждой итерацией
	}
	return nil, fmt.Errorf("unable to complete task after %d attempts", maxRetries)
}
