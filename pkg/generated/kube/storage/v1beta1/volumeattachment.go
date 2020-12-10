// Code generated by xns-informer-gen. DO NOT EDIT.

package v1beta1

import (
	xnsinformers "github.com/maistra/xns-informer/pkg/informers"
	v1beta1 "k8s.io/api/storage/v1beta1"
	informers "k8s.io/client-go/informers/storage/v1beta1"
	listers "k8s.io/client-go/listers/storage/v1beta1"
	"k8s.io/client-go/tools/cache"
)

type volumeAttachmentInformer struct {
	informer cache.SharedIndexInformer
}

var _ informers.VolumeAttachmentInformer = &volumeAttachmentInformer{}

func NewVolumeAttachmentInformer(f xnsinformers.SharedInformerFactory) informers.VolumeAttachmentInformer {
	resource := v1beta1.SchemeGroupVersion.WithResource("volumeattachments")
	converter := xnsinformers.NewListWatchConverter(
		f.GetScheme(),
		&v1beta1.VolumeAttachment{},
		&v1beta1.VolumeAttachmentList{},
	)

	informer := f.ForResource(resource, xnsinformers.ResourceOptions{
		ClusterScoped:      true,
		ListWatchConverter: converter,
	})

	return &volumeAttachmentInformer{informer: informer.Informer()}
}

func (i *volumeAttachmentInformer) Informer() cache.SharedIndexInformer {
	return i.informer
}

func (i *volumeAttachmentInformer) Lister() listers.VolumeAttachmentLister {
	return listers.NewVolumeAttachmentLister(i.informer.GetIndexer())
}
