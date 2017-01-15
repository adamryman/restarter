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

	. "github.com/y0ssar1an/q"
)

func main() {
	os.Exit(run())
}

func run() int {
	if len(os.Args) < 3 {
		fmt.Println("Usage:")
		fmt.Println("fsrestarter WATCHDIR BIN [ARGS]...")
		return 1
	}
	var args []string
	if len(os.Args) > 4 {
		args = os.Args[3:]
	}

	dir := os.Args[1]
	cmd := os.Args[2]

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

	cmdpwd := filepath.Join(dir, cmd)
	abscmd, err := filepath.Abs(cmdpwd)
	if err != nil {
		fmt.Println(err)
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
	if err := w.Add(os.Args[1]); err != nil {
		fmt.Println(err)
		cancel()
		return 1
	}

	for {
		select {
		case event := <-w.Events:
			Q(event)
			if event.Name == cmdpwd {
				restart <- true
			}
		case err := <-w.Errors:
			Q(err)
			cancel()
		case <-ctx.Done():
			return 0
		}
	}
}
