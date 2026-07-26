package main

import (
	"fmt"
	"time"
)

func main() {
	done := make(chan string)
	go func() {
		time.Sleep(1 * time.Second)
		done <- "done"
	}()

	msg := <-done
	fmt.Println(msg)

}
