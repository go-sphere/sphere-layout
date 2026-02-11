package shared

import (
	"context"
	"fmt"
	"strconv"

	sharedv1 "github.com/go-sphere/sphere-layout/api/shared/v1"
	"github.com/go-sphere/sphere/storage"
)

var _ sharedv1.StorageServiceHTTPServer = (*Service)(nil)

func (s *Service) UploadToken(ctx context.Context, req *sharedv1.UploadTokenRequest) (*sharedv1.UploadTokenResponse, error) {
	if req.Filename == "" {
		return nil, fmt.Errorf("filename is required")
	}
	id, err := s.GetCurrentID(ctx)
	if err != nil {
		return nil, err
	}
	key := storage.DefaultKeyBuilder(strconv.Itoa(int(id)))
	token, err := s.storage.GenerateUploadAuth(ctx, storage.UploadAuthRequest{
		FileName: key(req.Filename),
		Dir:      s.storageDir,
	})
	if err != nil {
		return nil, err
	}
	return &sharedv1.UploadTokenResponse{
		Token: token.Authorization.Value,
		Key:   token.File.Key,
		Url:   token.File.URL,
	}, nil
}
