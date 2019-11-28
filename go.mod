module github.com/fuku2014/cert-manager-webhook-nifcloud

go 1.13

require (
	github.com/go-acme/lego/v3 v3.2.0
	github.com/jetstack/cert-manager v0.12.0
	github.com/pkg/errors v0.8.1
	k8s.io/apiextensions-apiserver v0.0.0-20191121021419-88daf26ec3b8
	k8s.io/apimachinery v0.0.0-20191123233150-4c4803ed55e3
	k8s.io/client-go v0.0.0-20191121015835-571c0ef67034
	k8s.io/klog v1.0.0
)

replace github.com/miekg/dns => github.com/miekg/dns v0.0.0-20170721150254-0f3adef2e220
