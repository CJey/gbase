package utils

import (
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unsafe"

	"github.com/cjey/gbase/context"
)

var _MaxProcessNameLength = 0

func init() {
	for i := range os.Args {
		_MaxProcessNameLength += len(os.Args[i])
	}
}

// WriteMyPID write current process pid to the given file.
// If write failed, but the file content is equal to current process id,
// that's also ok.
func WritePID(ctx context.Context, file string, pid int) error {
	var pidstr = strconv.Itoa(pid)
	var werr = ioutil.WriteFile(file, []byte(pidstr), 0644)
	if werr != nil {
		// write failed, let's try read it,
		// maybe process manager(systemd, supervisor, etc.) written for me
		if oldpid, rerr := ioutil.ReadFile(file); rerr == nil {
			if pidstr == strings.TrimSpace(string(oldpid)) {
				// Yes! Lucky~
				return nil
			}
		}
		ctx.Warn("Unavailable to write pid", "err", werr, "file", file, "pid", pid)
	}
	return werr
}

// MaxProcessNameLength returns sum of all string length of elements in os.Args
func MaxProcessNameLength() int {
	// calculated at init()
	return _MaxProcessNameLength
}

// RenameMyProcess overwrite all args in os.Args,
// the maxium length of new name is equal to length sum of all os.Args.
// Because of space split between args, the real display mostly like this:
// e.g. this is a demo. =display=> this i s a  demo.
// You could define a customied encode/decode rules to avoid the problem:
// e.g. this is a demo. =encode=> this+is+a+demo. =display=> this+i s+a+ demo.
func RenameMyProcess(ctx context.Context, name string) {
	var done bool
	defer func() {
		if !done {
			var p = recover()
			ctx.Warn("Panic occured while renaming process", "panic", p)
		}
	}()

	if len(name) < _MaxProcessNameLength {
		name += strings.Repeat(" ", _MaxProcessNameLength-len(name))
	}

	var start, end int
	for i := range os.Args {
		var argstr = (*reflect.StringHeader)(unsafe.Pointer(&os.Args[i]))
		var argv = (*[1 << 30]byte)(unsafe.Pointer(argstr.Data))[:argstr.Len]

		end = start + len(argv)
		copy(argv, name[start:end])
		start = end
	}
	done = true
}
