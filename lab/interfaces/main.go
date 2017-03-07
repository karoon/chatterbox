package main

import "fmt"

type Driver interface {
	Compile(sources []string)
	Register()
	Save()
}

type Redis struct{}

func (gl Redis) Compile(sources []string) {
	fmt.Println("Redis compiling ...")
}
func (gl Redis) Register() {
	fmt.Println("Redis compiling ...")
}
func (gl Redis) Save() {
	fmt.Println("Redis compiling ...")
}

type MongodbDriver struct{}

func (cpp MongodbDriver) Compile(sources []string) {
	fmt.Println("C++ compiling ...")
}

func (cpp MongodbDriver) Register() {
	fmt.Println("C++ compiling ...")
}
func (c MongodbDriver) Save() {
	fmt.Println("C++ compiling ...")
}

type ModelDriver struct {
	Driver
}

func main() {
	var c Driver

	c = ModelDriver{Redis{}}
	c.Compile(nil) // Redis compiling ...
	c.Register()
	c = ModelDriver{MongodbDriver{}}
	c.Compile(nil) // C++ compiling ...
	c = ModelDriver{nil}
	c.Compile(nil) // will panic
}
