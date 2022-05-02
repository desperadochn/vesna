package cmds


import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"regexp"
	"vesna/pkg/utils"
	"vesna/pkg/cache"
/*	"vesna/pkg/kp/lib"*/
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strings"
	"context"
)
var MyConsoleWriter=  prompt.NewStdoutWriter()
func executorCmd(cmd *cobra.Command) func(in string) {
	return func(in string) {
		in = strings.TrimSpace(in)
		blocks := strings.Split(in, " ")
		args := []string{}
		if len(blocks) > 1 {
			args = blocks[1:]
		}
		switch blocks[0] {
		case "exit":
			fmt.Println("Bye!")
			utils.ResetSTTY() //这步要执行 ，否则无法打印命令
			os.Exit(0)
		case "list":
			cache.InitCache()
			RenderPods(args,cmd)
		case "get":
			clearConsole()
			runtea(args, cmd)
		case "clear":
			clearConsole()
		case "del":
			delPod(args, cmd)
		case "ns":
			fmt.Println("您当前所处的namespace是：",utils.GetNameSpace(cmd))
		case "use":
			utils.SetNameSpace(args,cmd)
		case "exec":
	         RunteaExec(args, cmd)
		case "ListNode":
			RenderNodes(args,cmd)
		case "top":
			utils.GetPodMetric(utils.GetNameSpace(cmd))
		case "deploy":
			RenderDeploy(args,cmd)
		case "topNode":
			utils.GetNodeMetric()
		case "scale":
			ScaleDeploy(args,cmd)
		case "getDeploy":
			clearConsole()
			runDeployInfo(args,cmd)
		}
	}
}

func clearConsole()  {
	MyConsoleWriter.EraseScreen()
	MyConsoleWriter.CursorGoTo(0,0)
	MyConsoleWriter.Flush()
}

func parseCmd(w string) (string,string) {
	w=regexp.MustCompile("\\s+").ReplaceAllString(w," ")
	l:=strings.Split(w," ")
	if len(l)>=2{
		return l[0],strings.Join(l[1:]," ")
	}
	return w,""
}
func completer(in prompt.Document) []prompt.Suggest {
	w := in.GetWordBeforeCursor()
	if w == "" {
		return []prompt.Suggest{}
	}
	cmd,opt:= parseCmd(in.TextBeforeCursor())
	if cmd=="get"{
		return prompt.FilterHasPrefix(utils.GetPodsList(),opt, true)
	}
	if cmd=="scale"{
		return prompt.FilterHasPrefix(GetDeployList(),opt,true)
	}
	if cmd=="exec"{
		return prompt.FilterHasPrefix(utils.GetPodsList(),opt, true)
	}
	if cmd=="del"{
		return prompt.FilterHasPrefix(utils.GetPodsList(),opt,true)
	}
	if cmd=="use"{
		return prompt.FilterHasPrefix(utils.GetNameSpaceList(),opt,true)
	}
	if cmd=="getDeploy"{
		return prompt.FilterHasPrefix(GetDeployList(),opt,true)
	}
	return prompt.FilterHasPrefix(suggestions, w, true)
}

var suggestions = []prompt.Suggest{
	// Command
	{"get", "获取POD详细"},
	{"list", "显示Pods列表"},
	{"deploy", "显示deploys列表"},
	{"exit", "退出交互式窗口"},
	{"clear", "清除屏幕"},
	{"use", "设置当前namespace,请填写名称"},
	{"ns", "显示当前命名空间"},
	{"del", "删除某个POD"},
	{"exec", "pod的shell操作"},
	{"top", "显示当前POD列表的指标数据"},
	{"topNode", "显示当前node列表的指标数据"},
	{"scale", "伸缩副本"},
	{"getDeploy", "获取deploy详细"},
	{"ListNode", "list node"},
}
var promptCmd= &cobra.Command{
	Use:          "prompt",
	Short:        "prompt pods ",
	Example:      "kubectl ingress prompt",
	SilenceUsage: true,
	RunE: func(c *cobra.Command, args []string) error {
		p := prompt.New(
			executorCmd(c),
			completer,
			prompt.OptionTitle("欢迎使用vesna，一款kubernetes命令行工具"),
			prompt.OptionPrefix(">>> "),
			prompt.OptionWriter(MyConsoleWriter), //设置自己的writer
		)
		p.Run()
		return nil
	},

}

func delPod(args []string,cmd *cobra.Command)  {
	if len(args)==0{
		log.Println("podname is required")
		return
	}
	ns:=utils.GetNameSpace(cmd)
	err:= Client.CoreV1().Pods(ns).Delete(context.Background(),args[0],metav1.DeleteOptions{})
	if err!=nil{
		log.Println("delete pod error:",err.Error())
		return
	}
	log.Println("删除POD:",args[0],"成功")
}
