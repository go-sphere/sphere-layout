package render

import (
	"errors"
	"strings"

	"buf.build/go/protovalidate"
	"github.com/go-sphere/sphere-layout/internal/pkg/database/ent"
	"github.com/go-sphere/sphere/database/mapper"
	"github.com/go-sphere/sphere/server/ginx"
)

func init() {
	ginx.SetDefaultErrorParser(func(err error) (int32, int32, string) {
		var ve *protovalidate.ValidationError
		if errors.As(err, &ve) {
			return ValidationError(ve)
		}
		var ne *ent.NotFoundError
		if errors.As(err, &ne) {
			return EntNotFoundError(ne)
		}
		var ce *ent.ConstraintError
		if errors.As(err, &ce) {
			return EntConstraintError(ce)
		}
		return ginx.ParseError(err)
	})
}

func ValidationError(err *protovalidate.ValidationError) (int32, int32, string) {
	return 0, 400, strings.Join(mapper.Map(err.Violations, func(s *protovalidate.Violation) string {
		return s.Proto.GetMessage()
	}), ",")
}

func EntNotFoundError(err *ent.NotFoundError) (int32, int32, string) {
	return 0, 404, err.Error()
}

func EntConstraintError(err *ent.ConstraintError) (int32, int32, string) {
	return 0, 400, err.Unwrap().Error()
}
