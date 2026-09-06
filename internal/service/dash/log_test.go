package dash

import (
	"context"
	"errors"
	"testing"

	dashv1 "github.com/go-sphere/sphere-layout/api/dash/v1"
	"github.com/go-sphere/sphere/log"
	"github.com/go-sphere/sphere/log/logbuffer"
)

func TestLogHistory(t *testing.T) {
	logs := logbuffer.New(4)
	logs.Log(t.Context(), log.LevelInfo, "first", log.String("request_id", "one"))
	logs.Log(t.Context(), log.LevelError, "second")
	service := &Service{logs: logs}

	response, err := service.History(t.Context(), &dashv1.HistoryRequest{
		Limit:    1,
		MinLevel: "warn",
	})
	if err != nil {
		t.Fatalf("History: %v", err)
	}
	if response.StreamId != logs.ID() || response.LatestSeq != 2 {
		t.Fatalf("cursor metadata = %#v, want stream ID and latest seq 2", response)
	}
	if len(response.Entries) != 1 || response.Entries[0].Message != "second" {
		t.Fatalf("Entries = %#v, want the error entry", response.Entries)
	}
}

func TestLogTailBackfillAndReset(t *testing.T) {
	logs := logbuffer.New(4)
	logs.Log(t.Context(), log.LevelInfo, "first")
	logs.Log(t.Context(), log.LevelInfo, "second")
	service := &Service{logs: logs}
	ctx, cancel := context.WithCancel(t.Context())
	var responses []*dashv1.TailResponse

	err := service.Tail(ctx, &dashv1.TailRequest{
		StreamId:  "old-process",
		FromSeq:   99,
		TailLimit: 1,
	}, func(response *dashv1.TailResponse) error {
		responses = append(responses, response)
		cancel()
		return nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Tail error = %v, want context canceled", err)
	}
	if len(responses) != 1 {
		t.Fatalf("response count = %d, want 1", len(responses))
	}
	response := responses[0]
	if !response.Reset_ || response.Truncated || response.StreamId != logs.ID() {
		t.Fatalf("response cursor metadata = %#v", response)
	}
	if len(response.Entries) != 1 || response.Entries[0].Seq != 2 || response.LastSeq != 2 {
		t.Fatalf("response entries = %#v, want seq 2", response)
	}
}

func TestLogTailSendsInitialFrameWhenEmpty(t *testing.T) {
	logs := logbuffer.New(4)
	service := &Service{logs: logs}
	ctx, cancel := context.WithCancel(t.Context())
	var initial *dashv1.TailResponse

	err := service.Tail(ctx, &dashv1.TailRequest{}, func(response *dashv1.TailResponse) error {
		initial = response
		cancel()
		return nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Tail error = %v, want context canceled", err)
	}
	if initial == nil || initial.StreamId != logs.ID() || len(initial.Entries) != 0 || initial.LastSeq != 0 {
		t.Fatalf("initial response = %#v, want empty cursor frame with last_seq 0", initial)
	}
}

func TestLogTailEmptyFrameKeepsCallerCursor(t *testing.T) {
	logs := logbuffer.New(4)
	logs.Log(t.Context(), log.LevelInfo, "first")
	logs.Log(t.Context(), log.LevelDebug, "second")
	service := &Service{logs: logs}
	ctx, cancel := context.WithCancel(t.Context())
	var initial *dashv1.TailResponse

	err := service.Tail(ctx, &dashv1.TailRequest{
		FromSeq:  1,
		MinLevel: "error",
	}, func(response *dashv1.TailResponse) error {
		initial = response
		cancel()
		return nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Tail error = %v, want context canceled", err)
	}
	if initial == nil || len(initial.Entries) != 0 || initial.LastSeq != 1 {
		t.Fatalf("initial response = %#v, want empty frame with last_seq 1", initial)
	}
	if logs.LatestSeq() != 2 {
		t.Fatalf("LatestSeq = %d, want 2 so LatestSeq would over-claim", logs.LatestSeq())
	}
}

func TestLogTailResetEmptyFrameClearsSeq(t *testing.T) {
	logs := logbuffer.New(4)
	service := &Service{logs: logs}
	ctx, cancel := context.WithCancel(t.Context())
	var initial *dashv1.TailResponse

	err := service.Tail(ctx, &dashv1.TailRequest{
		StreamId: "old-process",
		FromSeq:  99,
	}, func(response *dashv1.TailResponse) error {
		initial = response
		cancel()
		return nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Tail error = %v, want context canceled", err)
	}
	if initial == nil || !initial.Reset_ || initial.LastSeq != 0 || len(initial.Entries) != 0 {
		t.Fatalf("initial response = %#v, want reset empty frame with last_seq 0", initial)
	}
}

func TestInvalidLogLevel(t *testing.T) {
	service := &Service{logs: logbuffer.New(1)}
	_, err := service.History(t.Context(), &dashv1.HistoryRequest{MinLevel: "verbose"})
	if err == nil {
		t.Fatal("History error = nil, want invalid min_level error")
	}
}

func TestLogAttrsThatCannotEncodeBecomeDiagnostic(t *testing.T) {
	logs := logbuffer.New(1)
	logs.Log(t.Context(), log.LevelInfo, "entry", log.Any("channel", make(chan int)))
	service := &Service{logs: logs}

	response, err := service.History(t.Context(), &dashv1.HistoryRequest{})
	if err != nil {
		t.Fatalf("History: %v", err)
	}
	attrs := response.Entries[0].Attrs.AsMap()
	message, ok := attrs["attr_error"].(string)
	if !ok || message == "" {
		t.Fatalf("Attrs = %#v, want attr_error diagnostic", attrs)
	}
}
