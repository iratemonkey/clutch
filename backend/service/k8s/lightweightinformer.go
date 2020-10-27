package k8s

import (
	"log"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
)

type lightWeightCacheObject struct {
	metav1.Object
	UID        types.UID
	Name       string
	Namespace  string
	Finalizers []string
	Labels     map[string]string
}

func (cl *lightWeightCacheObject) GetUID() types.UID             { return cl.UID }
func (cl *lightWeightCacheObject) SetUID(uid types.UID)          { cl.UID = uid }
func (cl *lightWeightCacheObject) GetName() string               { return cl.Name }
func (cl *lightWeightCacheObject) SetName(name string)           { cl.Name = name }
func (cl *lightWeightCacheObject) GetNamespace() string          { return cl.Namespace }
func (cl *lightWeightCacheObject) SetNamespace(namespace string) { cl.Namespace = namespace }

func NewLightweightInformer(
	cs ContextClientset,
	lw cache.ListerWatcher,
	objType runtime.Object,
	resync time.Duration,
	h cache.ResourceEventHandler,
) cache.Controller {
	keyFunc := func(obj interface{}) (string, error) {
		theMeta, err := meta.Accessor(obj)
		if err != nil {
			return "", err
		}
		return string(theMeta.GetUID()), nil
	}

	deletehandler := func(obj interface{}) (string, error) {
		if d, ok := obj.(cache.DeletedFinalStateUnknown); ok {
			return d.Key, nil
		}

		return keyFunc(obj)
	}

	cacheStore := cache.NewIndexer(deletehandler, cache.Indexers{})
	fifo := cache.NewDeltaFIFOWithOptions(cache.DeltaFIFOOptions{
		KnownObjects:          cacheStore,
		EmitDeltaTypeReplaced: true,
		KeyFunction:           keyFunc,
	})

	return cache.New(&cache.Config{
		Queue:            fifo,
		ListerWatcher:    lw,
		ObjectType:       objType,
		FullResyncPeriod: resync,
		RetryOnError:     false,
		Process: func(obj interface{}) error {
			for _, d := range obj.(cache.Deltas) {

				incomeingObjectMeta, _ := meta.Accessor(d.Object)
				lightweightObj := &lightWeightCacheObject{}
				lightweightObj.SetUID(incomeingObjectMeta.GetUID())
				lightweightObj.SetName(incomeingObjectMeta.GetName())
				lightweightObj.SetNamespace(incomeingObjectMeta.GetNamespace())

				switch d.Type {
				case cache.Sync, cache.Replaced, cache.Added, cache.Updated:
					if _, exists, err := cacheStore.Get(lightweightObj); err == nil && exists {
						if err := cacheStore.Update(d.Object); err != nil {
							log.Printf("error updating %v", err)
						}
						h.OnUpdate(nil, d.Object)
					} else {
						if err := cacheStore.Add(lightweightObj); err != nil {
							log.Printf("error adding %v", err)
						}
						h.OnAdd(d.Object)
					}
				case cache.Deleted:
					if err := cacheStore.Delete(lightweightObj); err != nil {
						return err
					}
					h.OnDelete(d.Object)
				}
			}
			return nil
		},
	})
}
