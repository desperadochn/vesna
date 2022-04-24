package cache
import (
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/metrics/pkg/client/clientset/versioned"
	"log"
)

//临时放的一个 空的 handler

type DeployHandler struct {
}
func(this *DeployHandler) OnAdd(obj interface{})               {}
func(this *DeployHandler) OnUpdate(oldObj, newObj interface{}) {}
func(this *DeployHandler) OnDelete(obj interface{})            {}

type PodHandler struct {
}
func(this *PodHandler) OnAdd(obj interface{})               {}
func(this *PodHandler) OnUpdate(oldObj, newObj interface{}) {}
func(this *PodHandler) OnDelete(obj interface{})            {}

type EventHandler struct {

}

type NameSpaceHandler struct {

}
type NodeHandler struct {

}
type ReplicaSetsHandler struct {

}

func (this *ReplicaSetsHandler)OnAdd(obj interface{})  {

}
func (this *ReplicaSetsHandler)OnUpdate(oldObj,newObj interface{})  {

}
func (this *ReplicaSetsHandler)OnDelete(obj interface{})  {

}

func (this *NodeHandler)OnAdd(obj interface{})  {

}
func (this *NodeHandler)OnUpdate(oldObj,newObj interface{})  {

}
func (this *NodeHandler)OnDelete(obj interface{})  {

}
func (this *NameSpaceHandler)OnAdd(obj interface{})  {

}
func (this *NameSpaceHandler)OnUpdate(oldObj,newObj interface{})  {

}
func (this *NameSpaceHandler)OnDelete(obj interface{})  {

}

func (this *EventHandler)OnAdd(obj interface{})  {

}
func (this *EventHandler)OnUpdate(oldObj,newObj interface{})  {

}
func (this *EventHandler)OnDelete(obj interface{})  {

}


var  Client= InitClient() //这是 clientset
var  RestConfig *rest.Config
var MetricClient *versioned.Clientset
var Factory informers.SharedInformerFactory
var CfgFlags *genericclioptions.ConfigFlags
func InitClient() *kubernetes.Clientset{
	CfgFlags =genericclioptions.NewConfigFlags(true)
	config,err:= CfgFlags.ToRawKubeConfigLoader().ClientConfig()
	if err!=nil{log.Fatalln(err)}
	c,err:=kubernetes.NewForConfig(config)

	if err!=nil{log.Fatalln(err)}
	RestConfig=config
	MetricClient=versioned.NewForConfigOrDie(config)
	return c
}

func InitCache() {
	Factory =informers.NewSharedInformerFactory(Client,0)
	Factory.Apps().V1().Deployments().Informer().AddEventHandler(&DeployHandler{})
	Factory.Core().V1().Pods().Informer().AddEventHandler(&PodHandler{})
	Factory.Core().V1().Namespaces().Informer().AddEventHandler(&NameSpaceHandler{})
	Factory.Core().V1().Events().Informer().AddEventHandler(&EventHandler{})
	Factory.Core().V1().Nodes().Informer().AddEventHandler(&NodeHandler{})
	Factory.Apps().V1().ReplicaSets().Informer().AddEventHandler(&ReplicaSetsHandler{})
	ch:=make(chan struct{})
	Factory.Start(ch)
	Factory.WaitForCacheSync(ch)
}