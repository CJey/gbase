package context

import (
	"testing"

	"go.uber.org/zap"
)

func TestLogger(t *testing.T) {
	var l, _ = zap.NewDevelopmentConfig().Build()
	var L = NewLogger("", "", l.Sugar(), nil, nil)

	L = L.With("A", "B")
	L.Info("Yes")
	{
		var l = L.(*logger)
		if len(l.with) != 2 || l.with[0] != "A" || l.with[1] != "B" {
			t.Error("logger with fail")
		}
	}
	L = L.With("C", "D")
	L.Info("Yes")
	{
		var l = L.(*logger)
		if len(l.with) != 4 || l.with[0] != "A" || l.with[1] != "B" || l.with[2] != "C" || l.with[3] != "D" {
			t.Error("logger with fail")
		}
	}
}
