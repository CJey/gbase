package gbase

import (
	gcontext "context"
	"fmt"
	"testing"
)

var _ = fmt.Print

func TestSessionContext(t *testing.T) {
	var ctx = SessionContext().At("TestSessionContext")
	ctx.Info("Bingo!", "username", "cjey", "sex", "male", "session", GetSession(ctx))

	ctx = ctx.At("At")
	ctx.Warn("At test")

	var ctx1 = ctx.Fork()
	ctx1.Set("abc", "xyz")
	ctx1.Warn("Fork test 1", "abc", ctx1.GetString("abc"))
	var ctx2 = ctx.Fork()
	ctx2.Warn("Fork test 2", "abc", ctx2.GetString("abc"))

	var ctx11 = ctx1.Fork()
	ctx11.Warn("Fork test 1.1", "name", ctx11.Name(), "real session", GetRealSession(ctx11))
	var ctx12 = ctx1.Fork()
	var session = "7ec17674-1360-4fb1-9245-bd8d8d5866c4"
	SetSession(ctx12, session)
	ctx12.Warn("Fork test 1.2", "name", ctx12.Name(), "cutomized session", GetSession(ctx12))
	var ctx13 = ctx1.Fork()
	ctx13.Warn("Fork test 1.3", "name", ctx13.Name(), "session", GetSession(ctx13))
}

func TestNewNamedContext(t *testing.T) {
	var ctx = NamedContext("go test").At("TestNewNamedContext")
	ctx.Info("Bingo!", "username", "cjey", "sex", "male")
}

func TestToSessionContext(t *testing.T) {
	var ctx = ToSessionContext(gcontext.Background()).At("TestToContext")
	ctx.Info("Bingo!", "username", "cjey", "sex", "male")
}

func TestToNamedContext(t *testing.T) {
	var ctx = ToNamedContext(gcontext.Background(), "go test").At("TestToNamedContext")
	ctx.Info("Bingo!", "username", "cjey", "sex", "male")
}

func TestSession(t *testing.T) {
	var ctx = SessionContext().At("TestSession")
	var session = "7ec17674-1360-4fb1-9245-bd8d8d5866c4"
	SetSession(ctx, session)
	if got := GetSession(ctx); got == session {
		ctx.Info("ok")
	} else {
		ctx.Fatal("fail", "origin", session, "got", got)
	}
}
