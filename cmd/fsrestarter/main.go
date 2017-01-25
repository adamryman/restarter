package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/adamryman/restarter"
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	flag "github.com/spf13/pflag"
)

var (
	cmd = flag.StringP("binary", "b", "run", "filename of binary to watch for updates and restart")
	dir = flag.StringP("directory", "d", "/target", "directory where binary is located")
)

func init() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: fsrestarter [options] [arguments to pass to binary]")
		flag.PrintDefaults()
	}
}

func main() {
	os.Exit(run())
}

func run() int {
	flag.Parse()
	args := flag.Args()

	restart := make(chan bool)

	// Mechanical domain.
	errc := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())

	// Interrupt handler.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		err := fmt.Errorf("%s", <-c)
		cancel()
		errc <- err
	}()

	cmdpwd := filepath.Join(*dir, *cmd)
	abscmd, err := filepath.Abs(cmdpwd)
	if err != nil {
		fmt.Println(errors.Wrapf(err, "cannot open %q at %q", *cmd, *dir))
		return 1
	}

	go restarter.DoWithContext(ctx, abscmd, args, restart)
	w, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println(err)
		cancel()
		return 1
	}
	defer w.Close()
	if err := w.Add(*cmd); err != nil {
		fmt.Println(errors.Wrapf(err, "cannot watch %q at %q", *cmd, *dir))
		cancel()
		return 1
	}

	for {
		select {
		case event := <-w.Events:
			if event.Name == cmdpwd {
				restart <- true
			}
		case _ = <-w.Errors:
			cancel()
		case <-ctx.Done():
			return 0
		}
	}
}
