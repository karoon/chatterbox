package main

import "log"

type Driver interface {
	Compile(sources []string)
}

type RedisDriver struct{}

func (r RedisDriver) Compile(sources []string) {}

type MongoDriver struct{}

func (r MongodbDriver) Compile(sources []string) {}

type ModelDriver struct {
	Driver
}

func main() {
	d := ModelDriver{RedisDriver{}}
	log.Println(d)
}
