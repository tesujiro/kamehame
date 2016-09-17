// Kamehame sends http requests and prints results at each specified URL.
package main

import (
	"flag"
	//"fmt"
	"github.com/tesujiro/kamehame"
	"os"
)

var (
	// sema is a counting semaphore for limiting concurrency in fetch.
	concurrency *int = flag.Int("conc", 1, "semaphore count")
	tps         *int = flag.Int("tps", 60, "target transaction / second")
)

func main() {
	flag.Parse()
	//kamehame.Wave(*concurrency, *tps, os.Stdin)
	kamehame.Wave(*concurrency, *tps, os.Stdin)
}
