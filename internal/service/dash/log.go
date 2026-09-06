package dash

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-sphere/httpx"
	dashv1 "github.com/go-sphere/sphere-layout/api/dash/v1"
	"github.com/go-sphere/sphere/log"
	"github.com/go-sphere/sphere/log/logbuffer"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	defaultLogTailLimit    = 500
	defaultLogHistoryLimit = 100
	maxLogBatchSize        = 64
)

var _ dashv1.LogServiceHTTPServer = (*Service)(nil)

func (s *Service) Tail(ctx context.Context, request *dashv1.TailRequest, send func(*dashv1.TailResponse) error) error {
	level, err := parseLogLevel(request.MinLevel)
	if err != nil {
		return err
	}
	tailLimit := int(request.TailLimit)
	if tailLimit == 0 {
		tailLimit = defaultLogTailLimit
	}
	sub := s.logs.Subscribe(logbuffer.SubscribeOptions{
		StreamID:  request.StreamId,
		FromSeq:   request.FromSeq,
		TailLimit: tailLimit,
		MinLevel:  level,
	})
	defer sub.Cancel()

	// An empty cursor frame is required when backfill is empty so the lazy
	// SSE response commits before the next log, and heartbeats can start.
	if len(sub.Backfill) == 0 {
		lastSeq := request.FromSeq
		if sub.Reset {
			lastSeq = 0
		}
		if err := send(&dashv1.TailResponse{
			StreamId:  s.logs.ID(),
			LastSeq:   lastSeq,
			Truncated: sub.Truncated,
			Reset_:    sub.Reset,
		}); err != nil {
			return err
		}
	}
	for start := 0; start < len(sub.Backfill); start += maxLogBatchSize {
		end := min(start+maxLogBatchSize, len(sub.Backfill))
		response := newTailLogsResponse(s.logs.ID(), sub.Backfill[start:end])
		if start == 0 {
			response.Truncated = sub.Truncated
			response.Reset_ = sub.Reset
		}
		if err := send(response); err != nil {
			return err
		}
	}

	var reportedDropped uint64
	for {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case first := <-sub.C:
			batch := []logbuffer.Entry{first}
		drain:
			for len(batch) < maxLogBatchSize {
				select {
				case entry := <-sub.C:
					batch = append(batch, entry)
				default:
					break drain
				}
			}
			response := newTailLogsResponse(s.logs.ID(), batch)
			dropped := sub.Dropped()
			response.Dropped = dropped - reportedDropped
			reportedDropped = dropped
			if err := send(response); err != nil {
				return err
			}
		}
	}
}

func (s *Service) History(_ context.Context, request *dashv1.HistoryRequest) (*dashv1.HistoryResponse, error) {
	level, err := parseLogLevel(request.MinLevel)
	if err != nil {
		return nil, err
	}
	limit := int(request.Limit)
	if limit == 0 {
		limit = defaultLogHistoryLimit
	}
	entries, more := s.logs.History(request.BeforeSeq, limit, level)
	return &dashv1.HistoryResponse{
		StreamId:  s.logs.ID(),
		Entries:   logEntriesToProto(entries),
		More:      more,
		LatestSeq: s.logs.LatestSeq(),
	}, nil
}

func parseLogLevel(value string) (log.Level, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return log.LevelDebug, nil
	}
	level, ok := logbuffer.ParseLevel(value)
	if !ok {
		return log.LevelDebug, httpx.NewBadRequestError("min_level must be debug, info, warn, or error")
	}
	return level, nil
}

func newTailLogsResponse(streamID string, entries []logbuffer.Entry) *dashv1.TailResponse {
	response := &dashv1.TailResponse{
		StreamId: streamID,
		Entries:  logEntriesToProto(entries),
	}
	if len(entries) > 0 {
		response.LastSeq = entries[len(entries)-1].Seq
	}
	return response
}

func logEntriesToProto(entries []logbuffer.Entry) []*dashv1.LogEntry {
	out := make([]*dashv1.LogEntry, 0, len(entries))
	for _, entry := range entries {
		out = append(out, &dashv1.LogEntry{
			Seq:     entry.Seq,
			Time:    timestamppb.New(entry.Time),
			Level:   entry.Level,
			Name:    entry.Name,
			Message: entry.Message,
			Attrs:   attrsToProto(entry.Attrs),
		})
	}
	return out
}

func attrsToProto(attrs map[string]any) (out *structpb.Struct) {
	if len(attrs) == 0 {
		return nil
	}
	defer func() {
		if value := recover(); value != nil {
			out = attrError(fmt.Sprint(value))
		}
	}()
	raw, err := json.Marshal(attrs)
	if err != nil {
		return attrError(err.Error())
	}
	var normalized map[string]any
	if err := json.Unmarshal(raw, &normalized); err != nil {
		return attrError(err.Error())
	}
	out, err = structpb.NewStruct(normalized)
	if err != nil {
		return attrError(err.Error())
	}
	return out
}

func attrError(message string) *structpb.Struct {
	out, _ := structpb.NewStruct(map[string]any{"attr_error": message})
	return out
}
