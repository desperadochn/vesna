package main

import (
	"vesna/pkg/kp/lib"
)

func main()  {
	defer lib.ResetSTTY()
	lib.RunCmd()
}