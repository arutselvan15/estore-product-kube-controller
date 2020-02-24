package controllers

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"

	"github.com/arutselvan15/estore-common/clients"
	fakecc "github.com/arutselvan15/estore-common/clients/fake"
	pdtv1 "github.com/arutselvan15/estore-product-kube-client/pkg/apis/estore/v1"
)

func makeProduct(namespace, name, brand string, price float64, categories []string, phase pdtv1.ProductPhase) *pdtv1.Product {
	pdt := &pdtv1.Product{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec:       pdtv1.ProductSpec{Brand: brand, Price: price, Categories: categories},
		Status:     pdtv1.ProductStatus{CurrentStatus: pdtv1.CurrentStatus{Phase: phase}},
	}

	return pdt
}

func makeTestProduct() *pdtv1.Product {
	return makeProduct("testNs", "testPdt", "testBrand", 100, []string{"test"}, pdtv1.ProductAvailable)
}

func TestProcessItem(t *testing.T) {
	type args struct {
		pdt      *pdtv1.Product
		clients  clients.EstoreClientInterface
		recorder record.EventRecorder
	}

	pdt := makeTestProduct()
	pdtDelete := pdt.DeepCopy()

	tt := metav1.Now()
	pdtDelete.ObjectMeta.DeletionTimestamp = &tt

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "success process item update", args: args{pdt: pdt, clients: fakecc.NewEstoreFakeClientForConfig([]runtime.Object{pdt}, nil), recorder: record.NewFakeRecorder(fakeRecorderSize)}, wantErr: false},
		{name: "success process item delete", args: args{pdt: pdtDelete, clients: fakecc.NewEstoreFakeClientForConfig([]runtime.Object{pdt}, nil), recorder: record.NewFakeRecorder(fakeRecorderSize)}, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ProcessItem(tt.args.pdt, tt.args.clients, tt.args.recorder); (err != nil) != tt.wantErr {
				t.Errorf("ProcessItem() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_delete(t *testing.T) {
	type args struct {
		pdtCopy  *pdtv1.Product
		clients  clients.EstoreClientInterface
		recorder record.EventRecorder
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "success product create", args: args{pdtCopy: makeTestProduct(), clients: fakecc.NewEstoreFakeClientForConfig(nil, nil), recorder: record.NewFakeRecorder(fakeRecorderSize)}, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := delete(tt.args.pdtCopy, tt.args.clients, tt.args.recorder); (err != nil) != tt.wantErr {
				t.Errorf("delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_update(t *testing.T) {
	type args struct {
		pdtCopy  *pdtv1.Product
		clients  clients.EstoreClientInterface
		recorder record.EventRecorder
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "success product delete", args: args{pdtCopy: makeTestProduct(), clients: fakecc.NewEstoreFakeClientForConfig(nil, nil), recorder: record.NewFakeRecorder(fakeRecorderSize)}, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := update(tt.args.pdtCopy, tt.args.clients, tt.args.recorder); (err != nil) != tt.wantErr {
				t.Errorf("update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
