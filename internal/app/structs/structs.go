package structs

import (
	"database/sql"
	"encoding/json"
	"time"
)

type Roles string

type APIError struct {
	Status  int
	Message string
}

func (e APIError) Error() string {
	return e.Message
}

type OpenAIError struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Param   string `json:"param"`
		Code    string `json:"code"`
	} `json:"error"`
}

type OpenAIRequest struct {
	Model    string    `json:"model"`
	Stream   bool      `json:"stream"`
	Messages []Content `json:"messages"`
}

type GigaChatRequest struct {
	Model    string    `json:"model"`
	Stream   bool      `json:"stream"`
	Messages []Content `json:"messages"`
}

type GigaChatToken struct {
	Token   string `json:"access_token"`
	Expires int64  `json:"expires_at"`
}

type Content struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Usage   Usage    `json:"usage"`
	Choices []Choice `json:"choices"`
}

type GigaChatResponse struct {
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Object  string   `json:"object"`
	Usage   Usage    `json:"usage"`
	Choices []Choice `json:"choices"`
}

type AssistantFiles struct {
	Object string `json:"object"`
	Data   []struct {
		ID          string `json:"id"`
		Object      string `json:"object"`
		CreatedAt   int    `json:"created_at"`
		AssistantID string `json:"assistant_id"`
	} `json:"data"`
	FirstID string `json:"first_id"`
	LastID  string `json:"last_id"`
	HasMore bool   `json:"has_more"`
}

type AssistantsFileUploadResponse struct {
	Object        string `json:"object"`
	ID            string `json:"id"`
	Purpose       string `json:"purpose"`
	Filename      string `json:"filename"`
	Bytes         int    `json:"bytes"`
	CreatedAt     int    `json:"created_at"`
	Status        string `json:"status"`
	StatusDetails any    `json:"status_details"`
}

type LinkRequest struct {
	FileID string `json:"file_id"`
}

type ThreadMessageRequest struct {
	Role    string   `json:"role"`
	Content string   `json:"content"`
	FileIDS []string `json:"file_ids"`
}

type ThreadRunRequest struct {
	AssistantID string `json:"assistant_id"`
}

type AssistantsLinkFileToAssistant struct {
	ID          string `json:"id"`
	Object      string `json:"object"`
	CreatedAt   int    `json:"created_at"`
	AssistantID string `json:"assistant_id"`
}

type AssistantsThread struct {
	ID        string      `json:"id"`
	Object    string      `json:"object"`
	CreatedAt int         `json:"created_at"`
	Metadata  interface{} `json:"metadata"`
}

type Thread struct {
	ID         int         `db:"id" json:"id"`
	UserID     int         `db:"user_id" json:"user_id"`
	ThreadName string      `db:"thread_name" json:"thread_name"`
	ThreadID   string      `db:"thread_id" json:"thread_id"`
	Object     string      `db:"object" json:"object"`
	CreatedAt  time.Time   `db:"created_at" json:"created_at"`
	Metadata   interface{} `db:"metadata" json:"metadata"`
	Messages   int         `db:"messages_cnt" json:"messages"`
}

type ThreadFile struct {
	ID        string    `db:"id" json:"-"`
	UserID    int       `db:"user_id" json:"user_id"`
	ThreadID  string    `db:"thread_id" json:"thread_id"`
	FileID    string    `db:"file_id" json:"file_id"`
	FileName  string    `db:"file_name" json:"file_name"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	Purpose   string    `db:"purpose" json:"purpose"`
	Bytes     int       `db:"bytes" json:"bytes"`
}

type ThreadRun struct {
	ID           string    `json:"id"`
	RunID        string    `json:"thread_run_id"`
	Object       string    `json:"object"`
	CreatedAt    time.Time `json:"created_at"`
	AssistantID  string    `json:"assistant_id"`
	ThreadID     string    `json:"thread_id"`
	Status       string    `json:"status"`
	StartedAt    time.Time `json:"started_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	CancelledAt  time.Time `json:"cancelled_at"`
	FailedAt     time.Time `json:"failed_at"`
	CompletedAt  time.Time `json:"completed_at"`
	LastError    string    `json:"last_error"`
	Model        string    `json:"model"`
	Instructions string    `json:"instructions"`
	Tools        string    `json:"tools"`
	FileIds      string    `json:"file_ids"`
	Metadata     string    `json:"metadata"`
}

type AIThreadRun struct {
	ID           string             `json:"id"`
	Object       string             `json:"object"`
	CreatedAt    int                `json:"created_at"`
	AssistantID  string             `json:"assistant_id"`
	ThreadID     string             `json:"thread_id"`
	Status       string             `json:"status"`
	StartedAt    int                `json:"started_at"`
	ExpiresAt    int                `json:"expires_at"`
	CancelledAt  int                `json:"cancelled_at"`
	FailedAt     int                `json:"failed_at"`
	CompletedAt  int                `json:"completed_at"`
	LastError    LastError          `json:"last_error"`
	Model        string             `json:"model"`
	Instructions string             `json:"instructions"`
	Tools        []AIThreadRunTools `json:"tools"`
	FileIds      []string           `json:"file_ids"`
	Metadata     struct {
	} `json:"metadata"`
}

type AIThreadRunTools struct {
	Type string `json:"type"`
}

type LastError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type MessageContent struct {
	Type string `json:"type"`
	Text struct {
		Value       string `json:"value"`
		Annotations []any  `json:"annotations"`
	} `json:"text"`
}

type ListAIMessages struct {
	Object  string      `json:"object"`
	Data    []AIMessage `json:"data"`
	FirstID string      `json:"first_id"`
	LastID  string      `json:"last_id"`
	HasMore bool        `json:"has_more"`
}

type ListThreadMessages struct {
	Object  string          `json:"object"`
	Data    []ThreadMessage `json:"data"`
	FirstID string          `json:"first_id"`
	LastID  string          `json:"last_id"`
	HasMore bool            `json:"has_more"`
}

type AIMessage struct {
	ID          string           `json:"id"`
	Object      string           `json:"object"`
	CreatedAt   int              `json:"created_at"`
	ThreadID    string           `json:"thread_id"`
	Role        string           `json:"role"`
	Content     []MessageContent `json:"content"`
	FileIds     []string         `json:"file_ids"`
	AssistantID string           `json:"assistant_id"`
	RunID       string           `json:"run_id"`
	Metadata    struct{}         `json:"metadata"`
}

type ThreadMessage struct {
	ID          string          `db:"id" json:"id"`
	UserID      string          `db:"user_id" json:"user_id"`
	Object      string          `db:"object" json:"object"`
	CreatedAt   int             `db:"created_at" json:"created_at"`
	ThreadID    string          `db:"thread_id" json:"thread_id"`
	Role        string          `db:"role" json:"role"`
	Prompt      string          `db:"prompt" json:"prompt"`
	AssistantID string          `db:"assistant_id" json:"assistant_id"`
	RunID       string          `db:"run_id" json:"run_id"`
	Content     json.RawMessage `db:"content" json:"content"`
	FileIDs     json.RawMessage `db:"file_ids" json:"file_ids"`
	Tokens      int             `db:"tokens" json:"tokens"`
	Hidden      bool            `db:"hidden" json:"hidden"`
	System      bool            `db:"system" json:"system"`
	Metadata    any             `db:"metadata" json:"metadata"`
}

type UserMessagesStatistics struct {
	Email             string `json:"email" db:"email"`
	UserID            int    `json:"user_id" db:"user_id"`
	CurrentMonthCount int    `json:"current_month_count" db:"current_month_count"`
	LastMonthCount    int    `json:"last_month_count" db:"last_month_count"`
	YearlyCount       int    `json:"yearly_count" db:"year_count"`
	TotalCount        int    `json:"total_count" db:"total_count"`
}

type UsersCounter interface {
	GetRole() string
	IsHidden() bool
	IsSystem() bool
}

func (ai AIMessage) GetRole() string {
	return ai.Role
}

func (m ThreadMessage) GetRole() string {
	return m.Role
}

func (ai AIMessage) IsHidden() bool {
	return false
}

func (m ThreadMessage) IsHidden() bool {
	return m.Hidden
}

func (ai AIMessage) IsSystem() bool {
	return false
}

func (m ThreadMessage) IsSystem() bool {
	return m.System
}

type MessageText struct {
	Value       string `json:"value"`
	Annotations []any  `json:"annotations"`
}

type DeletedObject struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Deleted bool   `json:"deleted"`
}

type AIRunStepDetails struct {
	Object  string      `json:"object"`
	Data    []AIRunStep `json:"data"`
	FirstID string      `json:"first_id"`
	LastID  string      `json:"last_id"`
	HasMore bool        `json:"has_more"`
}

type AIRunStep struct {
	ID          string    `json:"id"`
	Object      string    `json:"object"`
	CreatedAt   int       `json:"created_at"`
	RunID       string    `json:"run_id"`
	AssistantID string    `json:"assistant_id"`
	ThreadID    string    `json:"thread_id"`
	Type        string    `json:"type"`
	Status      string    `json:"status"`
	CancelledAt int       `json:"cancelled_at"`
	CompletedAt int       `json:"completed_at"`
	ExpiresAt   int       `json:"expires_at"`
	FailedAt    int       `json:"failed_at"`
	LastError   LastError `json:"last_error"`
	StepDetails struct {
		Type            string `json:"type"`
		MessageCreation struct {
			MessageID string `json:"message_id,omitempty"`
		} `json:"message_creation,omitempty"`
		ToolCalls []struct {
			ID              string `json:"id,omitempty"`
			Type            string `json:"type,omitempty"`
			CodeInterpreter struct {
				Input   string                  `json:"input,omitempty"`
				Outputs []CodeInterpreterOutput `json:"outputs,omitempty"`
			} `json:"code_interpreter,omitempty"`
		} `json:"tool_calls,omitempty"`
	} `json:"step_details,omitempty"`
}

type RunStep struct {
	ID          string    `json:"id"`
	Object      string    `json:"object"`
	CreatedAt   time.Time `json:"created_at"`
	AssistantID string    `json:"assistant_id"`
	RunID       string    `json:"run_id"`
	Type        string    `json:"type"`
	Status      string    `json:"status"`
	StartedAt   time.Time `json:"started_at"`
	CancelledAt time.Time `json:"cancelled_at"`
	CompletedAt time.Time `json:"completed_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	FailedAt    time.Time `json:"failed_at"`
	LastError   string    `json:"last_error"`
	Tool        string    `json:"tool"`
	Message     string    `json:"message"`
}

type CodeInterpreterOutput struct {
	Type string `json:"type,omitempty"`
	Logs string `json:"logs,omitempty"`
}

type OpenAIStreamingResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

type HuggingFaceRequest struct {
	Inputs  string             `json:"inputs"`
	Stream  bool               `json:"stream"`
	Options HuggingFaceOptions `json:"options"`
}

type HuggingFaceOptions struct {
	UseCache bool `json:"use_cache"`
}

type HuggingFaceResponse []struct {
	GeneratedText string `json:"generated_text"`
}

type CreateUser struct {
	TenantID int    `json:"tenant_id" validate:"required"`
	Last     string `json:"lastName" validate:"required"`
	First    string `json:"firstName" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	Role     Roles  `db:"role" validate:"required"`
}

type IncomingUser struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type User struct {
	ID            int    `db:"id"`
	Sub           string `db:"sub"`
	Name          string `db:"name"`
	GivenName     string `db:"given_name"`
	FamilyName    string `db:"family_name"`
	Avatar        string `db:"avatar"`
	Locale        string `db:"locale"`
	Email         string `db:"email" validate:"required,email"`
	EmailVerified bool   `db:"email_verified"`
	Password      []byte `db:"hashed_password" validate:"required" json:"-"`
	Role          string `db:"role" validate:"required"`
}

type UserLayer struct {
	ID             int       `db:"id"`
	GUID           string    `db:"guid"`
	UUID           string    `db:"uuid"`
	UserID         int       `db:"user_id"`
	LayerName      string    `db:"layer_name"`
	SourceSize     int       `db:"source_size"`
	SourceName     string    `db:"source_name"`
	SourceType     string    `db:"source_type"`
	SourceFileLink string    `db:"source_file_link"`
	SourceData     string    `db:"source_data"`
	OptimizedData  string    `db:"optimized_data"`
	OpenaiFileID   string    `db:"openai_file_id"`
	AssistantID    string    `db:"assistant_id"`
	ThreadID       string    `db:"thread_id"`
	Loaded         time.Time `db:"loaded"`
	Status         string    `db:"status"`
}

type Prompt struct {
	ID          int    `db:"id"`
	PromptText  string `json:"prompt_text" db:"prompt_text" validate:"required"`
	PromptStage int    `json:"prompt_stage" db:"prompt_stage" validate:"required"`
}

type ResponseUser struct {
	ID    int    `json:"-"`
	First string `json:"firstName" validate:"required"`
	Last  string `json:"lastName" validate:"required"`
	Email string `json:"email" validate:"required,email"`
	Role  string `json:"role" validate:"required"`
}

type Token struct {
	Token string `json:"token" xml:"token"`
}

type AuthRes struct {
	Token string
}

type ErrorResponse struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

type Usage struct {
	PromptToken      int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type Choice struct {
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
	Index        int     `json:"index"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIModelsResponse struct {
	Data   []Model `json:"data"`
	Object string  `json:"object"`
}

type GigaChatModelsResponse struct {
	Data   []Model `json:"data"`
	Object string  `json:"object"`
}

type Model struct {
	ID         string       `json:"id"`
	Object     string       `json:"object"`
	OwnedBy    string       `json:"owned_by"`
	Permission []Permission `json:"permission"`
}

type Permission struct {
	ID                 string `json:"id"`
	AllowCreateEngine  bool   `json:"allow_create_engine"`
	AllowSampling      bool   `json:"allow_sampling"`
	AllowLogprobs      bool   `json:"allow_logprobs"`
	AllowSearchIndices bool   `json:"allow_search_indices"`
	AllowView          bool   `json:"allow_view"`
	AllowFineTuning    bool   `json:"allow_fine_tuning"`
	IsBlocking         bool   `json:"is_blocking"`
	Organization       string `json:"organization"`
	Group              string `json:"group"`
	Object             string `json:"object"`
	Created            int    `json:"created"`
}

type StorageUploadResponse struct {
	FilePath string `json:"filePath" xml:"filePath"`
	Result   string `json:"result" xml:"result"`
}

type StorageFileInfoResponse struct {
	ID           int            `db:"id" validate:"required"`
	Name         string         `db:"file_name" validate:"required"`
	UploadStatus string         `db:"upload_status" validate:"required"`
	StorageLink  sql.NullString `db:"storage_link"`
}

type Notification struct {
	GUID     string
	UUID     string
	State    string
	FileName string
}

type AnalyseData struct {
	Title string `json:"title"`
	Data  string `json:"data"`
	Error bool   `json:"error"`
}

type SSEvent struct {
	UUID  string `json:"uuid"`
	Event string `json:"event"`
	Error bool   `json:"error"`
	Data  any    `json:"data"`
}
