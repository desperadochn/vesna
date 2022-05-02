package cmds


import (
	"k8s.io/apimachinery/pkg/types"
	"time"
	"vesna/pkg/cache"
	"vesna/pkg/utils"
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"log"
	"os"
	"sort"
	"strconv"
)
type V1Deployment []*appv1.Deployment
type V1Node []*corev1.Node
func(this V1Deployment) Len() int{
	return len(this)
}
func(this V1Deployment) Less(i, j int) bool{
	//根据时间排序    倒排序
	return this[i].CreationTimestamp.Time.After(this[j].CreationTimestamp.Time)
}
func(this V1Deployment) Swap(i, j int){
	this[i],this[j]=this[j],this[i]
}
func GetDeployList() (ret []prompt.Suggest)  {
	CurrentNS:=utils.CurrentNS
	deploys,err:= cache.Factory.Apps().V1().Deployments().Lister().Deployments(CurrentNS).List(labels.Everything())
	sort.Sort(V1Deployment(deploys))
	if err!=nil{
		return
	}
	for _,depoy := range deploys{
		ret=append(ret,prompt.Suggest{
			Text: depoy.Name,
			Description: "当前副本数:" + strconv.FormatInt(int64(depoy.Status.Replicas),10),
		})
	}
	return
}
//取出deploy列表
func listDeploys(ns string) []*appv1.Deployment {
	cache.InitCache()
	list,err := cache.Factory.Apps().V1().Deployments().Lister().Deployments(ns).
		List(labels.Everything())
	if err != nil{
		log.Println(err)
		return nil
	}
	sort.Sort(V1Deployment(list)) // 排序
	return list
}
//用于提示 用
func RecommendDeployment(ns string) (ret []prompt.Suggest) {
	deplist:= listDeploys(ns)
	if deplist == nil{
		return
	}
	for _,dep:= range deplist{
		ret=append(ret,prompt.Suggest{
			Text:        dep.Name,
			Description: fmt.Sprintf("副本:%d/%d",dep.Status.Replicas,
				dep.Status.Replicas),
		})

	}
	return
}

//渲染 deploys 列表

func RenderDeploy(args []string,cmd *cobra.Command)  {
	deplist:= listDeploys(utils.GetNameSpace(cmd))
	ns:=utils.GetNameSpace(cmd)
	if deplist==nil{
		fmt.Println("nil deplist")
		return
	}
	table := tablewriter.NewWriter(os.Stdout)
	//设置头
	table.SetHeader(utils.DeployHeader(table))
	for _,dep:=range deplist {
		t1 := time.Now()
		sub := t1.Sub(dep.CreationTimestamp.Time)
		h:=sub.Hours()
		depRow:=[]string{dep.Name,dep.Namespace,GetDeployStatus(dep.Namespace,dep.Name),
			strconv.FormatInt(int64(dep.Status.UpdatedReplicas),10),
			strconv.FormatInt(int64(dep.Status.AvailableReplicas),10),
			fmt.Sprintf("%.0fH", h),
			getLatestDeployEvent(dep.UID,ns),
		}

		table.Append(depRow)
	}
	utils.SetTable(table)
	table.Render()
}
func GetDeployStatus(ns,name string) string  {
	deploy,err:= cache.Factory.Apps().V1().Deployments().Lister().Deployments(ns).Get(name)
	if err != nil{

	}
	replicas:= strconv.FormatInt(int64(deploy.Status.Replicas),10)
	readyNum := strconv.FormatInt(int64(deploy.Status.ReadyReplicas),10)
	readyStatus:= fmt.Sprintf("%s/%s",replicas,readyNum)
	return readyStatus
}


func Listpods(ns string) []*corev1.Pod  {
	list,err := cache.Factory.Core().V1().Pods().Lister().Pods(ns).
		List(labels.Everything())
	if err != nil{
		log.Println(err)
		return nil
	}
	sort.Sort(CoreV1Pods(list))
	return list
}
func ListNode() []*corev1.Node {
	list,err := cache.Factory.Core().V1().Nodes().Lister().List(labels.Everything())
	if err!=nil{
		log.Println(err)
		return nil
	}
	sort.Sort(V1Node(list))
	return list
}
func RenderPods(args []string,cmd *cobra.Command)  {
	podlist:= Listpods(utils.GetNameSpace(cmd))
	if podlist == nil{
		return
	}
	table := tablewriter.NewWriter(os.Stdout)
	//设置头
	table.SetHeader(utils.InitHeader(table))
	for _,pod:= range podlist{
		podRow:= []string{pod.Name,pod.Namespace,pod.Status.PodIP,
			string(pod.Status.Phase)}
		table.Append(podRow)
	}
	utils.SetTable(table)
	table.Render()
}
type CoreV1Pods []*corev1.Pod

func (this CoreV1Pods)Len() int {
	return len(this)
}
func (this CoreV1Pods)Less(i, j int) bool {
	//根据时间排序    倒排序
	return this[i].CreationTimestamp.Time.After(this[j].CreationTimestamp.Time)
}

func (this CoreV1Pods)Swap(i, j int)  {
	this[i],this[j]=this[j],this[i]
}

type V1Events []*corev1.Event

func (this V1Events)Len() int {
	return len(this)
}

func (this V1Events)Swap(i, j int)  {
	this[i],this[j]=this[j],this[i]
}

func (this V1Events)Less(i, j int) bool {
	//根据时间排序    倒排序
	return this[i].CreationTimestamp.Time.After(this[j].CreationTimestamp.Time)
}

//获取deployment的最新事件
func getLatestDeployEvent(uid types.UID ,ns string) string   {
	list,err:=cache.Factory.Core().V1().Events().Lister().Events(ns).
		List(labels.Everything())
	if err!=nil{
		return ""
	}
	sort.Sort(V1Events(list)) //排序
	for _,e:=range list{
		if e.InvolvedObject.UID==uid {
			return e.Message
		}
	}
	return ""
}

func (this V1Node)Len() int {
	return len(this)
}
func (this V1Node)Swap(i,j int)   {
	this[i],this[j]=this[j],this[i]
}
func (this V1Node)Less(i,j int) bool  {
	return this[i].CreationTimestamp.Time.After(this[j].CreationTimestamp.Time)
}
func RenderNodes(args []string,cmd *cobra.Command)  {
	nodeList:= ListNode()
	if nodeList==nil{
		return
	}
	table := tablewriter.NewWriter(os.Stdout)
	//设置头
	table.SetHeader(utils.NodeHeader(table))
	for  _,node:= range nodeList{
		t1 := time.Now()
		sub:= t1.Sub(node.CreationTimestamp.Time)
		h:=sub.Hours()
		nodeRow:=[]string{
			node.Name,
			GetNodeConditions(node.Name),
			node.Labels["kubernetes.io/role"],
			fmt.Sprintf("%.0fH", h),
			node.Status.NodeInfo.KubeletVersion,
		}
		table.Append(nodeRow)
	}
	utils.SetTable(table)
	table.Render()

}
func GetNodeConditions(name string) string {
	var Conditions string
	node,err:= cache.Factory.Core().V1().Nodes().Lister().Get(name)
	if err !=nil{
	}
	for _, value := range node.Status.Conditions {
		if value.Type == "Ready"{
			if value.Status == "True"{
				Conditions = "Ready"
			} else {
				Conditions = "No Ready"
				
			}
		}
	}
	return Conditions
}

