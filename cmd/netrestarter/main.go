package cli

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/adamryman/restarter"

	. "github.com/y0ssar1an/q"
)

func main() {
	os.Exit(run())
}

func run() int {
	if len(os.Args) < 2 {
		fmt.Println("YOU MUST CONSTRUCT ADDITIONAL PYLONS")
		return 1
	}
	var args []string
	if len(os.Args) > 2 {
		args = os.Args[2:]
	}

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

	go restarter.DoWithContext(ctx, os.Args[1], args, restart)

	l, err := net.Listen("tcp", "localhost:5040")
	if err != nil {
		fmt.Println(err)
		return 1
	}
	conns := make(chan io.ReadWriteCloser)
	go func() {
		for {
			c, err := l.Accept()
			Q("connection got")
			if err != nil {
				fmt.Println(err)
				continue
			}
			conns <- c
		}
	}()
	for {
		Q("selecting")
		select {
		case <-ctx.Done():
			Q("done")
			return 0
		case c := <-conns:
			Q("sending to restart")
			restart <- true
			Q("restart sent, closeing connection")
			c.Close()
		}
	}
}
