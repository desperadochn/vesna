package cmds


import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"vesna/pkg/cache"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"log"
	"os"
	"vesna/pkg/utils"
	"context"
	"sigs.k8s.io/yaml"
	"k8s.io/apimachinery/pkg/util/json"
)

type podjson struct {
	title string
	path string
}

type podmodel struct {
	items []*podjson
	index int
	cmd *cobra.Command
	podName string

}

func (m podmodel)Init() tea.Cmd {
	return nil
}

func (m podmodel)Update(msg tea.Msg) (tea.Model,tea.Cmd)  {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up":
			if m.index > 0 {
				m.index--
			}
		case "down":
			if m.index < len(m.items)-1 {
				m.index++
			}
		case "enter" :
			fmt.Println(m.podName)
			getPodDetailByJSON(m.podName,m.items[m.index].path,m.cmd)
			return m,tea.Quit
			
			
		}
	}
	return m,nil

}
var eventHeaders=[]string{"事件类型", "REASON", "所属对象","消息"}

func printEvent(events []*v1.Event)  {
	table := tablewriter.NewWriter(os.Stdout)
	//设置头
	table.SetHeader(eventHeaders)
	for _,e:=range events {
		podRow:=[]string{e.Type,e.Reason,
			fmt.Sprintf("%s/%s",e.InvolvedObject.Kind,e.InvolvedObject.Name),e.Message}

		table.Append(podRow)
	}
	utils.SetTable(table)
	table.Render()
}
func getPodDetailByJSON(podName,path string,cmd *cobra.Command)  {
	/*	ns,err:= cmd.Flags().GetString("namespace")
		if err != nil{
			log.Println("error ns param")
			return
		}*/
	ns:=utils.GetNameSpace(cmd)
	if ns == ""{
		ns = "default"
	}
	pod,err := cache.Factory.Core().V1().Pods().Lister().Pods(ns).Get(podName)
	if err != nil{
		log.Println(err)
		return
	}
	if path==PodEventType{ //代表 是取 POD事件
		eventList,err:= cache.Factory.Core().V1().Events().Lister().Events(ns).List(labels.Everything())
		if err != nil{
			log.Println(err)
			return
		}
		podEvents:=[]*v1.Event{}
		for _,e:= range eventList{
			if e.InvolvedObject.UID == pod.UID{
				podEvents=append(podEvents,e)
			}
		}
		printEvent(podEvents)
		return
	}
	if  path==PodLogType{
		req := Client.CoreV1().Pods(ns).GetLogs(podName,&v1.PodLogOptions{})
		ret := req.Do(context.Background())
		b,err:= ret.Raw()
		if err != nil{
			log.Println(err)
			return
		}
		fmt.Println(string(b))
		return
	}
	jsonStr,_:=json.Marshal(pod)
	ret:=gjson.Get(string(jsonStr),path)
	/*	container := gjson.Get(string(jsonStr),"spec.containers")
		fmt.Println(container)*/
	if !ret.Exists(){
		log.Println("无法找到对应的内容:"+path)
		return
	}
	if !ret.IsObject() && !ret.IsArray(){ //不是对象不是 数组，直接打印
		fmt.Println(ret.Raw)
		return
	}
	/*	tempMap:=make(map[string]interface{})
		xxx:=make([]map[string]interface{},10)*/
	tempMap:=make([]map[string]interface{},10)
	tempMap1:=make(map[string]interface{})
	err=yaml.Unmarshal([]byte(ret.Raw),&tempMap1)
	if err!=nil{
		yaml.Unmarshal([]byte(ret.Raw),&tempMap)
		b,_:=yaml.Marshal(tempMap)
		fmt.Println(string(b))
	} else  {
		b,_:=yaml.Marshal(tempMap1)
		fmt.Println(string(b))
	}
	/*	b,err:=yaml.Marshal(tempMap1)
		fmt.Println(string(b))*/

}

func (m podmodel)View() string  {
	s := "按上下键选择要查看的内容\n\n"
	for i, item := range m.items {
		selected := " "
		if m.index == i {
			selected = "»"
		}
		s += fmt.Sprintf("%s %s\n", selected, item.title)
	}

	s += "\n按Q退出\n"
	return s
}

const (
	PodEventType = "__event__"
	PodLogType= "__log__"
)

func runtea(args []string,cmd *cobra.Command)  {
	cache.InitClient()
	if len(args)==0{
		log.Println("podname is required")
		return
	}
	var podModel=podmodel{
		items:   []*podjson{},
		cmd:     cmd,
		podName: args[0],
	}
	//v1.Pod{}
	podModel.items=append(podModel.items,
		&podjson{title:"元信息", path: "metadata"},
		&podjson{title:"标签", path: "metadata.labels"},
		&podjson{title:"注解", path: "metadata.annotations"},
		&podjson{title:"容器列表", path: "spec.containers"},
		&podjson{title:"全部", path: "@this"},
		&podjson{title:"*事件*", path: PodEventType},
		&podjson{title: "*日志*",path: PodLogType},
		)
	teaCmd := tea.NewProgram(podModel)
	if err := teaCmd.Start(); err != nil {
		fmt.Println("start failed:", err)
		os.Exit(1)
	}

}
