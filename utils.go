package gbase

import (
	"net"
	"os"
	"sync"
	"time"

	"github.com/cjey/gbase/utils"
)

// ResolveTCPAddr use 0.0.0.0 as default ip
// to completion the given raw ip
// ""      => 0.0.0.0:defaultPort
// 80      => 0.0.0.0:80
// 1.1.1.1 => 1.1.1.1:defaultPort
func ResolveTCPAddr(ctx Context, raw string, defaultPort uint16) (*net.TCPAddr, error) {
	return utils.ResolveTCPAddr(raw, nil, defaultPort)
}

// WriteMyPID write current process pid to /var/run/<appname>.pid
// If write failed, but the file content is equal to current process id,
// that's also ok. Any error will be ignored.
func WriteMyPID(ctx Context, appname string) {
	utils.WritePID(ctx, "/var/run/"+appname+".pid", os.Getpid())
}

var _LiveProcessName = struct {
	sync.Mutex
	stop chan struct{}
}{}

// LiveProcessName change current process name every interval duration.
// You can use this for exposing your core runtime information to maintainer.
func LiveProcessName(ctx Context, interval time.Duration, liveName func(int) string) chan struct{} {
	_LiveProcessName.Lock()
	defer _LiveProcessName.Unlock()

	if _LiveProcessName.stop != nil {
		// stop last one
		close(_LiveProcessName.stop)
	}

	var (
		cnt  int
		last string
		stop = make(chan struct{})
	)
	go func() {
		for {
			if liveName != nil {
				var name = liveName(utils.MaxProcessNameLength())
				if cnt == 0 || name != last {
					utils.RenameMyProcess(ctx, name)
					cnt++
					last = name
				}
			}

			select {
			case <-stop:
				return
			case <-time.After(interval):
			}
		}
	}()

	// record this one
	_LiveProcessName.stop = stop
	return stop
}
