package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-acme/lego/v3/providers/dns/nifcloud"
	"github.com/jetstack/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/jetstack/cert-manager/pkg/acme/webhook/cmd"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	"github.com/pkg/errors"
	extapi "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

const ProviderName = "nifcloud"

var GroupName = os.Getenv("GROUP_NAME")

func main() {
	if GroupName == "" {
		klog.Fatal("GROUP_NAME must be specified")
	}

	// Start webhook server
	cmd.RunWebhookServer(GroupName, NewSolver())
}

// Config is a structure that is used to decode into when
// solving a DNS01 challenge.
type Config struct {
	AccessKeySecretRef cmmeta.SecretKeySelector `json:"accessKeySecretRef"`
	SecretKeySecretRef cmmeta.SecretKeySelector `json:"secretKeySecretRef"`
}

// solver implements webhook.Solver
// and will allow cert-manager to create & delete
// DNS TXT records for the DNS01 Challenge
type solver struct {
	client *kubernetes.Clientset
}

// NewSolver returns Solver
func NewSolver() *solver {
	return &solver{}
}

// Name is used as the name for this DNS solver when referencing it on the ACME
// Issuer resource.
func (s solver) Name() string {
	return ProviderName
}

// Present is create TXT DNS record for DNS01
func (s solver) Present(ch *v1alpha1.ChallengeRequest) error {
	klog.Infof("Presenting txt record: %v %v", ch.ResolvedFQDN, ch.ResolvedZone)

	provider, err := s.newDNSProvider(ch)
	if err != nil {
		klog.Errorf("New DNSProvider from challenge error: %v", err)
		return err
	}

	if err := provider.Present(ch.ResolvedFQDN, "", ch.Key); err != nil {
		klog.Errorf("Add txt record %q error: %v", ch.ResolvedFQDN, err)
		return err
	}

	klog.Infof("Presented txt record %v", ch.ResolvedFQDN)
	return nil
}

// Delete TXT DNS record for DNS01
func (s solver) CleanUp(ch *v1alpha1.ChallengeRequest) error {
	klog.Infof("Cleaning up txt record: %v %v", ch.ResolvedFQDN, ch.ResolvedZone)

	provider, err := s.newDNSProvider(ch)
	if err != nil {
		klog.Errorf("New DNSProvider from challenge error: %v", err)
		return err
	}

	if err := provider.CleanUp(ch.ResolvedFQDN, "", ch.Key); err != nil {
		klog.Errorf("Add txt record %q error: %v", ch.ResolvedFQDN, err)
		return err
	}

	klog.Infof("Cleaned up txt record: %v %v", ch.ResolvedFQDN, ch.ResolvedZone)
	return nil
}

// Initialize will be called when the webhook first starts.
// This method can be used to instantiate the webhook, i.e. initialising
// connections or warming up caches.
//
// Typically, the kubeClientConfig parameter is used to build a Kubernetes
// client that can be used to fetch resources from the Kubernetes API, e.g.
// Secret resources containing credentials used to authenticate with DNS
// provider accounts.
//
// The stopCh can be used to handle early termination of the webhook, in cases
// where a SIGTERM or similar signal is sent to the webhook process.
func (s solver) Initialize(kubeClientConfig *rest.Config, stopCh <-chan struct{}) error {
	cl, err := kubernetes.NewForConfig(kubeClientConfig)
	if err != nil {
		return err
	}

	s.client = cl
	return nil
}

func (s *solver) newDNSProvider(ch *v1alpha1.ChallengeRequest) (*nifcloud.DNSProvider, error) {
	cfg, err := loadConfig(ch.Config)
	if err != nil {
		return nil, err
	}

	klog.Infof("Decoded config: %v", cfg)

	accessKey, err := s.getSecretData(cfg.AccessKeySecretRef, ch.ResourceNamespace)
	if err != nil {
		return nil, err
	}

	secretKey, err := s.getSecretData(cfg.SecretKeySecretRef, ch.ResourceNamespace)
	if err != nil {
		return nil, err
	}

	config := nifcloud.NewDefaultConfig()
	config.AccessKey = string(accessKey)
	config.SecretKey = string(secretKey)

	return nifcloud.NewDNSProviderConfig(config)
}

func (s *solver) getSecretData(selector cmmeta.SecretKeySelector, ns string) ([]byte, error) {
	secret, err := s.client.CoreV1().Secrets(ns).Get(selector.Name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load secret %q", ns+"/"+selector.Name)
	}

	if data, ok := secret.Data[selector.Key]; ok {
		return data, nil
	}

	return nil, errors.Errorf("no key %q in secret %q", selector.Key, ns+"/"+selector.Name)
}

// loadConfig is a small helper function that decodes JSON configuration into
// the typed config struct.
func loadConfig(cfgJSON *extapi.JSON) (Config, error) {
	cfg := Config{}

	// handle the 'base case' where no configuration has been provided
	if cfgJSON == nil {
		return cfg, nil
	}

	if err := json.Unmarshal(cfgJSON.Raw, &cfg); err != nil {
		return cfg, fmt.Errorf("error decoding solver config: %v", err)
	}

	return cfg, nil
}
