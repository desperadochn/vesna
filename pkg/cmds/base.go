package cmds

import (
	"vesna/pkg/cache"
	"github.com/spf13/cobra"
	"log"
)
func MergeFlags(cmds ...*cobra.Command){
	for _,cmd:=range cmds{
		cache.CfgFlags.AddFlags(cmd.Flags())
	}
}
func RunCmd( ) {
	cmd := &cobra.Command{
		Use:          "kubectl ingress prompt",
		Short:        "list ingress ",
		Example:      "kubectl ingress prompt",
		SilenceUsage: true,
	}
	cache.InitClient() //初始化k8s client
	cache.InitCache() //初始化本地 缓存---informer
	//合并主命令的参数
	MergeFlags(cmd, promptCmd)
	//加入子命令
	cmd.AddCommand( promptCmd)
	err:=cmd.Execute()
	if err!=nil{
		log.Fatalln(err)
	}
}


