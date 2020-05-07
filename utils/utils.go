package utils

import (
	gcontext "context"

	"go.uber.org/zap"

	"github.com/cjey/gbase/context"
)

func SimpleContext() context.Context {
	var logger, _ = zap.NewDevelopment()
	return context.New(
		gcontext.Background(),
		context.NewEnv(),
		context.NewLogger("", "", logger.Sugar(), nil, nil),
	)
}
