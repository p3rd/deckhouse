# Copyright 2021 Flant JSC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

{{- $experimentalOption := "--experimental-patches" -}}
{{- if semverCompare "<1.19" .kubernetesVersion -}}
  {{- $experimentalOption = "--experimental-kustomize" -}}
{{- end }}

kubeadm init phase certs all --config /var/lib/bashible/kubeadm/config.yaml
kubeadm init phase kubeconfig all --config /var/lib/bashible/kubeadm/config.yaml
kubeadm init phase etcd local --config /var/lib/bashible/kubeadm/config.yaml {{ $experimentalOption }} /var/lib/bashible/kubeadm/patches
kubeadm init phase control-plane all --config /var/lib/bashible/kubeadm/config.yaml {{ $experimentalOption }} /var/lib/bashible/kubeadm/patches
kubeadm init phase mark-control-plane --config /var/lib/bashible/kubeadm/config.yaml

# Upload pki for deckhouse
bb-kubectl --kubeconfig=/etc/kubernetes/admin.conf -n kube-system delete secret d8-pki || true
bb-kubectl --kubeconfig=/etc/kubernetes/admin.conf -n kube-system create secret generic d8-pki \
  --from-file=ca.crt=/etc/kubernetes/pki/ca.crt \
  --from-file=ca.key=/etc/kubernetes/pki/ca.key \
  --from-file=sa.pub=/etc/kubernetes/pki/sa.pub \
  --from-file=sa.key=/etc/kubernetes/pki/sa.key \
  --from-file=front-proxy-ca.crt=/etc/kubernetes/pki/front-proxy-ca.crt \
  --from-file=front-proxy-ca.key=/etc/kubernetes/pki/front-proxy-ca.key \
  --from-file=etcd-ca.crt=/etc/kubernetes/pki/etcd/ca.crt \
  --from-file=etcd-ca.key=/etc/kubernetes/pki/etcd/ca.key

# Setup kubectl for root user
if [ ! -f /root/.kube/config ]; then
  mkdir -p /root/.kube
  ln -s /etc/kubernetes/admin.conf /root/.kube/config
fi
