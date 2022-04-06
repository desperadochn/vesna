package lib

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"sigs.k8s.io/yaml"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"log"
	"os"
	"regexp"
	"strings"
	"os/exec"
	"k8s.io/apimachinery/pkg/labels"
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)
var CurrentNS string
type CurrentResource struct {
	NameSpace string
}
func ResetSTTY(){
	cc:=exec.Command("stty", "-F", "/dev/tty", "echo")
	cc.Stdout=os.Stdout
	cc.Stderr=os.Stderr
	if err:=cc.Run();err!=nil{
		log.Println(err)
	}
}

func checkError(msg string,err error)  {
	if err != nil{
		errMsg:=fmt.Sprintf("%s:%s\n",msg,err.Error())
		log.Fatalln(errMsg)
	}
}
//两个返回值， 一个是 命令 第二个是options
func parseCmd(w string) (string,string) {
	w=regexp.MustCompile("\\s+").ReplaceAllString(w," ")
	l:=strings.Split(w," ")
	if len(l)>=2{
		return l[0],strings.Join(l[1:]," ")
	}
	return w,""
}
// item is in []string{}
func inArray(arr []string,item string ) bool  {
	for _,p:=range arr{
		if p==item{
			return true
		}
	}
	return false
}
func Map2String(m map[string]string) (ret string )  {
	for k,v:=range m{
		ret+=fmt.Sprintf("%s=%s\n",k,v)
	}
	return
}
func String2Map(str string) map[string]string {
	m := make(map[string]string )
	list:=strings.Split(str,",")
	for _,s:=range list{
		kvs:=strings.Split(s,"=")
		if len(kvs)==2{
			m[kvs[0]]=kvs[1]
		}
	}
	return m
}
//初始化头
func InitHeader(table *tablewriter.Table) []string  {
	commonHeaders:=[]string{"名称", "命名空间", "IP","状态"}
	if  ShowLabels{
		commonHeaders=append(commonHeaders,"标签")
	}
	return commonHeaders
}
func FilterListByJSON(list  *v1.PodList)    {
	jsonStr,_:=json.Marshal(list)

	podSet:=[]string{} //最终 需要返回的 podname
	isSearch:=false // 是否有搜索过滤条件
	if Search_PodName!=""{
		isSearch=true
		ret:=gjson.Get(string(jsonStr),"items.#.metadata.name")
		for _,pod:=range ret.Array(){
			if m,err:=regexp.MatchString(Search_PodName,pod.String());err==nil && m{
				podSet=append(podSet,pod.String())
			}
		}
	}
	if !isSearch{
		return   //没有设置搜索， 原样返回
	}
	podsList:= []v1.Pod{}
	for _,pod:=range list.Items{
		if inArray(podSet,pod.Name){
			podsList=append(podsList,pod)
		}
	}
	list.Items=podsList

}

func getPodDetail(args []string,cmd* cobra.Command)  {
	if len(args)==0{
		log.Println("podname is required")
		return
	}
	ns,err := cmd.Flags().GetString("namespace")
	if err != nil{
		log.Println("error ns param")
		return
	}
	if ns==""{
		ns="default"
	}
	podName:=args[0]
	pod,err := fact.Core().V1().Pods().Lister().Pods(ns).Get(podName)
	if err != nil{
		log.Println(err)
		return
	}
	b,err := yaml.Marshal(pod)
	if err != nil{
		log.Println(err)
		return
	}
	fmt.Println(string(b))
}

func getPodDetailByJSON(podName,path string,cmd *cobra.Command)  {
/*	ns,err:= cmd.Flags().GetString("namespace")
	if err != nil{
		log.Println("error ns param")
		return
	}*/
	ns:=getNameSpace(cmd)
	if ns == ""{
		ns = "default"
	}
	pod,err := fact.Core().V1().Pods().Lister().Pods(ns).Get(podName)
	if err != nil{
		log.Println(err)
		return
	}
	if path==PodEventType{ //代表 是取 POD事件
		eventList,err:= fact.Core().V1().Events().Lister().List(labels.Everything())
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
		req := client.CoreV1().Pods(ns).GetLogs(podName,&v1.PodLogOptions{})
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

//设置table的样式，不重要 。看看就好
func setTable(table *tablewriter.Table){
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

//事件要显示的头
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
	setTable(table)
	table.Render()
}

const DefaultNameSpace="default"


func getNameSpace(cmd *cobra.Command) string {
	ns,err := cmd.Flags().GetString("namespace")
	if err != nil{
		log.Println("error ns param")
		return DefaultNameSpace
	}
	if ns==""{
		ns=DefaultNameSpace
	}
	return ns

}

func delPod(args []string,cmd *cobra.Command)  {
	if len(args)==0{
		log.Println("podname is required")
		return
	}
	ns:=getNameSpace(cmd)
	err:= client.CoreV1().Pods(ns).Delete(context.Background(),args[0],metav1.DeleteOptions{})
	if err!=nil{
		log.Println("delete pod error:",err.Error())
		return
	}
	log.Println("删除POD:",args[0],"成功")
}

func showNameSpace(cmd *cobra.Command)  {
	ns:=getNameSpace(cmd)
	fmt.Println("您当前所处的namespace是：",ns)
}
func ToCurrentNameSpace(cmd *cobra.Command)  {
	ns := getNameSpace(cmd)
	//var CurrentResource CurrentResource
	//CurrentResource.NameSpace = ns
	//return CurrentResource.NameSpace
	CurrentResource := &CurrentResource{
		NameSpace: ns,
	}
	CurrentNS = CurrentResource.NameSpace
}
func setNameSpace(args []string,cmd *cobra.Command)  {
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

func getPodMetric(ns string)  {
	mlist,err:=metricClient.MetricsV1beta1().PodMetricses(ns).
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
