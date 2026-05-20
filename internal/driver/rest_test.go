package driver

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ghazlabs/wa-scheduler/internal/core"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

type retryMessageMockService struct{}

func (m *retryMessageMockService) InitializeService(ctx context.Context) {}

func (m *retryMessageMockService) GetAllMessages(ctx context.Context, input core.GetAllMessagesInput) ([]core.Message, error) {
	return nil, nil
}

func (m *retryMessageMockService) SendMessage(ctx context.Context, input core.ScheduleMessageInput) error {
	return nil
}

func (m *retryMessageMockService) RetryMessage(ctx context.Context, input core.RetryMessageInput) error {
	return core.ErrMessageNotFound
}

func newRetryTestAPI() *API {
	api, _ := NewAPI(APIConfig{
		Service:            &retryMessageMockService{},
		ClientUsername:     "admin",
		ClientPassword:     "admin",
		WebClientPublicDir: ".",
	})
	return api
}

func parseBody(w *httptest.ResponseRecorder) map[string]interface{} {
	var body map[string]interface{}
	json.NewDecoder(w.Body).Decode(&body)
	return body
}

func TestRetryMessage_MessageNotFound_Returns404(t *testing.T) {
	api := newRetryTestAPI()

	reqBody := RetryMessageRequest{
		ScheduledSendingAt: 1234567890,
	}

	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(
		http.MethodPost,
		"/messages/200/retry",
		bytes.NewBuffer(jsonBody),
	)

	req.SetBasicAuth("admin", "admin")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "200")

	req = req.WithContext(
		context.WithValue(req.Context(), chi.RouteCtxKey, rctx),
	)

	w := httptest.NewRecorder()

	api.serveRetryMessage(w, req)

	body := parseBody(w)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.False(t, body["ok"].(bool))
	assert.Equal(t, "ERR_MESSAGES_NOT_FOUND", body["err"])
	assert.Equal(t, "message id not found", body["msg"])
}
