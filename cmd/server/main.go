package main

import (
	"core/app"
	"core/router"
	"fmt"
)

func main() {
	defer router.Recover()
	fmt.Println("________Start API________")
	app.Setup()
	router.Setup()
}
