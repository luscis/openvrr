package main

import "github.com/luscis/openvrr/pkg/vrr"

func main() {
	v := vrr.Gateway{}
	v.Init()
	v.Start()
	v.Wait()
}
