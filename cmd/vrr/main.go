package main

import "github.com/luscis/openvrr/pkg/vrr"

func main() {
	v := vrr.Vrr{}
	v.Init()
	v.Start()
	v.Wait()
}
