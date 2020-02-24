package controllers

import (
	"fmt"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	"github.com/arutselvan15/estore-common/clients"
	fakecc "github.com/arutselvan15/estore-common/clients/fake"
	pdtInformers "github.com/arutselvan15/estore-product-kube-client/pkg/client/informers/externalversions"
	v1 "github.com/arutselvan15/estore-product-kube-client/pkg/client/informers/externalversions/estore/v1"

	cfg "github.com/arutselvan15/estore-product-kube-controller/config"
)

const (
	fakeRecorderSize = 10
)

func TestNewController(t *testing.T) {
	type args struct {
		pdtInformer v1.ProductInformer
		pdtQueue    workqueue.RateLimitingInterface
		clients     clients.EstoreClientInterface
		recorder    record.EventRecorder
		processItem ProcessItemType
	}

	fakeClients := fakecc.NewEstoreFakeClientForConfig(nil, nil)
	pdtInformerFactory := pdtInformers.NewSharedInformerFactory(fakeClients.GetProductClient(), cfg.ResyncDuration)
	pdtInformer := pdtInformerFactory.Estore().V1().Products()
	pdtQueue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "test")
	recorder := record.NewFakeRecorder(fakeRecorderSize)

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "success new controller", args: args{pdtInformer: pdtInformer, pdtQueue: pdtQueue, clients: fakeClients, recorder: recorder, processItem: ProcessItem}, want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewController(tt.args.pdtInformer, tt.args.pdtQueue, tt.args.clients, tt.args.recorder, tt.args.processItem); (got != nil) != tt.want {
				t.Errorf("NewController() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestController_doSync(t *testing.T) {
	type fields struct {
		pdtInformer     cache.SharedIndexInformer
		pdtListerSynced cache.InformerSynced
		pdtQueue        workqueue.RateLimitingInterface
		clients         clients.EstoreClientInterface
		recorder        record.EventRecorder
		processItem     ProcessItemType
	}

	type args struct {
		key string
	}

	pdt := makeTestProduct()

	fakeClients := fakecc.NewEstoreFakeClientForConfig([]runtime.Object{pdt}, nil)
	pdtInformerFactory := pdtInformers.NewSharedInformerFactory(fakeClients.GetProductClient(), cfg.ResyncDuration)
	pdtInformer := pdtInformerFactory.Estore().V1().Products()
	pdtQueue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "test")
	recorder := record.NewFakeRecorder(fakeRecorderSize)

	// add item to product informer index
	_ = pdtInformer.Informer().GetIndexer().Add(pdt)

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "success do sync", args: args{key: fmt.Sprintf("%s/%s", pdt.Namespace, pdt.Name)}, fields: fields{pdtInformer: pdtInformer.Informer(), pdtQueue: pdtQueue, clients: fakeClients, recorder: recorder, processItem: ProcessItem}, wantErr: false,
		},
		{
			name: "failure do sync key not found in store", args: args{key: "unknown-key"}, fields: fields{pdtInformer: pdtInformer.Informer(), pdtQueue: pdtQueue, clients: fakeClients, recorder: recorder, processItem: ProcessItem}, wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Controller{
				pdtInformer:     tt.fields.pdtInformer,
				pdtListerSynced: tt.fields.pdtListerSynced,
				pdtQueue:        tt.fields.pdtQueue,
				clients:         tt.fields.clients,
				recorder:        tt.fields.recorder,
				processItem:     tt.fields.processItem,
			}
			if err := c.doSync(tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("doSync() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestController_processNextItem(t *testing.T) {
	type fields struct {
		pdtInformer     cache.SharedIndexInformer
		pdtListerSynced cache.InformerSynced
		pdtQueue        workqueue.RateLimitingInterface
		clients         clients.EstoreClientInterface
		recorder        record.EventRecorder
		processItem     ProcessItemType
	}

	fakeClients := fakecc.NewEstoreFakeClientForConfig(nil, nil)
	pdtInformerFactory := pdtInformers.NewSharedInformerFactory(fakeClients.GetProductClient(), cfg.ResyncDuration)
	pdtInformer := pdtInformerFactory.Estore().V1().Products()
	pdtQueue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "test")
	recorder := record.NewFakeRecorder(fakeRecorderSize)

	// add item to product informer index
	pdt := makeTestProduct()
	_ = pdtInformer.Informer().GetIndexer().Add(pdt)
	// add key to the queue
	pdtQueue.Add(fmt.Sprintf("%s/%s", pdt.Namespace, pdt.Name))

	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "success process next item", fields: fields{pdtInformer: pdtInformer.Informer(), pdtQueue: pdtQueue, clients: fakeClients, recorder: recorder, processItem: ProcessItem}, want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Controller{
				pdtInformer:     tt.fields.pdtInformer,
				pdtListerSynced: tt.fields.pdtListerSynced,
				pdtQueue:        tt.fields.pdtQueue,
				clients:         tt.fields.clients,
				recorder:        tt.fields.recorder,
				processItem:     tt.fields.processItem,
			}
			if got := c.processNextItem(); got != tt.want {
				t.Errorf("processNextItem() = %v, want %v", got, tt.want)
			}
		})
	}
}
