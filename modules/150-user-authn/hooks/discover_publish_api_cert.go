/*
Copyright 2021 Flant JSC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package hooks

import (
	"fmt"

	"github.com/flant/addon-operator/pkg/module_manager/go_hook"
	"github.com/flant/addon-operator/sdk"
	"github.com/flant/shell-operator/pkg/kube_events_manager/types"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/deckhouse/deckhouse/go_lib/module"
)

type PublishAPICert struct {
	Name string `json:"name"`
	Data []byte `json:"data"`
}

func applyPublishAPICertFilter(obj *unstructured.Unstructured) (go_hook.FilterResult, error) {
	s := &v1.Secret{}
	err := sdk.FromUnstructured(obj, s)
	if err != nil {
		return nil, fmt.Errorf("cannot convert kubernetes secret to secret: %v", err)
	}

	return PublishAPICert{Name: obj.GetName(), Data: s.Data["ca.crt"]}, nil
}

var _ = sdk.RegisterFunc(&go_hook.HookConfig{
	OnBeforeHelm: &go_hook.OrderedConfig{Order: 10},
	Kubernetes: []go_hook.KubernetesConfig{
		{
			Name:       "secret",
			ApiVersion: "v1",
			Kind:       "Secret",
			NamespaceSelector: &types.NamespaceSelector{
				NameSelector: &types.NameSelector{
					MatchNames: []string{"d8-user-authn"},
				},
			},
			NameSelector: &types.NameSelector{
				MatchNames: []string{
					"kubernetes-tls",
					"kubernetes-tls-selfsigned",
					"kubernetes-tls-customcertificate",
				},
			},
			FilterFunc: applyPublishAPICertFilter,
		},
	},
}, discoverPublishAPICA)

func discoverPublishAPICA(input *go_hook.HookInput) error {
	var (
		secretPath = "userAuthn.internal.publishedAPIKubeconfigGeneratorMasterCA"
		modePath   = "userAuthn.publishAPI.https.mode"
	)

	caCertificates := make(map[string][]byte)
	for _, s := range input.Snapshots["secret"] {
		publishCert := s.(PublishAPICert)
		caCertificates[publishCert.Name] = publishCert.Data
	}

	var cert []byte

	switch input.Values.Get(modePath).String() {
	case "Global":
		switch module.GetHTTPSMode("userAuthn", input) {
		case "CertManager":
			cert = caCertificates["kubernetes-tls"]
		case "CustomCertificate":
			cert = caCertificates["kubernetes-tls-customcertificate"]
		case "OnlyInURI", "Disabled":
		}
	case "SelfSigned":
		cert = caCertificates["kubernetes-tls-selfsigned"]
	}

	if len(cert) > 0 {
		input.Values.Set(secretPath, string(cert))
	} else {
		input.Values.Remove(secretPath)
	}

	return nil
}
