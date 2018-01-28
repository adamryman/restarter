package main

import (
	"fmt"
	//"os"
	"time"

	flag "github.com/spf13/pflag"
)

var (
	interval = flag.DurationP("interval", "n", time.Second, "Time between stdout writes.")
	message  = flag.StringP("message", "m", "Hello, stdout!", "String to write to stdout.")
)

func main() {
	flag.Parse()

	for ticker := time.Tick(*interval); ; <-ticker {
		//fmt.Fprintln(os.Stdout, *message)
		fmt.Println(*message)
	}
}
