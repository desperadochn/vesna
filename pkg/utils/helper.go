package utils

import (
	"context"
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/olekukonko/tablewriter"
	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"log"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"vesna/pkg/cache"
	"strconv"
	v1 "k8s.io/api/core/v1"
)
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

// item is in []string{}
func InArray(arr []string,item string ) bool  {
	for _,p:=range arr{
		if p==item{
			return true
		}
	}
	return false
}
//设置table的样式
func SetTable(table *tablewriter.Table){
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t") // pad with tabs
	table.SetNoWhiteSpace(true)
}
//两个返回值， 一个是 命令 第二个是options
func ParseCmd(w string) (string,string){
	w=regexp.MustCompile("\\s+").ReplaceAllString(w," ")
	l:=strings.Split(w," ")
	if len(l)>=2{
		return l[0],strings.Join(l[1:]," ")
	}
	return w,""
}

func SetNameSpace(args []string,cmd *cobra.Command){
	if len(args)==0{
		log.Println("namespace name is required")
		return
	}
	err:=cmd.Flags().Set("namespace",args[0])
	if err!=nil{
		log.Println("设置namespace失败:",err.Error())
		return
	}
	fmt.Println("设置namespace成功")
	ToCurrentNameSpace(cmd)
}
type CurrentResource struct {
	NameSpace string
}
var CurrentNS string = "default"
func ToCurrentNameSpace(cmd *cobra.Command)  {
	ns := GetNameSpace(cmd)
	//var CurrentResource CurrentResource
	//CurrentResource.NameSpace = ns
	//return CurrentResource.NameSpace
	CurrentResource := &CurrentResource{
		NameSpace: ns,
	}
	CurrentNS = CurrentResource.NameSpace
}
func GetPodsList() (ret []prompt.Suggest) {
/*	cache.InitCache()*/
	if CurrentNS == ""{
		CurrentNS = "default"
		return
	}
	pods,err := cache.Factory.Core().V1().Pods().Lister().Pods(CurrentNS).List(labels.Everything())
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

const DefaultNameSpace="default"
func GetNameSpace(cmd *cobra.Command) string{
	ns,err:=cmd.Flags().GetString("namespace")
	if err!=nil{
		log.Println("error ns param")
		return  DefaultNameSpace}
	if ns==""{ns=DefaultNameSpace}
	return ns
}
func GetNameSpaceList() (ret []prompt.Suggest)  {
	ns,err := cache.Factory.Core().V1().Namespaces().Lister().List(labels.Everything())
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
func ResetSTTY()  {
	cc:=exec.Command("stty", "-F", "/dev/tty", "echo")
	cc.Stdout=os.Stdout
	cc.Stderr=os.Stderr
	if err:=cc.Run();err!=nil{
		log.Println(err)
	}
}

//初始化头
func InitHeader(table *tablewriter.Table) []string  {
	commonHeaders:=[]string{"名称", "命名空间", "IP","状态"}
	if  ShowLabels{
		commonHeaders=append(commonHeaders,"标签")
	}
	return commonHeaders
}
func DeployHeader(table *tablewriter.Table) []string {
	deployHeaders:=[]string{"名称","命名空间", "就绪情况","UP-TO-DATE","AVAILABLE","AGE","最新事件"}
	return deployHeaders

}
var ShowLabels bool


func GetPodMetric(ns string)  {
	mlist,err:=cache.MetricClient.MetricsV1beta1().PodMetricses(ns).
		List(context.Background(),metav1.ListOptions{})
	if err!=nil{
		return
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"名称","cpu/内存"})

	data := [][]string{}
	for _,p:=range mlist.Items{
		for _,c:=range p.Containers{
			podRow:=[]string{}
			if c.Name=="POD"{
				continue
			}
			mem:=c.Usage.Memory().Value()/1024/1024
			podRow=append(podRow,p.Name,
				fmt.Sprintf("%s(%sm/%dM)",c.Name,c.Usage.Cpu().String(),mem))
			data=append(data,podRow)
		}

	}
	table.AppendBulk(data)
	table.SetRowLine(true)
	table.SetAutoMergeCells(true)
	table.Render()
}

func GetNodeMetric()  {
	mlist,err:=cache.MetricClient.MetricsV1beta1().NodeMetricses().
		List(context.Background(),metav1.ListOptions{})
	if err!=nil{
		fmt.Println(err)
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"name","cpu核数","cpu占用","cpu占用百分比","内存占用","内存占用百分比"})
	data := [][]string{}
	for _,p:=range mlist.Items{
		nodeRow:=[]string{}
		name:= p.Name
		cpuUsageInit:=p.Usage.Cpu().AsApproximateFloat64() * 1000
		memUsageInit:=int(p.Usage.Memory().Value()/1024/1024)
		memUsage:=p.Usage.Memory().AsApproximateFloat64()/1024/1024
		cpuUsageString:=strconv.FormatFloat(cpuUsageInit,'f',2,64) + "m"
		memUsageString:=strconv.FormatFloat(memUsage,'f',2,64) + "mb"
		node,err:= cache.Factory.Core().V1().Nodes().Lister().Get(name)
		if err != nil{
			fmt.Println(err)
			fmt.Println("33333")
		}
		cpuCapacity:=node.Status.Capacity.Cpu().AsApproximateFloat64() * 1000
		cpuCapacityString:=node.Status.Capacity.Cpu().String()
		cpuPercentageFloat:= (cpuUsageInit / cpuCapacity)
		cpuPercentage:= strconv.FormatFloat(cpuPercentageFloat,'f',2,64) + "%"
		memoryCapacityInit:= int(node.Status.Capacity.Memory().Value()/1024/1024)
/*		cpuCapacityInit:= int(node.Status.Capacity.Cpu().Value()/1024/1024)*/
		memPercentage:= Percentage(memUsageInit,memoryCapacityInit)
/*		cpuPercentage:= Percentage(cpuUsageInit,cpuCapacityInit)*/
		nodeRow=append(nodeRow,name,cpuCapacityString,cpuUsageString,cpuPercentage,memUsageString,memPercentage,
			)
		data=append(data,nodeRow)
	}
	table.AppendBulk(data)
	table.SetRowLine(true)
	table.SetAutoMergeCells(true)
	table.Render()

}
func Percentage(arg1,arg2 int) string{
	percentagetmp, _ := (decimal.NewFromFloat(float64(arg1)).Div(decimal.NewFromFloat(float64(arg2)))).Float64()
	percentageDiv, _ :=decimal.NewFromFloat(percentagetmp).Round(4).Float64()
	percentageInit:=percentageDiv * 100
	percentage:=(strconv.FormatFloat(percentageInit,'f',2,64)) +"%"
/*	percentaget:=fmt.Sprintf("%f",percentagetInit)*/
    return percentage
}

var eventHeaders=[]string{"事件类型", "REASON", "所属对象","消息"}
func PrintEvent(events []*v1.Event){
	table := tablewriter.NewWriter(os.Stdout)
	//设置头
	table.SetHeader(eventHeaders)
	for _,e:=range events {
		podRow:=[]string{e.Type,e.Reason,
			fmt.Sprintf("%s/%s",e.InvolvedObject.Kind,e.InvolvedObject.Name),e.Message}

		table.Append(podRow)
	}
	SetTable(table)
	table.Render()
}

var podHeaders=[]string{"POD名称","IP","状态","节点"}
func PrintPods(pods []*v1.Pod){
	table := tablewriter.NewWriter(os.Stdout)
	//设置头
	table.SetHeader(podHeaders)
	for _,pod:=range pods {
		podRow:=[]string{pod.Name,pod.Status.PodIP,
			string(pod.Status.Phase),pod.Spec.NodeName}
		table.Append(podRow)
	}
	SetTable(table)
	table.Render()
}

