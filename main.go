package main

import (
	"sync"
)

var wg sync.WaitGroup

func main() {
	wg.Add(2)
	go startSshServer()
	go startWebServer()
	wg.Wait()
}
