package ctxaws_test

import (
	"net/http"
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/mnussbaum/ctxaws"
)

func TestShouldRetry_ContextWithoutTimeout(t *testing.T) {
	ctx := context.Background()

	retryer := ctxaws.NewContextAwareRetryer(ctx, client.DefaultRetryer{5})
	req := &request.Request{
		HTTPResponse: &http.Response{
			StatusCode: http.StatusInternalServerError,
		},
	}
	if !retryer.ShouldRetry(req) {
		t.Error("Request should be retried the context has no deadline")
	}
}

func TestShouldRetry_ContextAlreadyExceeded(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Millisecond)
	time.Sleep(20 * time.Millisecond)

	retryer := ctxaws.NewContextAwareRetryer(ctx, client.DefaultRetryer{5})
	req := &request.Request{}
	if retryer.ShouldRetry(req) {
		t.Error("Request should not be retried after context exceeded")
	}
	if req.Error != context.DeadlineExceeded {
		t.Errorf("Error should be '%s', not '%s'", context.DeadlineExceeded, req.Error)
	}
}

func TestShouldRetry_ContextWouldExpireWhileWaiting(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	retryer := ctxaws.NewContextAwareRetryer(ctx, client.DefaultRetryer{5})
	// a retry count of 10 causes a wait interval of ~58s
	req := &request.Request{RetryCount: 10}
	if retryer.ShouldRetry(req) {
		t.Error("Request should not be retried when the wait period exceeds the context's deadline")
	}
	if req.Error != ctxaws.ErrDeadlineWouldExceedBeforeRetry {
		t.Errorf("Error should be '%s', not '%s'", ctxaws.ErrDeadlineWouldExceedBeforeRetry.Error(), req.Error)
	}
}

func TestShouldRetry_RespectsConfiguredMaxRetries(t *testing.T) {
	ctx := context.Background()

	req := &request.Request{
		HTTPResponse: &http.Response{
			StatusCode: http.StatusInternalServerError,
		},
		Retryer: client.DefaultRetryer{1},
	}
	retryer := ctxaws.NewContextAwareRetryer(ctx, req.Retryer)
	if retryer.MaxRetries() != 1 {
		t.Errorf("Error max retries for ContextAwareRetryer should be 1 not %d", retryer.MaxRetries())
	}
}
