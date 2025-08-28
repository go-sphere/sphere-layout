package api

import (
	"context"
	"fmt"

	apiv1 "github.com/go-sphere/sphere-layout/api/api/v1"
	"github.com/go-sphere/sphere-layout/internal/pkg/auth"
	"github.com/go-sphere/sphere-layout/internal/pkg/database/ent"
	"github.com/go-sphere/sphere/utils/idgenerator"
)

var _ apiv1.AuthServiceHTTPServer = (*Service)(nil)

func (s *Service) AuthWithWxMini(ctx context.Context, request *apiv1.AuthWithWxMiniRequest) (*apiv1.AuthWithWxMiniResponse, error) {
	data, err := s.wechat.JsCode2Session(ctx, request.Code)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, fmt.Errorf("failed to get session data from WeChat")
	}
	res, err := auth.Auth(
		ctx, s.db, data.OpenID, auth.PlatformWechatMini,
		auth.WithAuthMode(auth.CreateWithoutCheck),
		auth.WithOnCreateUser(func(user *ent.UserCreate) *ent.UserCreate {
			return user.SetUsername(fmt.Sprintf("wx_%d", idgenerator.NextId()))
		}),
		auth.WithOnCreatePlatform(func(platform *ent.UserPlatformCreate) *ent.UserPlatformCreate {
			return platform.SetSecondID(data.UnionID)
		}),
	)
	if err != nil {
		return nil, err
	}
	token, err := s.authorizer.GenerateToken(ctx, auth.RenderClaims(res.User, res.Platform, auth.AppTokenValidDuration))
	if err != nil {
		return nil, err
	}
	return &apiv1.AuthWithWxMiniResponse{
		IsNew: res.IsNew,
		Token: token,
		User:  s.render.UserFull(res.User),
	}, nil
}
