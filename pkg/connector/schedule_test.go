package connector

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	ogClient "github.com/opsgenie/opsgenie-go-sdk-v2/client"
	ogSchedule "github.com/opsgenie/opsgenie-go-sdk-v2/schedule"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// mockOpsGenieError is the JSON body the OpsGenie API returns for error responses.
type mockOpsGenieError struct {
	Message   string  `json:"message"`
	Took      float32 `json:"took"`
	RequestId string  `json:"requestId"`
}

func TestScheduleGrants_404ReturnsNotFound(t *testing.T) {
	scheduleName := "IC - IT 911 Only_test Schedule"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/on-calls") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(mockOpsGenieError{
				Message:   "No schedule exists with name [" + scheduleName + "]",
				Took:      0.0,
				RequestId: "test-request-id",
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	// The OpsGenie SDK uses HTTP (not HTTPS) when the apiUrl doesn't contain "api".
	// Stripping the scheme gives us a host:port that satisfies that condition.
	host := strings.TrimPrefix(srv.URL, "http://")

	config := &ogClient.Config{
		ApiKey:         "test-key",
		OpsGenieAPIURL: ogClient.ApiUrl(host),
		RetryCount:     1,
	}

	builder := scheduleBuilder(config)

	resource, err := scheduleResource(&ogSchedule.Schedule{
		Id:   "test-schedule-id",
		Name: scheduleName,
	})
	if err != nil {
		t.Fatalf("failed to build schedule resource: %v", err)
	}

	_, _, _, err = builder.Grants(context.Background(), resource, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if got := status.Code(err); got != codes.NotFound {
		t.Errorf("expected codes.NotFound, got %v (err: %v)", got, err)
	}
}
