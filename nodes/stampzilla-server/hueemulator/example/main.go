package main

import (
	"fmt"

	"github.com/stampzilla/stampzilla-go/nodes/stampzilla-server/hueemulator"
)

func main() {
	//hueemulator.SetLogger(os.Stdout)
	hueemulator.Handle("test1", func(req hueemulator.Request, res *hueemulator.Response) {
		fmt.Println("im handling test from", req.RemoteAddr, req.RequestedOnState)
		res.OnState = req.RequestedOnState
		// res.ErrorState = true //set ErrorState to true to have the echo respond with "unable to reach device"
		return
	})
	hueemulator.Handle("test2", func(req hueemulator.Request, res *hueemulator.Response) {
		fmt.Println("im handling test from", req.RemoteAddr, req.RequestedOnState)
		res.OnState = req.RequestedOnState
		// res.ErrorState = true //set ErrorState to true to have the echo respond with "unable to reach device"
		return
	})

	// it is very important to use a full IP here or the UPNP does not work correctly.
	// one day ill fix this
	panic(hueemulator.ListenAndServe("192.168.13.86:80"))
	//panic(hueemulator.ListenAndServe("192.168.13.86:5000"))
}
