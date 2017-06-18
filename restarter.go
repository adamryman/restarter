// Package restarter restarts a command with arguments when a restart channel
// is sent to.
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

func DoWithContext(ctx context.Context, name string, args []string, restart <-chan bool) error {
	var cmd *exec.Cmd
	var err error
	errc := make(chan error)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			cctx, done := context.WithCancel(ctx)
			cmd = exec.CommandContext(cctx, name, args...)
			cmd.Env = os.Environ()
			cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
			err = cmd.Start()
			if err != nil {
				debug(err)
				time.Sleep(time.Second)
				continue
			}
			go func() {
				errc <- cmd.Wait()
			}()
			debug("waiting for restart")

			select {
			case err = <-errc:
				return err
			case <-ctx.Done():
				return nil
			case <-restart:
				debug("got restart")
				done()
				err := <-errc
				if err != nil {
					debug(err)
				}
			}
		}
	}
}
