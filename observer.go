package main

import (
"fmt"
	"time"
	"net/url"
)

func observer(stop <-chan token, channels...chan *url.URL) {
	for {
		select {
		case <- stop:
			fmt.Println("Observer Stopping...")
			return
		default:
			for i, c := range channels {
				fmt.Println(i, len(c))
			}
			time.Sleep(time.Second * 2)
		}

	}
}

// Starts a goroutine running the Observer
func StartObserver(stop chan token, channels ...chan *url.URL ) {
	go func() {
		observer(stop, channels...)
	}()
}