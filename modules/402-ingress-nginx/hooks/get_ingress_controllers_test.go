package hooks

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/deckhouse/deckhouse/testing/hooks"
)

var _ = Describe("ingress-nginx :: hooks :: get_ingress_controllers ::", func() {
	f := HookExecutionConfigInit(`{"ingressNginx":{"defaultControllerVersion": 0.25, "internal": {}}}`, "")
	f.RegisterCRD("deckhouse.io", "v1", "IngressNginxController", false)

	Context("Fresh cluster", func() {
		BeforeEach(func() {
			f.BindingContexts.Set(f.KubeStateSet(""))
			f.RunHook()
		})
		It("Should run", func() {
			Expect(f).To(ExecuteSuccessfully())
			Expect(f.BindingContexts.Array()).ShouldNot(BeEmpty())
		})

		Context("After adding ingress nginx controller object and webhook certificate", func() {
			BeforeEach(func() {
				f.BindingContexts.Set(f.KubeStateSet(`
---
apiVersion: deckhouse.io/v1
kind: IngressNginxController
metadata:
  name: test
spec:
  ingressClass: nginx
  inlet: LoadBalancer
  controllerVersion: "0.26"
  acceptRequestsFrom:
  - 127.0.0.1/32
  - 192.168.0.0/24
`))
				f.RunHook()
			})

			It("Should store ingress controller crds to values", func() {
				Expect(f).To(ExecuteSuccessfully())
				Expect(f.BindingContexts.Array()).ShouldNot(BeEmpty())

				Expect(f.ValuesGet("ingressNginx.internal.ingressControllers").String()).To(MatchJSON(`[{
"name": "test",
"spec": {
  "config": {},
  "ingressClass": "nginx",
  "controllerVersion": "0.26",
  "inlet": "LoadBalancer",
  "loadBalancer": {},
  "hstsOptions": {},
  "geoIP2": {},
  "resourcesRequests": {
    "mode": "VPA",
    "static": {},
    "vpa": {
      "cpu": {},
      "memory": {}
    }
  },
  "hostPortWithProxyProtocol": {},
  "hostPort": {},
  "hostWithFailover": {},
  "loadBalancerWithProxyProtocol": {},
  "acceptRequestsFrom": [
    "127.0.0.1/32",
    "192.168.0.0/24"
  ]
}
}]`))
			})
		})
	})

	Context("With Ingress Nginx Controller resource", func() {
		BeforeEach(func() {
			f.BindingContexts.Set(f.KubeStateSet(`
---
apiVersion: deckhouse.io/v1
kind: IngressNginxController
metadata:
  name: test
spec:
  ingressClass: nginx
  inlet: LoadBalancer
  resourcesRequests:
    mode: Static
---
apiVersion: deckhouse.io/v1
kind: IngressNginxController
metadata:
  name: test-2
spec:
  ingressClass: test
  inlet: HostPortWithProxyProtocol
  resourcesRequests:
    mode: VPA
    vpa:
      mode: Auto
      cpu:
        max: 100m
      memory:
        max: 200Mi
  hostPortWithProxyProtocol:
    httpPort: 80
    httpsPort: 443
---
apiVersion: deckhouse.io/v1
kind: IngressNginxController
metadata:
  name: test-3
spec:
  ingressClass: test
  inlet: LoadBalancerWithProxyProtocol
`))
			f.RunHook()
		})
		It("Should store ingress controller crds to values", func() {
			Expect(f).To(ExecuteSuccessfully())
			Expect(f.BindingContexts.Array()).ShouldNot(BeEmpty())

			Expect(f.ValuesGet("ingressNginx.internal.ingressControllers").Array()).To(HaveLen(3))

			Expect(f.ValuesGet("ingressNginx.internal.ingressControllers.0.name").String()).To(Equal("test"))
			Expect(f.ValuesGet("ingressNginx.internal.ingressControllers.0.spec").String()).To(MatchJSON(`{
"config": {},
"ingressClass": "nginx",
"controllerVersion": "0.25",
"inlet": "LoadBalancer",
"hstsOptions": {},
"geoIP2": {},
"resourcesRequests": {
  "mode": "Static",
  "static": {},
  "vpa": {"cpu": {}, "memory": {}}
},
"loadBalancer": {},
"loadBalancerWithProxyProtocol": {},
"hostPortWithProxyProtocol": {},
"hostWithFailover": {},
"hostPort": {}
}`))

			Expect(f.ValuesGet("ingressNginx.internal.ingressControllers.1.name").String()).To(Equal("test-2"))
			Expect(f.ValuesGet("ingressNginx.internal.ingressControllers.1.spec").String()).To(MatchJSON(`{
"config": {},
"ingressClass": "test",
"controllerVersion": "0.25",
"inlet": "HostPortWithProxyProtocol",
"hstsOptions": {},
"geoIP2": {},
"resourcesRequests": {
  "mode": "VPA",
  "static": {},
  "vpa": {"cpu": {"max": "100m"}, "memory": {"max": "200Mi"}, "mode": "Auto"}
},
"loadBalancer": {},
"loadBalancerWithProxyProtocol": {},
"hostPortWithProxyProtocol": {"httpPort": 80, "httpsPort": 443},
"hostWithFailover": {},
"hostPort": {}
}`))

			Expect(f.ValuesGet("ingressNginx.internal.ingressControllers.2.name").String()).To(Equal("test-3"))
			Expect(f.ValuesGet("ingressNginx.internal.ingressControllers.2.spec").String()).To(MatchJSON(`{
"config": {},
"ingressClass": "test",
"controllerVersion": "0.25",
"inlet": "LoadBalancerWithProxyProtocol",
"hstsOptions": {},
"geoIP2": {},
"resourcesRequests": {
  "mode": "VPA",
  "static": {},
  "vpa": {"cpu": {}, "memory": {}}
},
"loadBalancer": {},
"loadBalancerWithProxyProtocol": {},
"hostPortWithProxyProtocol": {},
"hostWithFailover": {},
"hostPort": {}
}`))
		})
	})
})
