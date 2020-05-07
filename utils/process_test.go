package utils

import (
	"testing"
)

func TestRenameMyProcess(t *testing.T) {
	var ctx = SimpleContext()
	RenameMyProcess(ctx, "Hello")
}
