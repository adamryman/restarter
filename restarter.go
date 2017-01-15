package restarter

import (
	"os"
	"os/exec"
	"time"

	"context"

	. "github.com/y0ssar1an/q"
)

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
			Q(err)
			time.Sleep(time.Second)
			continue
		}
		Q("waiting for restart")
		<-restart
		Q("RESTART GOT")
		done()
		err = cmd.Wait()
		if err != nil {
			Q(err)
		}
		select {
		case <-ctx.Done():
			break
		default:
		}
	}
}
