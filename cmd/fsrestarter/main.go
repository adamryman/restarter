package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/adamryman/restarter"
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	flag "github.com/spf13/pflag"
)

var (
	initCmd = flag.StringP("init", "i", "", `Inital command to run before launching binary. It is executed with '/bin/sh -c "$init"'`)
)

func init() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: fsrestarter [binary] [arguments to pass to binary]")
		flag.PrintDefaults()
	}
}

func main() {
	os.Exit(run())
}

func run() int {
	flag.Parse()
	args := flag.Args()
	debug(`init: sh -c "`, *initCmd, `"`)
	debug("restarter args: ", args)
	if len(args) < 1 {
		flag.Usage()
	}
	binaryRelPath := args[0]
	binaryArgs := args[0:]
	debug(binaryRelPath)
	debug(binaryArgs)

	fmt.Fprintf(os.Stderr, "Running init cmd 'sh -c \"%s\"\n", *initCmd)

	// Execute inital cmd
	t := time.Now()
	cmd := exec.Command("sh", "-c", *initCmd)
	cmd.Env = os.Environ()
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "error running init cmd 'sh -c \"%s\"'", *initCmd))
		return 1
	}

	fmt.Fprintf(os.Stderr, "took: %v\n", time.Since(t))

	// send on restart to restart cmd
	restartChan := make(chan bool)

	// Mechanical domain.
	errc := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Interrupt handler.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		debug("waiting for singal")
		err := fmt.Errorf("%s", <-c)
		debug("about to send signal on error channel")
		errc <- err
		debug("sent signal on error channel")
	}()

	binaryAbsPath, err := filepath.Abs(binaryRelPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "cannot open %q", binaryRelPath))
		return 1
	}

	go func() {
		errc <- restarter.DoWithContext(ctx, binaryAbsPath, binaryArgs, restartChan)
	}()

	// watch the filesystem
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	// watch the binary for changes
	if err := watcher.Add(binaryAbsPath); err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "cannot watch %q", binaryAbsPath))
		return 1
	}

	for {
		select {
		// If we get an event and it is from the binary path, restart
		case event := <-watcher.Events:
			if event.Name == binaryRelPath {
				fmt.Fprintf(os.Stderr, "restarting '%s'", args)
				restartChan <- true
			}
			err := bounceWatcher(binaryAbsPath, watcher)
			if err != nil {
				debug(err)
			}
		case err := <-watcher.Errors:
			debug("watcher error")
			debug(err)
			cancel()
		case err := <-errc:
			debug("we got an error from restarter")
			debug(err)
			cancel()
		case <-ctx.Done():
			return 0
		}
	}
}

// bounceWatcher removes and adds a file from the watch list.
// When a binary is overwritten by `go build`, it can be lost. If we bounce
// every time then we will never lose it.
func bounceWatcher(name string, w *fsnotify.Watcher) error {
	err := w.Remove(name)
	if err != nil {
		debug(err)
	}
	return w.Add(name)
}

var debug func(...interface{})

func noop(v ...interface{}) {}

func init() {
	if os.Getenv("RESTARTER_DEBUG") == "" {
		debug = noop
	} else {
		debug = log.New(os.Stderr, "FS_RESTARTER: ", log.LstdFlags).Println
	}
}
