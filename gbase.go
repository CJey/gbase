package gbase

import (
	crand "crypto/rand"
	"encoding/binary"
	mrand "math/rand"
	"time"

	"github.com/google/uuid"
)

var (
	BootID   = uuid.New().String()
	BootTime = time.Now()
)

func init() {
	initMathRand()
	initZapLogger()
}

// initMathRand make math/rand using more easily
func initMathRand() {
	var buf = make([]byte, 8)
	if _, err := crand.Read(buf); err != nil {
		panic(err)
	}
	var seed = binary.BigEndian.Uint64(buf)
	mrand.Seed(int64(seed))
}

// initZapLogger should used when your process startup,
// and you should replace it with a new logger policy after parsing your config.
// If you do not like it's style, just use zap.ReplaceGlobals() by you self, before using Context
func initZapLogger() {
	var _, err = ReplaceZapLogger("debug", "stderr", "console", false)
	if err != nil {
		panic(err)
	}
}
