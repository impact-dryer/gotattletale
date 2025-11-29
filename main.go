package main

import "github.com/impact-dryer/gotattletale/pkg"

func main() {
	dev := pkg.Device{
		Name:        "enp18s0",
		Description: "Ethernet interface",
	}
	dev.Start()
}
