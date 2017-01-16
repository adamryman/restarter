package restarter

import (
	"log"
	"os"
	"os/exec"
	"time"

	"context"
)

var debug func(...interface{})

func noop(v ...interface{}) {}

func init() {
	if os.Getenv("RELOADER_DEBUG") == "" {
		debug = noop
	} else {
		debug = log.New(os.Stderr, "RESTARTER: ", log.LstdFlags).Println
	}
}

func DoWithContext(ctx context.Context, name string, args []string, restart <-chan bool) {
	var cmd *exec.Cmd
	var err error

	for {
		cctx, done := context.WithCancel(ctx)
		cmd = exec.CommandContext(cctx, name, args...)
		cmd.Env = os.Environ()
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
		err = cmd.Start()
		if err != nil {
			debug(err)
			time.Sleep(time.Second)
			continue
		}
		debug("waiting for restart")
		<-restart
		debug("got restart")
		done()
		err = cmd.Wait()
		if err != nil {
			debug(err)
		}
		select {
		case <-ctx.Done():
			break
		default:
		}
	}
}
