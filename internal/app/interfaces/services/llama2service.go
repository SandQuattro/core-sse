package services

import (
	"sse-demo-core/internal/app/structs"
)

type Llama2Service interface {
	ExecuteAnalyze(layers *[]structs.UserLayer, prefix string) (*structs.HuggingFaceResponse, error)
	AskLlama2Batch(message string) (*structs.HuggingFaceResponse, error)
}
