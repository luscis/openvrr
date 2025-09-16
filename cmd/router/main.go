package main

import "github.com/luscis/openvrr/pkg/router"

func main() {
	r := router.Router{}
	r.Init()
	r.Start()
	r.Wait()
}
