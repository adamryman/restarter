package main

import (
	"context"
	"fmt"
	"log"
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
	binaryRelPath = flag.StringP("binary", "b", "./run", "Relative or full path to binary to restart")
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

	binaryAbsPath, err := filepath.Abs(*binaryRelPath)
	if err != nil {
		fmt.Println(errors.Wrapf(err, "cannot open %q", *binaryRelPath))
		return 1
	}

	go func() {
		// flag.Args() contains flags to pass to the binary
		errc <- restarter.DoWithContext(ctx, binaryAbsPath, flag.Args(), restartChan)
	}()

	// watch the filesystem
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println(err)
		return 1
	}

	// watch the binary for changes
	if err := watcher.Add(binaryAbsPath); err != nil {
		fmt.Println(errors.Wrapf(err, "cannot watch %q", binaryAbsPath))
		return 1
	}

	for {
		select {
		// If we get an event and it is from the binary path, restart
		case event := <-watcher.Events:
			if event.Name == *binaryRelPath {
				debug("got event with binary name")
				restartChan <- true
			}
			bounceWatcher(binaryAbsPath, watcher)
		case err := <-watcher.Errors:
			debug("watcher error")
			fmt.Println(err)
			cancel()
		case err := <-errc:
			debug("we got an error from restarter")
			fmt.Println(err)
			cancel()
		case <-ctx.Done():
			return 0
		}
	}
}

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
	if os.Getenv("DEBUG") == "" {
		debug = noop
	} else {
		debug = log.New(os.Stderr, "FS_RESTARTER: ", log.LstdFlags).Println
	}
}
