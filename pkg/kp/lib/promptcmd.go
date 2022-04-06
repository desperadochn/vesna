package lib

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"log"
	"k8s.io/apimachinery/pkg/labels"
	"github.com/spf13/cobra"
	"os"
	"sort"
	"strings"
	corev1 "k8s.io/api/core/v1"

)
var MyConsoleWriter=  prompt.NewStdoutWriter()  //定义一个自己的writer
func executorCmd(cmd *cobra.Command) func(in string) {
	return func(in string) {
		in = strings.TrimSpace(in)
		blocks := strings.Split(in, " ")
		args:=[]string{}
		if len(blocks)>1{
			args=blocks[1:]
		}
		switch blocks[0] {
		case "exit":
			fmt.Println("Bye!")
			ResetSTTY() //这步要执行 ，否则无法打印命令
			os.Exit(0)
		case "list":
			InitCache()
			err := cacheCmd.RunE(cmd,args)
			if err!=nil{
				log.Fatalln(err)
			}
		case "get":
			clearConsole()
			runtea(args,cmd)
		case "clear":
			clearConsole()
		case "del":
			delPod(args,cmd)
		case "ns":
			showNameSpace(cmd)
		case "use":
			setNameSpace(args,cmd)
		case "exec":
			runteaExec(args,cmd)
		case "top":
			getPodMetric(getNameSpace(cmd))
			}
		}
}
func clearConsole()  {
	MyConsoleWriter.EraseScreen()
	MyConsoleWriter.CursorGoTo(0,0)
	MyConsoleWriter.Flush()
}

var suggestions = []prompt.Suggest{
	// Command
	{"get", "获取POD详细"},
	{"list", "显示Pods列表"},
	{"exit", "退出交互式窗口"},
	{"clear", "清除屏幕"},
	{"use", "设置当前namespace,请填写名称"},
	{"ns", "显示当前命名空间"},
	{"del", "删除某个POD"},
	{"exec", "pod的shell操作"},
	{"top", "显示当前POD列表的指标数据"},
}

type CoreV1POD []*corev1.Pod

func (this CoreV1POD)Len() int {
	return len(this)
}

func (this CoreV1POD)Less(i,j int) bool {
	//根据时间排序    倒排序
	return this[i].CreationTimestamp.Time.After(this[j].CreationTimestamp.Time)
}
func (this CoreV1POD)Swap(i,j int)  {
	this[i],this[j]=this[j],this[i]
}
func getPodsList() (ret []prompt.Suggest) {
	InitCache()
	pods,err := fact.Core().V1().Pods().Lister().Pods(CurrentNS).List(labels.Everything())
	sort.Sort(CoreV1POD(pods))
	if err != nil{
		return
	}

	for _,pod := range pods{
		ret=append(ret,prompt.Suggest{
			Text: pod.Name,
			Description:"节点:"+pod.Spec.NodeName+" 状态:"+
				string(pod.Status.Phase)+" IP:"+pod.Status.PodIP,
		})
	}
	return
}
func getNameSpaceList() (ret []prompt.Suggest)  {
	InitCache()
	ns,err := fact.Core().V1().Namespaces().Lister().List(labels.Everything())
	if err != nil{
		return
	}
	for _,namespace := range ns{
		ret=append(ret,prompt.Suggest{
			Text:        namespace.Name,
			Description: "ns:"+namespace.Name+" 状态:"+
				string(namespace.Status.Phase),
		})
	}
	return

}
func completer(in prompt.Document) []prompt.Suggest {
	w := in.GetWordBeforeCursor()
	if w == "" {
		return []prompt.Suggest{}
	}
	cmd,opt:= parseCmd(in.TextBeforeCursor())
	if cmd=="get"{
		return prompt.FilterHasPrefix(getPodsList(),opt, true)
	}
	if cmd=="exec"{
		return prompt.FilterHasPrefix(getPodsList(),opt, true)
	}

	if cmd=="use"{
		return prompt.FilterHasPrefix(getNameSpaceList(),opt,true)
	}
	return prompt.FilterHasPrefix(suggestions, w, true)
}

var promptCmd= &cobra.Command{
	Use:          "prompt",
	Short:        "prompt pods ",
	Example:      "kubectl pods prompt",
	SilenceUsage: true,
	RunE: func(c *cobra.Command, args []string) error {
		p := prompt.New(
			executorCmd(c),
			completer,
			prompt.OptionPrefix(">>> "),
			prompt.OptionWriter(MyConsoleWriter),//设置自己的writer
		)
		p.Run()
		return nil
	},

}