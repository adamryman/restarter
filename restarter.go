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

func DoWithContext(ctx context.Context, name string, args []string, restart <-chan bool) error {
	var cmd *exec.Cmd
	var err error
	errc := make(chan error)

	for {
		select {
		case <-ctx.Done():
			debug("parent done")
			return nil
		default:
		}

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		debug("starting cmd:")
		debug(name)
		cmd = exec.CommandContext(ctx, name, args...)
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
			debug("got error")
			if err != nil {
				debug("got error from cmd")
				debug(err)
				return err
			}
		case <-restart:
			debug("got restart")
			cancel()
			err := <-errc
			if err != nil {
				debug("got error out of errc")
				debug(err)
			}
		case <-ctx.Done():
			debug("ctx Done")
			return nil
		}
	}
}

var debug func(...interface{})

func noop(v ...interface{}) {}

func init() {
	if os.Getenv("RESTARTER_DEBUG") == "" {
		debug = noop
	} else {
		debug = log.New(os.Stderr, "RESTARTER: ", log.LstdFlags).Println
	}
}
