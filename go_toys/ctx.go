package main

import (
	"context"
	"log"
	"time"
	"reflect"
)

func Opt() {
	bgCtx := context.Background()
	ctx, cancel := context.WithCancel(bgCtx)

	go DoOpt(ctx)

	ctxTo, cancelTo := context.WithTimeout(bgCtx, 2*time.Second)
	go DoOpt(ctxTo)
	cancelTo()

	time.Sleep(5 * time.Second)
	cancel()

	time.Sleep(time.Second)
	ch := ctx.Done()
	_, ok := <-ch
	log.Println(ok)
}

func DoOpt(ctx context.Context) {
	t := reflect.TypeOf(ctx)
	for {
		time.Sleep(time.Second)
		ch := ctx.Done()
		select {
		case <-ch:
			log.Println("done ", t)
			return
		default:
			log.Println("work ", t)
		}
	}
}

func main() {
	Opt()
	log.Println("down ")
}
