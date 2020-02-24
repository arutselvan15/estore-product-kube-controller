package controllers

import (
	"testing"

	"k8s.io/client-go/tools/record"

	pdtv1 "github.com/arutselvan15/estore-product-kube-client/pkg/apis/estore/v1"
)

func Test_onAdd(t *testing.T) {
	type args struct {
		pdt      *pdtv1.Product
		recorder record.EventRecorder
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "success onAdd already processed", args: args{pdt: makeProduct("testNs", "testPdt", "testBrand", 100, []string{"test"}, pdtv1.ProductAvailable), recorder: record.NewFakeRecorder(fakeRecorderSize)}, want: ""},
		{name: "success onAdd return key", args: args{pdt: makeProduct("testNs", "testPdt", "testBrand", 100, []string{"test"}, pdtv1.ProductUnknown), recorder: record.NewFakeRecorder(fakeRecorderSize)}, want: "testNs/testPdt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := onAdd(tt.args.pdt, tt.args.recorder); got != tt.want {
				t.Errorf("onAdd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_onDelete(t *testing.T) {
	type args struct {
		pdt      *pdtv1.Product
		recorder record.EventRecorder
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "success onDelete already processed", args: args{pdt: makeProduct("testNs", "testPdt", "testBrand", 100, []string{"test"}, pdtv1.ProductAvailable), recorder: record.NewFakeRecorder(fakeRecorderSize)}, want: "testNs/testPdt"},
		{name: "success onDelete return key", args: args{pdt: makeProduct("testNs", "testPdt", "testBrand", 100, []string{"test"}, pdtv1.ProductUnknown), recorder: record.NewFakeRecorder(fakeRecorderSize)}, want: "testNs/testPdt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := onDelete(tt.args.pdt, tt.args.recorder); got != tt.want {
				t.Errorf("onDelete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_onUpdate(t *testing.T) {
	type args struct {
		oldPdt   *pdtv1.Product
		pdt      *pdtv1.Product
		recorder record.EventRecorder
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "success onUpdate already processed", args: args{pdt: makeProduct("testNs", "testPdt", "testBrand", 100, []string{"test"}, pdtv1.ProductAvailable), recorder: record.NewFakeRecorder(fakeRecorderSize)}, want: ""},
		{name: "success onUpdate return key", args: args{pdt: makeProduct("testNs", "testPdt", "testBrand", 100, []string{"test"}, pdtv1.ProductUnknown), recorder: record.NewFakeRecorder(fakeRecorderSize)}, want: "testNs/testPdt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := onUpdate(tt.args.oldPdt, tt.args.pdt, tt.args.recorder); got != tt.want {
				t.Errorf("onUpdate() = %v, want %v", got, tt.want)
			}
		})
	}
}
