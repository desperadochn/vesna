package main

import (
	"vesna/pkg/cmds"
	"vesna/pkg/utils"
)

func main()  {
/*	defer lib.ResetSTTY()
	lib.RunCmd()*/

	defer utils.ResetSTTY()
	cmds.RunCmd()
}