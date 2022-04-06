package lib

import (
	"k8s.io/client-go/informers"
)

type PodHandler struct {
}
type EventHandler struct {

}
type NameSpaceHandler struct {

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
func (this *PodHandler)OnAdd(obj interface{})  {
}
func (this *PodHandler)OnUpdate(oldObj,newObj interface{})  {
}
func (this *PodHandler)OnDelete(obj interface{})  {
}
var fact informers.SharedInformerFactory

func InitCache()  {
	fact= informers.NewSharedInformerFactory(client,0)
	fact.Core().V1().Pods().Informer().AddEventHandler(&PodHandler{})
	fact.Core().V1().Events().Informer().AddEventHandler(&EventHandler{})//为了偷懒
	fact.Core().V1().Namespaces().Informer().AddEventHandler(&NameSpaceHandler{})
	ch:= make(chan struct{})
	fact.Start(ch)
	fact.WaitForCacheSync(ch)
}