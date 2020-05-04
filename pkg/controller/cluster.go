package controller

import (
	"encoding/base64"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// Config is synka configuration
type Config struct {
	Clusters []Cluster
}

// Cluster is a Kubernetes cluster to witch synka will post resources to
type Cluster struct {
	Name                  string `yaml:"name,omitempty"`
	Server                string `yaml:"server,omitempty"`
	InsecureSkipTLSVerify bool   `yaml:"insecure-skip-tls-verify,omitempty"`
	Cert                  string `yaml:"cert,omitempty"`
	Key                   string `yaml:"key,omitempty"`
	Ca                    string `yaml:"ca,omitempty"`
	Token                 string `yaml:"token,omitempty"`
	client                dynamic.Interface
}

// GetClient creates and returns a dynamic client that can be used to interact with a cluster
func (c *Cluster) GetClient(gvr *schema.GroupVersionResource) (dynamic.Interface, error) {
	if c.client != nil {
		return c.client, nil
	}
	config := api.NewConfig()
	config.Clusters[c.Name] = api.NewCluster()
	config.Clusters[c.Name].Server = c.Server
	config.Clusters[c.Name].InsecureSkipTLSVerify = c.InsecureSkipTLSVerify
	config.Clusters[c.Name].CertificateAuthorityData = b64ToBytes(c.Ca)
	config.AuthInfos[c.Name] = api.NewAuthInfo()
	config.AuthInfos[c.Name].ClientCertificateData = b64ToBytes(c.Cert)
	config.AuthInfos[c.Name].ClientKeyData = b64ToBytes(c.Key)
	config.AuthInfos[c.Name].Token = c.Token
	config.Contexts[c.Name] = api.NewContext()
	config.Contexts[c.Name].Cluster = c.Name
	config.Contexts[c.Name].AuthInfo = c.Name
	config.CurrentContext = c.Name

	clientconfig, err := clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		return nil, err
	}
	clientconfig.GroupVersion = &v1.SchemeGroupVersion
	clientconfig.NegotiatedSerializer = scheme.Codecs.WithoutConversion()

	client, err := dynamic.NewForConfig(clientconfig)
	if err != nil {
		return nil, err
	}
	c.client = client

	return c.client, nil
}

// Takes a base64 encoded string, decodes it and converts it into a byte array
func b64ToBytes(str string) []byte {
	var b []byte
	data, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return b
	}
	return data
}
