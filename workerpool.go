package main


type WorkerToken = struct{}

type WorkerPool = struct {
	Pool chan WorkerToken
}