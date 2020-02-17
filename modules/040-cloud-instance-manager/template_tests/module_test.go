package template_tests

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/deckhouse/deckhouse/testing/helm"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "")
}

const globalValues = `
enabledModules: ["vertical-pod-autoscaler-crd"]
modulesImages:
  registry: registry.flant.com
  registryDockercfg: cfg
  tags:
    cloudInstanceManager:
      clusterAutoscaler: imagehash
      machineControllerManager: imagehash
discovery:
  clusterMasterCount: "3"
  clusterUUID: f49dd1c3-a63a-4565-a06c-625e35587eab
  clusterVersion: 1.15.4
`

const cloudInstanceManagerAWS = `
instancePrefix: myprefix
internal:
  clusterMasterAddresses: ["10.0.0.1", "10.0.0.2", "10.0.0.3"]
  clusterCA: myclusterca
  bootstrapToken: mysecrettoken
  cloudProvider:
    type: aws
    machineClassKind: AWSInstanceClass
    aws:
      providerAccessKeyId: myprovaccesskeyid
      providerSecretAccessKey: myprovsecretaccesskey
      region: myregion
      loadBalancerSecurityGroupID: mylbsecuritygroupid
      keyName: mykeyname
      instances:
        iamProfileName: myiamprofilename
        securityGroupIDs: ["mysecgroupid1", "mysecgroupid2"]
        extraTags: ["extratag1", "extratag2"]
      internal:
        zoneToSubnetIdMap:
          zonea: mysubnetida
          zoneb: mysubnetidb
  instanceGroups:
  - instanceClass:
      ami: myami
      bashible:
        bundle: ubuntu-18.04-1.0
        dynamicOptions: {}
        options:
          kubernetesVersion: 1.16.6
      diskSizeGb: 50
      diskType: gp2
      iops: 42
      instanceType: t2.medium
    instanceClassReference:
      kind: AWSInstanceClass
      name: worker
    maxInstancesPerZone: 5
    minInstancesPerZone: 2
    name: worker
    zones:
    - zonea
    - zoneb
`

const cloudInstanceManagerGCP = `
instancePrefix: myprefix
internal:
  clusterMasterAddresses: ["10.0.0.1", "10.0.0.2", "10.0.0.3"]
  clusterCA: myclusterca
  bootstrapToken: mysecrettoken
  cloudProvider:
    type: gcp
    machineClassKind: GCPMachineClass
    gcp:
      networkName: mynetwork
      subnetworkName: mysubnet
      region: myreg
      extraInstanceTags: [aaa,bbb] #optional
      sshKey: mysshkey
      serviceAccountKey: '{"my":"key"}'
      disableExternalIP: true
  instanceGroups:
  - instanceClass: # maximum filled
      bashible:
        bundle: ubuntu-18.04-1.0
        dynamicOptions: {}
        options:
          kubernetesVersion: 1.15.4
      flavorName: m1.large
      imageName: ubuntu-18-04-cloud-amd64
      machineType: mymachinetype
      preemptible: true #optional
      diskType: superdisk #optional
      diskSizeGb: 42 #optional
    instanceClassReference:
      kind: GCPInstanceClass
      name: worker
    maxInstancesPerZone: 5
    minInstancesPerZone: 2
    name: worker
    zones:
    - zonea
    - zoneb
`

const cloudInstanceManagerOpenstack = `
instancePrefix: myprefix
internal:
  clusterMasterAddresses: ["10.0.0.1", "10.0.0.2", "10.0.0.3"]
  clusterCA: myclusterca
  bootstrapToken: mysecrettoken
  cloudProvider:
    type: openstack
    machineClassKind: OpenStackMachineClass
    openstack:
      addPodSubnetToPortWhitelist: true
      authURL: https://mycloud.qqq/3/
      caCert: mycacert
      domainName: Default
      internalNetworkName: mynetwork
      networkName: shared
      password: pPaAsS
      region: myreg
      securityGroups: [groupa, groupb]
      sshKeyPairName: mysshkey
      internalSubnet: "10.0.0.1/24"
      tenantName: mytname
      username: myuname
  instanceGroups:
  - instanceClass:
      bashible:
        bundle: ubuntu-18.04-1.0
        dynamicOptions: {}
        options:
          kubernetesVersion: 1.15.4
      flavorName: m1.large
      imageName: ubuntu-18-04-cloud-amd64
    instanceClassReference:
      kind: OpenStackInstanceClass
      name: worker
    maxInstancesPerZone: 5
    minInstancesPerZone: 2
    name: worker
    zones:
    - zonea
    - zoneb
`

const cloudInstanceManagerVsphere = `
instancePrefix: myprefix
internal:
  clusterMasterAddresses: ["10.0.0.1", "10.0.0.2", "10.0.0.3"]
  clusterCA: myclusterca
  bootstrapToken: mysecrettoken
  cloudProvider:
    type: vsphere
    machineClassKind: VsphereMachineClass
    vsphere:
      host: myhost.qqq
      username: myname
      password: pAsSwOrd
      insecure: true #
      regionTagCategory: myregtagcat #
      zoneTagCategory: myzonetagcateg #
      region: myreg
      sshKeys: [key1, key2] #
      vmFolderPath: dev/test
  instanceGroups:
  - instanceClass:
      bashible:
        bundle: ubuntu-18.04-1.0
        dynamicOptions: {}
        options:
          kubernetesVersion: 1.15.4
      flavorName: m1.large
      imageName: ubuntu-18-04-cloud-amd64
      numCPUs: 3
      memory: 3
      rootDiskSize: 42
      template: dev/test
      mainNetwork: mymainnetwork
      additionalNetworks: [aaa, bbb]
      datastore: lun-111
      runtimeOptions: # optional
        nestedHardwareVirtualization: true
        memoryReservation: 42
    instanceClassReference:
      kind: VsphereInstanceClass
      name: worker
    maxInstancesPerZone: 5
    minInstancesPerZone: 2
    name: worker
    zones:
    - zonea
    - zoneb
`

const cloudInstanceManagerYandex = `
instancePrefix: myprefix
internal:
  clusterMasterAddresses: ["10.0.0.1", "10.0.0.2", "10.0.0.3"]
  clusterCA: myclusterca
  bootstrapToken: mysecrettoken
  cloudProvider:
    type: yandex
    machineClassKind: YandexMachineClass
    yandex:
      serviceAccountJSON: '{"my":"svcacc"}'
      region: myreg
      folderID: myfolder
      sshKey: mysshkey
      sshUser: mysshuser
      nameservers: ["4.2.2.2"]
      zoneToSubnetIdMap:
        zonea: subneta
        zoneb: subnetb
  instanceGroups:
  - instanceClass:
      bashible:
        bundle: ubuntu-18.04-1.0
        dynamicOptions: {}
        options:
          kubernetesVersion: 1.15.4
      flavorName: m1.large
      imageName: ubuntu-18-04-cloud-amd64
      platformID: myplaid
      cores: 42
      coreFraction: 50 #optional
      memory: 42
      gpus: 2
      imageID: myimageid
      preemptible: true #optional
      diskType: ssd #optional
      diskSizeGB: 42 #optional
      assignPublicIPAddress: true #optional
      mainSubnet: mymainsubnet
      additionalSubnets: [aaa, bbb]
      labels: # optional
        my: label
    instanceClassReference:
      kind: YandexInstanceClass
      name: worker
    maxInstancesPerZone: 5
    minInstancesPerZone: 2
    name: worker
    zones:
    - zonea
    - zoneb
`

var _ = Describe("Module :: cloud-instance-manager :: helm template ::", func() {
	f := SetupHelmConfig(``)

	BeforeEach(func() {
		f.ValuesSetFromYaml("global", globalValues)
	})

	Context("AWS", func() {
		BeforeEach(func() {
			f.ValuesSetFromYaml("cloudInstanceManager", cloudInstanceManagerAWS)
			f.HelmRender()
		})

		It("Everything must render properly", func() {
			namespace := f.KubernetesGlobalResource("Namespace", "d8-cloud-instance-manager")
			registrySecret := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "deckhouse-registry")

			userAuthzClusterRoleUser := f.KubernetesGlobalResource("ClusterRole", "d8-cloud-instance-manager:user-authz:user")
			userAuthzClusterRoleClusterEditor := f.KubernetesGlobalResource("ClusterRole", "d8-cloud-instance-manager:user-authz:cluster-editor")
			userAuthzClusterRoleClusterAdmin := f.KubernetesGlobalResource("ClusterRole", "d8-cloud-instance-manager:user-authz:cluster-admin")

			mcmDeploy := f.KubernetesResource("Deployment", "d8-cloud-instance-manager", "machine-controller-manager")
			mcmServiceAccount := f.KubernetesResource("ServiceAccount", "d8-cloud-instance-manager", "machine-controller-manager")
			mcmRole := f.KubernetesResource("Role", "d8-cloud-instance-manager", "machine-controller-manager")
			mcmRoleBinding := f.KubernetesResource("RoleBinding", "d8-cloud-instance-manager", "machine-controller-manager")
			mcmClusterRole := f.KubernetesGlobalResource("ClusterRole", "d8-cloud-instance-manager:machine-controller-manager")
			mcmClusterRoleBinding := f.KubernetesGlobalResource("ClusterRoleBinding", "d8-cloud-instance-manager:machine-controller-manager")

			clusterAutoscalerDeploy := f.KubernetesResource("Deployment", "d8-cloud-instance-manager", "cluster-autoscaler")
			clusterAutoscalerServiceAccount := f.KubernetesResource("ServiceAccount", "d8-cloud-instance-manager", "cluster-autoscaler")
			clusterAutoscalerRole := f.KubernetesResource("Role", "d8-cloud-instance-manager", "cluster-autoscaler")
			clusterAutoscalerRoleBinding := f.KubernetesResource("RoleBinding", "d8-cloud-instance-manager", "cluster-autoscaler")
			clusterAutoscalerClusterRole := f.KubernetesGlobalResource("ClusterRole", "d8-cloud-instance-manager:cluster-autoscaler")
			clusterAutoscalerClusterRoleBinding := f.KubernetesGlobalResource("ClusterRoleBinding", "d8-cloud-instance-manager:cluster-autoscaler")

			machineClassA := f.KubernetesResource("AWSMachineClass", "d8-cloud-instance-manager", "worker-02320933")
			machineClassSecretA := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "worker-02320933")
			machineDeploymentA := f.KubernetesResource("MachineDeployment", "d8-cloud-instance-manager", "myprefix-worker-02320933")
			machineClassB := f.KubernetesResource("AWSMachineClass", "d8-cloud-instance-manager", "worker-6bdb5b0d")
			machineClassSecretB := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "worker-6bdb5b0d")
			machineDeploymentB := f.KubernetesResource("MachineDeployment", "d8-cloud-instance-manager", "myprefix-worker-6bdb5b0d")

			bashibleRole := f.KubernetesResource("Role", "d8-cloud-instance-manager", "machine-bootstrap")
			bashibleRoleBinding := f.KubernetesResource("RoleBinding", "d8-cloud-instance-manager", "machine-bootstrap")

			bashibleBundleCentos := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-centos-7-1.0")
			bashibleBundleCentosBootstrap := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-centos-7-1.0-bootstrap")
			bashibleBundleCentosWorker := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-centos-7-1.0-worker")
			bashibleBundleCentosWorkerBootstrap := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-centos-7-1.0-worker-bootstrap")
			bashibleBundlePreCooked := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-pre-cooked-1.0")
			bashibleBundlePreCookedBootstrap := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-pre-cooked-1.0-bootstrap")
			bashibleBundlePreCookedWorker := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-pre-cooked-1.0-worker")
			bashibleBundlePreCookedWorkerBootstrap := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-pre-cooked-1.0-worker-bootstrap")
			bashibleBundleUbuntu := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-ubuntu-18.04-1.0")
			bashibleBundleUbuntuBootstrap := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-ubuntu-18.04-1.0-bootstrap")
			bashibleBundleUbuntuWorker := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-ubuntu-18.04-1.0-worker")
			bashibleBundleUbuntuWorkerBootstrap := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-ubuntu-18.04-1.0-worker-bootstrap")

			Expect(namespace.Exists()).To(BeTrue())
			Expect(registrySecret.Exists()).To(BeTrue())

			Expect(userAuthzClusterRoleUser.Exists()).To(BeTrue())
			Expect(userAuthzClusterRoleClusterEditor.Exists()).To(BeTrue())
			Expect(userAuthzClusterRoleClusterAdmin.Exists()).To(BeTrue())

			Expect(mcmDeploy.Exists()).To(BeTrue())
			Expect(mcmServiceAccount.Exists()).To(BeTrue())
			Expect(mcmRole.Exists()).To(BeTrue())
			Expect(mcmRoleBinding.Exists()).To(BeTrue())
			Expect(mcmClusterRole.Exists()).To(BeTrue())
			Expect(mcmClusterRoleBinding.Exists()).To(BeTrue())

			Expect(clusterAutoscalerDeploy.Exists()).To(BeTrue())
			Expect(clusterAutoscalerServiceAccount.Exists()).To(BeTrue())
			Expect(clusterAutoscalerRole.Exists()).To(BeTrue())
			Expect(clusterAutoscalerRoleBinding.Exists()).To(BeTrue())
			Expect(clusterAutoscalerClusterRole.Exists()).To(BeTrue())
			Expect(clusterAutoscalerClusterRoleBinding.Exists()).To(BeTrue())

			Expect(machineClassA.Exists()).To(BeTrue())
			Expect(machineClassSecretA.Exists()).To(BeTrue())
			Expect(machineDeploymentA.Exists()).To(BeTrue())
			Expect(machineDeploymentA.Field("spec.template.metadata.annotations.checksum/bashible-bundles-options").String()).To(Equal("d801592ae7c43d3b0fba96a805c8d9f7fd006b9726daf97ba7f7abc399a56b09"))
			Expect(machineDeploymentA.Field("spec.template.metadata.annotations.checksum/machine-class").String()).To(Equal("21b7f37222f1cbad6c644c0aa4eef85aa309b874ec725dc0cdc087ca06fc6c19"))

			Expect(machineClassB.Exists()).To(BeTrue())
			Expect(machineClassSecretB.Exists()).To(BeTrue())
			Expect(machineDeploymentB.Exists()).To(BeTrue())
			Expect(machineDeploymentB.Field("spec.template.metadata.annotations.checksum/bashible-bundles-options").String()).To(Equal("d801592ae7c43d3b0fba96a805c8d9f7fd006b9726daf97ba7f7abc399a56b09"))
			Expect(machineDeploymentB.Field("spec.template.metadata.annotations.checksum/machine-class").String()).To(Equal("21b7f37222f1cbad6c644c0aa4eef85aa309b874ec725dc0cdc087ca06fc6c19"))

			Expect(bashibleRole.Exists()).To(BeTrue())
			Expect(bashibleRoleBinding.Exists()).To(BeTrue())

			Expect(bashibleBundleCentos.Exists()).To(BeTrue())
			Expect(bashibleBundleCentosBootstrap.Exists()).To(BeTrue())
			Expect(bashibleBundleCentosWorker.Exists()).To(BeTrue())
			Expect(bashibleBundleCentosWorkerBootstrap.Exists()).To(BeTrue())
			Expect(bashibleBundlePreCooked.Exists()).To(BeTrue())
			Expect(bashibleBundlePreCookedBootstrap.Exists()).To(BeTrue())
			Expect(bashibleBundlePreCookedWorker.Exists()).To(BeTrue())
			Expect(bashibleBundlePreCookedWorkerBootstrap.Exists()).To(BeTrue())
			Expect(bashibleBundleUbuntu.Exists()).To(BeTrue())
			Expect(bashibleBundleUbuntuBootstrap.Exists()).To(BeTrue())
			Expect(bashibleBundleUbuntuWorker.Exists()).To(BeTrue())
			Expect(bashibleBundleUbuntuWorkerBootstrap.Exists()).To(BeTrue())
		})
	})

	Context("GCP", func() {
		BeforeEach(func() {
			f.ValuesSetFromYaml("cloudInstanceManager", cloudInstanceManagerGCP)
			f.HelmRender()
		})

		It("Everything must render properly", func() {
			namespace := f.KubernetesGlobalResource("Namespace", "d8-cloud-instance-manager")
			registrySecret := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "deckhouse-registry")

			userAuthzClusterRoleUser := f.KubernetesGlobalResource("ClusterRole", "d8-cloud-instance-manager:user-authz:user")
			userAuthzClusterRoleClusterEditor := f.KubernetesGlobalResource("ClusterRole", "d8-cloud-instance-manager:user-authz:cluster-editor")
			userAuthzClusterRoleClusterAdmin := f.KubernetesGlobalResource("ClusterRole", "d8-cloud-instance-manager:user-authz:cluster-admin")

			mcmDeploy := f.KubernetesResource("Deployment", "d8-cloud-instance-manager", "machine-controller-manager")
			mcmServiceAccount := f.KubernetesResource("ServiceAccount", "d8-cloud-instance-manager", "machine-controller-manager")
			mcmRole := f.KubernetesResource("Role", "d8-cloud-instance-manager", "machine-controller-manager")
			mcmRoleBinding := f.KubernetesResource("RoleBinding", "d8-cloud-instance-manager", "machine-controller-manager")
			mcmClusterRole := f.KubernetesGlobalResource("ClusterRole", "d8-cloud-instance-manager:machine-controller-manager")
			mcmClusterRoleBinding := f.KubernetesGlobalResource("ClusterRoleBinding", "d8-cloud-instance-manager:machine-controller-manager")

			clusterAutoscalerDeploy := f.KubernetesResource("Deployment", "d8-cloud-instance-manager", "cluster-autoscaler")
			clusterAutoscalerServiceAccount := f.KubernetesResource("ServiceAccount", "d8-cloud-instance-manager", "cluster-autoscaler")
			clusterAutoscalerRole := f.KubernetesResource("Role", "d8-cloud-instance-manager", "cluster-autoscaler")
			clusterAutoscalerRoleBinding := f.KubernetesResource("RoleBinding", "d8-cloud-instance-manager", "cluster-autoscaler")
			clusterAutoscalerClusterRole := f.KubernetesGlobalResource("ClusterRole", "d8-cloud-instance-manager:cluster-autoscaler")
			clusterAutoscalerClusterRoleBinding := f.KubernetesGlobalResource("ClusterRoleBinding", "d8-cloud-instance-manager:cluster-autoscaler")

			machineClassA := f.KubernetesResource("GCPMachineClass", "d8-cloud-instance-manager", "worker-02320933")
			machineClassSecretA := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "worker-02320933")
			machineDeploymentA := f.KubernetesResource("MachineDeployment", "d8-cloud-instance-manager", "myprefix-worker-02320933")
			machineClassB := f.KubernetesResource("GCPMachineClass", "d8-cloud-instance-manager", "worker-6bdb5b0d")
			machineClassSecretB := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "worker-6bdb5b0d")
			machineDeploymentB := f.KubernetesResource("MachineDeployment", "d8-cloud-instance-manager", "myprefix-worker-6bdb5b0d")

			bashibleRole := f.KubernetesResource("Role", "d8-cloud-instance-manager", "machine-bootstrap")
			bashibleRoleBinding := f.KubernetesResource("RoleBinding", "d8-cloud-instance-manager", "machine-bootstrap")

			bashibleBundleCentos := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-centos-7-1.0")
			bashibleBundleCentosBootstrap := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-centos-7-1.0-bootstrap")
			bashibleBundleCentosWorker := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-centos-7-1.0-worker")
			bashibleBundleCentosWorkerBootstrap := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-centos-7-1.0-worker-bootstrap")
			bashibleBundlePreCooked := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-pre-cooked-1.0")
			bashibleBundlePreCookedBootstrap := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-pre-cooked-1.0-bootstrap")
			bashibleBundlePreCookedWorker := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-pre-cooked-1.0-worker")
			bashibleBundlePreCookedWorkerBootstrap := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-pre-cooked-1.0-worker-bootstrap")
			bashibleBundleUbuntu := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-ubuntu-18.04-1.0")
			bashibleBundleUbuntuBootstrap := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-ubuntu-18.04-1.0-bootstrap")
			bashibleBundleUbuntuWorker := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-ubuntu-18.04-1.0-worker")
			bashibleBundleUbuntuWorkerBootstrap := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-ubuntu-18.04-1.0-worker-bootstrap")

			Expect(namespace.Exists()).To(BeTrue())
			Expect(registrySecret.Exists()).To(BeTrue())

			Expect(userAuthzClusterRoleUser.Exists()).To(BeTrue())
			Expect(userAuthzClusterRoleClusterEditor.Exists()).To(BeTrue())
			Expect(userAuthzClusterRoleClusterAdmin.Exists()).To(BeTrue())

			Expect(mcmDeploy.Exists()).To(BeTrue())
			Expect(mcmServiceAccount.Exists()).To(BeTrue())
			Expect(mcmRole.Exists()).To(BeTrue())
			Expect(mcmRoleBinding.Exists()).To(BeTrue())
			Expect(mcmClusterRole.Exists()).To(BeTrue())
			Expect(mcmClusterRoleBinding.Exists()).To(BeTrue())

			Expect(clusterAutoscalerDeploy.Exists()).To(BeTrue())
			Expect(clusterAutoscalerServiceAccount.Exists()).To(BeTrue())
			Expect(clusterAutoscalerRole.Exists()).To(BeTrue())
			Expect(clusterAutoscalerRoleBinding.Exists()).To(BeTrue())
			Expect(clusterAutoscalerClusterRole.Exists()).To(BeTrue())
			Expect(clusterAutoscalerClusterRoleBinding.Exists()).To(BeTrue())

			Expect(machineClassA.Exists()).To(BeTrue())
			Expect(machineClassSecretA.Exists()).To(BeTrue())
			Expect(machineDeploymentA.Exists()).To(BeTrue())

			Expect(machineDeploymentA.Field("spec.template.metadata.annotations.checksum/bashible-bundles-options").String()).To(Equal("d98bbed20612cd12e463d29a0d76837bb821a14810944aea2a2c19542e3d71be"))
			Expect(machineDeploymentA.Field("spec.template.metadata.annotations.checksum/machine-class").String()).To(Equal("a9e6ed184c6eab25aa7e47d3d4c7e5647fee9fa5bc2d35eb0232eab45749d3ae"))

			Expect(machineClassB.Exists()).To(BeTrue())
			Expect(machineClassSecretB.Exists()).To(BeTrue())
			Expect(machineDeploymentB.Exists()).To(BeTrue())
			Expect(machineDeploymentB.Field("spec.template.metadata.annotations.checksum/bashible-bundles-options").String()).To(Equal("d98bbed20612cd12e463d29a0d76837bb821a14810944aea2a2c19542e3d71be"))
			Expect(machineDeploymentB.Field("spec.template.metadata.annotations.checksum/machine-class").String()).To(Equal("a9e6ed184c6eab25aa7e47d3d4c7e5647fee9fa5bc2d35eb0232eab45749d3ae"))

			Expect(bashibleRole.Exists()).To(BeTrue())
			Expect(bashibleRoleBinding.Exists()).To(BeTrue())

			Expect(bashibleBundleCentos.Exists()).To(BeTrue())
			Expect(bashibleBundleCentosBootstrap.Exists()).To(BeTrue())
			Expect(bashibleBundleCentosWorker.Exists()).To(BeTrue())
			Expect(bashibleBundleCentosWorkerBootstrap.Exists()).To(BeTrue())
			Expect(bashibleBundlePreCooked.Exists()).To(BeTrue())
			Expect(bashibleBundlePreCookedBootstrap.Exists()).To(BeTrue())
			Expect(bashibleBundlePreCookedWorker.Exists()).To(BeTrue())
			Expect(bashibleBundlePreCookedWorkerBootstrap.Exists()).To(BeTrue())
			Expect(bashibleBundleUbuntu.Exists()).To(BeTrue())
			Expect(bashibleBundleUbuntuBootstrap.Exists()).To(BeTrue())
			Expect(bashibleBundleUbuntuWorker.Exists()).To(BeTrue())
			Expect(bashibleBundleUbuntuWorkerBootstrap.Exists()).To(BeTrue())
		})
	})

	Context("Openstack", func() {
		BeforeEach(func() {
			f.ValuesSetFromYaml("cloudInstanceManager", cloudInstanceManagerOpenstack)
			f.HelmRender()
		})

		It("Everything must render properly", func() {
			namespace := f.KubernetesGlobalResource("Namespace", "d8-cloud-instance-manager")
			registrySecret := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "deckhouse-registry")

			userAuthzClusterRoleUser := f.KubernetesGlobalResource("ClusterRole", "d8-cloud-instance-manager:user-authz:user")
			userAuthzClusterRoleClusterEditor := f.KubernetesGlobalResource("ClusterRole", "d8-cloud-instance-manager:user-authz:cluster-editor")
			userAuthzClusterRoleClusterAdmin := f.KubernetesGlobalResource("ClusterRole", "d8-cloud-instance-manager:user-authz:cluster-admin")

			mcmDeploy := f.KubernetesResource("Deployment", "d8-cloud-instance-manager", "machine-controller-manager")
			mcmServiceAccount := f.KubernetesResource("ServiceAccount", "d8-cloud-instance-manager", "machine-controller-manager")
			mcmRole := f.KubernetesResource("Role", "d8-cloud-instance-manager", "machine-controller-manager")
			mcmRoleBinding := f.KubernetesResource("RoleBinding", "d8-cloud-instance-manager", "machine-controller-manager")
			mcmClusterRole := f.KubernetesGlobalResource("ClusterRole", "d8-cloud-instance-manager:machine-controller-manager")
			mcmClusterRoleBinding := f.KubernetesGlobalResource("ClusterRoleBinding", "d8-cloud-instance-manager:machine-controller-manager")

			clusterAutoscalerDeploy := f.KubernetesResource("Deployment", "d8-cloud-instance-manager", "cluster-autoscaler")
			clusterAutoscalerServiceAccount := f.KubernetesResource("ServiceAccount", "d8-cloud-instance-manager", "cluster-autoscaler")
			clusterAutoscalerRole := f.KubernetesResource("Role", "d8-cloud-instance-manager", "cluster-autoscaler")
			clusterAutoscalerRoleBinding := f.KubernetesResource("RoleBinding", "d8-cloud-instance-manager", "cluster-autoscaler")
			clusterAutoscalerClusterRole := f.KubernetesGlobalResource("ClusterRole", "d8-cloud-instance-manager:cluster-autoscaler")
			clusterAutoscalerClusterRoleBinding := f.KubernetesGlobalResource("ClusterRoleBinding", "d8-cloud-instance-manager:cluster-autoscaler")

			machineClassA := f.KubernetesResource("OpenstackMachineClass", "d8-cloud-instance-manager", "worker-02320933")
			machineClassSecretA := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "worker-02320933")
			machineDeploymentA := f.KubernetesResource("MachineDeployment", "d8-cloud-instance-manager", "myprefix-worker-02320933")
			machineClassB := f.KubernetesResource("OpenstackMachineClass", "d8-cloud-instance-manager", "worker-6bdb5b0d")
			machineClassSecretB := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "worker-6bdb5b0d")
			machineDeploymentB := f.KubernetesResource("MachineDeployment", "d8-cloud-instance-manager", "myprefix-worker-6bdb5b0d")

			bashibleRole := f.KubernetesResource("Role", "d8-cloud-instance-manager", "machine-bootstrap")
			bashibleRoleBinding := f.KubernetesResource("RoleBinding", "d8-cloud-instance-manager", "machine-bootstrap")

			bashibleBundleCentos := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-centos-7-1.0")
			bashibleBundleCentosBootstrap := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-centos-7-1.0-bootstrap")
			bashibleBundleCentosWorker := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-centos-7-1.0-worker")
			bashibleBundleCentosWorkerBootstrap := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-centos-7-1.0-worker-bootstrap")
			bashibleBundlePreCooked := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-pre-cooked-1.0")
			bashibleBundlePreCookedBootstrap := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-pre-cooked-1.0-bootstrap")
			bashibleBundlePreCookedWorker := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-pre-cooked-1.0-worker")
			bashibleBundlePreCookedWorkerBootstrap := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-pre-cooked-1.0-worker-bootstrap")
			bashibleBundleUbuntu := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-ubuntu-18.04-1.0")
			bashibleBundleUbuntuBootstrap := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-ubuntu-18.04-1.0-bootstrap")
			bashibleBundleUbuntuWorker := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-ubuntu-18.04-1.0-worker")
			bashibleBundleUbuntuWorkerBootstrap := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-ubuntu-18.04-1.0-worker-bootstrap")

			Expect(namespace.Exists()).To(BeTrue())
			Expect(registrySecret.Exists()).To(BeTrue())

			Expect(userAuthzClusterRoleUser.Exists()).To(BeTrue())
			Expect(userAuthzClusterRoleClusterEditor.Exists()).To(BeTrue())
			Expect(userAuthzClusterRoleClusterAdmin.Exists()).To(BeTrue())

			Expect(mcmDeploy.Exists()).To(BeTrue())
			Expect(mcmServiceAccount.Exists()).To(BeTrue())
			Expect(mcmRole.Exists()).To(BeTrue())
			Expect(mcmRoleBinding.Exists()).To(BeTrue())
			Expect(mcmClusterRole.Exists()).To(BeTrue())
			Expect(mcmClusterRoleBinding.Exists()).To(BeTrue())

			Expect(clusterAutoscalerDeploy.Exists()).To(BeTrue())
			Expect(clusterAutoscalerServiceAccount.Exists()).To(BeTrue())
			Expect(clusterAutoscalerRole.Exists()).To(BeTrue())
			Expect(clusterAutoscalerRoleBinding.Exists()).To(BeTrue())
			Expect(clusterAutoscalerClusterRole.Exists()).To(BeTrue())
			Expect(clusterAutoscalerClusterRoleBinding.Exists()).To(BeTrue())

			Expect(machineClassA.Exists()).To(BeTrue())
			Expect(machineClassSecretA.Exists()).To(BeTrue())
			Expect(machineDeploymentA.Exists()).To(BeTrue())

			Expect(machineDeploymentA.Field("spec.template.metadata.annotations.checksum/bashible-bundles-options").String()).To(Equal("d98bbed20612cd12e463d29a0d76837bb821a14810944aea2a2c19542e3d71be"))
			Expect(machineDeploymentA.Field("spec.template.metadata.annotations.checksum/machine-class").String()).To(Equal("bbfc6f35c09ffb41b71cbb1670803013cd247118a83169d6170bc5699176242f"))

			Expect(machineClassB.Exists()).To(BeTrue())
			Expect(machineClassSecretB.Exists()).To(BeTrue())
			Expect(machineDeploymentB.Exists()).To(BeTrue())
			Expect(machineDeploymentB.Field("spec.template.metadata.annotations.checksum/bashible-bundles-options").String()).To(Equal("d98bbed20612cd12e463d29a0d76837bb821a14810944aea2a2c19542e3d71be"))
			Expect(machineDeploymentB.Field("spec.template.metadata.annotations.checksum/machine-class").String()).To(Equal("bbfc6f35c09ffb41b71cbb1670803013cd247118a83169d6170bc5699176242f"))

			Expect(bashibleRole.Exists()).To(BeTrue())
			Expect(bashibleRoleBinding.Exists()).To(BeTrue())

			Expect(bashibleBundleCentos.Exists()).To(BeTrue())
			Expect(bashibleBundleCentosBootstrap.Exists()).To(BeTrue())
			Expect(bashibleBundleCentosWorker.Exists()).To(BeTrue())
			Expect(bashibleBundleCentosWorkerBootstrap.Exists()).To(BeTrue())
			Expect(bashibleBundlePreCooked.Exists()).To(BeTrue())
			Expect(bashibleBundlePreCookedBootstrap.Exists()).To(BeTrue())
			Expect(bashibleBundlePreCookedWorker.Exists()).To(BeTrue())
			Expect(bashibleBundlePreCookedWorkerBootstrap.Exists()).To(BeTrue())
			Expect(bashibleBundleUbuntu.Exists()).To(BeTrue())
			Expect(bashibleBundleUbuntuBootstrap.Exists()).To(BeTrue())
			Expect(bashibleBundleUbuntuWorker.Exists()).To(BeTrue())
			Expect(bashibleBundleUbuntuWorkerBootstrap.Exists()).To(BeTrue())
		})
	})

	Context("Vsphere", func() {
		BeforeEach(func() {
			f.ValuesSetFromYaml("cloudInstanceManager", cloudInstanceManagerVsphere)
			f.HelmRender()
		})

		It("Everything must render properly", func() {
			namespace := f.KubernetesGlobalResource("Namespace", "d8-cloud-instance-manager")
			registrySecret := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "deckhouse-registry")

			userAuthzClusterRoleUser := f.KubernetesGlobalResource("ClusterRole", "d8-cloud-instance-manager:user-authz:user")
			userAuthzClusterRoleClusterEditor := f.KubernetesGlobalResource("ClusterRole", "d8-cloud-instance-manager:user-authz:cluster-editor")
			userAuthzClusterRoleClusterAdmin := f.KubernetesGlobalResource("ClusterRole", "d8-cloud-instance-manager:user-authz:cluster-admin")

			mcmDeploy := f.KubernetesResource("Deployment", "d8-cloud-instance-manager", "machine-controller-manager")
			mcmServiceAccount := f.KubernetesResource("ServiceAccount", "d8-cloud-instance-manager", "machine-controller-manager")
			mcmRole := f.KubernetesResource("Role", "d8-cloud-instance-manager", "machine-controller-manager")
			mcmRoleBinding := f.KubernetesResource("RoleBinding", "d8-cloud-instance-manager", "machine-controller-manager")
			mcmClusterRole := f.KubernetesGlobalResource("ClusterRole", "d8-cloud-instance-manager:machine-controller-manager")
			mcmClusterRoleBinding := f.KubernetesGlobalResource("ClusterRoleBinding", "d8-cloud-instance-manager:machine-controller-manager")

			clusterAutoscalerDeploy := f.KubernetesResource("Deployment", "d8-cloud-instance-manager", "cluster-autoscaler")
			clusterAutoscalerServiceAccount := f.KubernetesResource("ServiceAccount", "d8-cloud-instance-manager", "cluster-autoscaler")
			clusterAutoscalerRole := f.KubernetesResource("Role", "d8-cloud-instance-manager", "cluster-autoscaler")
			clusterAutoscalerRoleBinding := f.KubernetesResource("RoleBinding", "d8-cloud-instance-manager", "cluster-autoscaler")
			clusterAutoscalerClusterRole := f.KubernetesGlobalResource("ClusterRole", "d8-cloud-instance-manager:cluster-autoscaler")
			clusterAutoscalerClusterRoleBinding := f.KubernetesGlobalResource("ClusterRoleBinding", "d8-cloud-instance-manager:cluster-autoscaler")

			machineClassA := f.KubernetesResource("VsphereMachineClass", "d8-cloud-instance-manager", "worker-02320933")
			machineClassSecretA := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "worker-02320933")
			machineDeploymentA := f.KubernetesResource("MachineDeployment", "d8-cloud-instance-manager", "myprefix-worker-02320933")
			machineClassB := f.KubernetesResource("VsphereMachineClass", "d8-cloud-instance-manager", "worker-6bdb5b0d")
			machineClassSecretB := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "worker-6bdb5b0d")
			machineDeploymentB := f.KubernetesResource("MachineDeployment", "d8-cloud-instance-manager", "myprefix-worker-6bdb5b0d")

			bashibleRole := f.KubernetesResource("Role", "d8-cloud-instance-manager", "machine-bootstrap")
			bashibleRoleBinding := f.KubernetesResource("RoleBinding", "d8-cloud-instance-manager", "machine-bootstrap")

			bashibleBundleCentos := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-centos-7-1.0")
			bashibleBundleCentosBootstrap := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-centos-7-1.0-bootstrap")
			bashibleBundleCentosWorker := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-centos-7-1.0-worker")
			bashibleBundleCentosWorkerBootstrap := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-centos-7-1.0-worker-bootstrap")
			bashibleBundlePreCooked := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-pre-cooked-1.0")
			bashibleBundlePreCookedBootstrap := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-pre-cooked-1.0-bootstrap")
			bashibleBundlePreCookedWorker := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-pre-cooked-1.0-worker")
			bashibleBundlePreCookedWorkerBootstrap := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-pre-cooked-1.0-worker-bootstrap")
			bashibleBundleUbuntu := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-ubuntu-18.04-1.0")
			bashibleBundleUbuntuBootstrap := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-ubuntu-18.04-1.0-bootstrap")
			bashibleBundleUbuntuWorker := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-ubuntu-18.04-1.0-worker")
			bashibleBundleUbuntuWorkerBootstrap := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-ubuntu-18.04-1.0-worker-bootstrap")

			Expect(namespace.Exists()).To(BeTrue())
			Expect(registrySecret.Exists()).To(BeTrue())

			Expect(userAuthzClusterRoleUser.Exists()).To(BeTrue())
			Expect(userAuthzClusterRoleClusterEditor.Exists()).To(BeTrue())
			Expect(userAuthzClusterRoleClusterAdmin.Exists()).To(BeTrue())

			Expect(mcmDeploy.Exists()).To(BeTrue())
			Expect(mcmServiceAccount.Exists()).To(BeTrue())
			Expect(mcmRole.Exists()).To(BeTrue())
			Expect(mcmRoleBinding.Exists()).To(BeTrue())
			Expect(mcmClusterRole.Exists()).To(BeTrue())
			Expect(mcmClusterRoleBinding.Exists()).To(BeTrue())

			Expect(clusterAutoscalerDeploy.Exists()).To(BeTrue())
			Expect(clusterAutoscalerServiceAccount.Exists()).To(BeTrue())
			Expect(clusterAutoscalerRole.Exists()).To(BeTrue())
			Expect(clusterAutoscalerRoleBinding.Exists()).To(BeTrue())
			Expect(clusterAutoscalerClusterRole.Exists()).To(BeTrue())
			Expect(clusterAutoscalerClusterRoleBinding.Exists()).To(BeTrue())

			Expect(machineClassA.Exists()).To(BeTrue())
			Expect(machineClassSecretA.Exists()).To(BeTrue())
			Expect(machineDeploymentA.Exists()).To(BeTrue())

			fmt.Println(machineDeploymentA.Field("spec.template.metadata.annotations.checksum/machine-class").String())
			Expect(machineDeploymentA.Field("spec.template.metadata.annotations.checksum/bashible-bundles-options").String()).To(Equal("d98bbed20612cd12e463d29a0d76837bb821a14810944aea2a2c19542e3d71be"))
			Expect(machineDeploymentA.Field("spec.template.metadata.annotations.checksum/machine-class").String()).To(Equal("e54154626facdf7ba3937af03fb11ac3e626cf1ebab8e36fb17c8320ed4ae906"))

			Expect(machineClassB.Exists()).To(BeTrue())
			Expect(machineClassSecretB.Exists()).To(BeTrue())
			Expect(machineDeploymentB.Exists()).To(BeTrue())
			Expect(machineDeploymentB.Field("spec.template.metadata.annotations.checksum/bashible-bundles-options").String()).To(Equal("d98bbed20612cd12e463d29a0d76837bb821a14810944aea2a2c19542e3d71be"))
			Expect(machineDeploymentB.Field("spec.template.metadata.annotations.checksum/machine-class").String()).To(Equal("e54154626facdf7ba3937af03fb11ac3e626cf1ebab8e36fb17c8320ed4ae906"))

			Expect(bashibleRole.Exists()).To(BeTrue())
			Expect(bashibleRoleBinding.Exists()).To(BeTrue())

			Expect(bashibleBundleCentos.Exists()).To(BeTrue())
			Expect(bashibleBundleCentosBootstrap.Exists()).To(BeTrue())
			Expect(bashibleBundleCentosWorker.Exists()).To(BeTrue())
			Expect(bashibleBundleCentosWorkerBootstrap.Exists()).To(BeTrue())
			Expect(bashibleBundlePreCooked.Exists()).To(BeTrue())
			Expect(bashibleBundlePreCookedBootstrap.Exists()).To(BeTrue())
			Expect(bashibleBundlePreCookedWorker.Exists()).To(BeTrue())
			Expect(bashibleBundlePreCookedWorkerBootstrap.Exists()).To(BeTrue())
			Expect(bashibleBundleUbuntu.Exists()).To(BeTrue())
			Expect(bashibleBundleUbuntuBootstrap.Exists()).To(BeTrue())
			Expect(bashibleBundleUbuntuWorker.Exists()).To(BeTrue())
			Expect(bashibleBundleUbuntuWorkerBootstrap.Exists()).To(BeTrue())
		})
	})

	Context("Yandex", func() {
		BeforeEach(func() {
			f.ValuesSetFromYaml("cloudInstanceManager", cloudInstanceManagerYandex)
			f.HelmRender()
		})

		It("Everything must render properly", func() {
			namespace := f.KubernetesGlobalResource("Namespace", "d8-cloud-instance-manager")
			registrySecret := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "deckhouse-registry")

			userAuthzClusterRoleUser := f.KubernetesGlobalResource("ClusterRole", "d8-cloud-instance-manager:user-authz:user")
			userAuthzClusterRoleClusterEditor := f.KubernetesGlobalResource("ClusterRole", "d8-cloud-instance-manager:user-authz:cluster-editor")
			userAuthzClusterRoleClusterAdmin := f.KubernetesGlobalResource("ClusterRole", "d8-cloud-instance-manager:user-authz:cluster-admin")

			mcmDeploy := f.KubernetesResource("Deployment", "d8-cloud-instance-manager", "machine-controller-manager")
			mcmServiceAccount := f.KubernetesResource("ServiceAccount", "d8-cloud-instance-manager", "machine-controller-manager")
			mcmRole := f.KubernetesResource("Role", "d8-cloud-instance-manager", "machine-controller-manager")
			mcmRoleBinding := f.KubernetesResource("RoleBinding", "d8-cloud-instance-manager", "machine-controller-manager")
			mcmClusterRole := f.KubernetesGlobalResource("ClusterRole", "d8-cloud-instance-manager:machine-controller-manager")
			mcmClusterRoleBinding := f.KubernetesGlobalResource("ClusterRoleBinding", "d8-cloud-instance-manager:machine-controller-manager")

			clusterAutoscalerDeploy := f.KubernetesResource("Deployment", "d8-cloud-instance-manager", "cluster-autoscaler")
			clusterAutoscalerServiceAccount := f.KubernetesResource("ServiceAccount", "d8-cloud-instance-manager", "cluster-autoscaler")
			clusterAutoscalerRole := f.KubernetesResource("Role", "d8-cloud-instance-manager", "cluster-autoscaler")
			clusterAutoscalerRoleBinding := f.KubernetesResource("RoleBinding", "d8-cloud-instance-manager", "cluster-autoscaler")
			clusterAutoscalerClusterRole := f.KubernetesGlobalResource("ClusterRole", "d8-cloud-instance-manager:cluster-autoscaler")
			clusterAutoscalerClusterRoleBinding := f.KubernetesGlobalResource("ClusterRoleBinding", "d8-cloud-instance-manager:cluster-autoscaler")

			machineClassA := f.KubernetesResource("YandexMachineClass", "d8-cloud-instance-manager", "worker-02320933")
			machineClassSecretA := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "worker-02320933")
			machineDeploymentA := f.KubernetesResource("MachineDeployment", "d8-cloud-instance-manager", "myprefix-worker-02320933")
			machineClassB := f.KubernetesResource("YandexMachineClass", "d8-cloud-instance-manager", "worker-6bdb5b0d")
			machineClassSecretB := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "worker-6bdb5b0d")
			machineDeploymentB := f.KubernetesResource("MachineDeployment", "d8-cloud-instance-manager", "myprefix-worker-6bdb5b0d")

			bashibleRole := f.KubernetesResource("Role", "d8-cloud-instance-manager", "machine-bootstrap")
			bashibleRoleBinding := f.KubernetesResource("RoleBinding", "d8-cloud-instance-manager", "machine-bootstrap")

			bashibleBundleCentos := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-centos-7-1.0")
			bashibleBundleCentosBootstrap := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-centos-7-1.0-bootstrap")
			bashibleBundleCentosWorker := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-centos-7-1.0-worker")
			bashibleBundleCentosWorkerBootstrap := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-centos-7-1.0-worker-bootstrap")
			bashibleBundlePreCooked := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-pre-cooked-1.0")
			bashibleBundlePreCookedBootstrap := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-pre-cooked-1.0-bootstrap")
			bashibleBundlePreCookedWorker := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-pre-cooked-1.0-worker")
			bashibleBundlePreCookedWorkerBootstrap := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-pre-cooked-1.0-worker-bootstrap")
			bashibleBundleUbuntu := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-ubuntu-18.04-1.0")
			bashibleBundleUbuntuBootstrap := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-ubuntu-18.04-1.0-bootstrap")
			bashibleBundleUbuntuWorker := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-ubuntu-18.04-1.0-worker")
			bashibleBundleUbuntuWorkerBootstrap := f.KubernetesResource("Secret", "d8-cloud-instance-manager", "bashible-bundle-ubuntu-18.04-1.0-worker-bootstrap")

			Expect(namespace.Exists()).To(BeTrue())
			Expect(registrySecret.Exists()).To(BeTrue())

			Expect(userAuthzClusterRoleUser.Exists()).To(BeTrue())
			Expect(userAuthzClusterRoleClusterEditor.Exists()).To(BeTrue())
			Expect(userAuthzClusterRoleClusterAdmin.Exists()).To(BeTrue())

			Expect(mcmDeploy.Exists()).To(BeTrue())
			Expect(mcmServiceAccount.Exists()).To(BeTrue())
			Expect(mcmRole.Exists()).To(BeTrue())
			Expect(mcmRoleBinding.Exists()).To(BeTrue())
			Expect(mcmClusterRole.Exists()).To(BeTrue())
			Expect(mcmClusterRoleBinding.Exists()).To(BeTrue())

			Expect(clusterAutoscalerDeploy.Exists()).To(BeTrue())
			Expect(clusterAutoscalerServiceAccount.Exists()).To(BeTrue())
			Expect(clusterAutoscalerRole.Exists()).To(BeTrue())
			Expect(clusterAutoscalerRoleBinding.Exists()).To(BeTrue())
			Expect(clusterAutoscalerClusterRole.Exists()).To(BeTrue())
			Expect(clusterAutoscalerClusterRoleBinding.Exists()).To(BeTrue())

			Expect(machineClassA.Exists()).To(BeTrue())
			Expect(machineClassSecretA.Exists()).To(BeTrue())
			Expect(machineDeploymentA.Exists()).To(BeTrue())

			Expect(machineDeploymentA.Field("spec.template.metadata.annotations.checksum/bashible-bundles-options").String()).To(Equal("d98bbed20612cd12e463d29a0d76837bb821a14810944aea2a2c19542e3d71be"))
			Expect(machineDeploymentA.Field("spec.template.metadata.annotations.checksum/machine-class").String()).To(Equal("74795e5fe09827e6c1b0a44968e667aa93a9c1ee34e9c6f0bb6994dbdb2bb2fd"))

			Expect(machineClassB.Exists()).To(BeTrue())
			Expect(machineClassSecretB.Exists()).To(BeTrue())
			Expect(machineDeploymentB.Exists()).To(BeTrue())
			Expect(machineDeploymentB.Field("spec.template.metadata.annotations.checksum/bashible-bundles-options").String()).To(Equal("d98bbed20612cd12e463d29a0d76837bb821a14810944aea2a2c19542e3d71be"))
			Expect(machineDeploymentB.Field("spec.template.metadata.annotations.checksum/machine-class").String()).To(Equal("74795e5fe09827e6c1b0a44968e667aa93a9c1ee34e9c6f0bb6994dbdb2bb2fd"))

			Expect(bashibleRole.Exists()).To(BeTrue())
			Expect(bashibleRoleBinding.Exists()).To(BeTrue())

			Expect(bashibleBundleCentos.Exists()).To(BeTrue())
			Expect(bashibleBundleCentosBootstrap.Exists()).To(BeTrue())
			Expect(bashibleBundleCentosWorker.Exists()).To(BeTrue())
			Expect(bashibleBundleCentosWorkerBootstrap.Exists()).To(BeTrue())
			Expect(bashibleBundlePreCooked.Exists()).To(BeTrue())
			Expect(bashibleBundlePreCookedBootstrap.Exists()).To(BeTrue())
			Expect(bashibleBundlePreCookedWorker.Exists()).To(BeTrue())
			Expect(bashibleBundlePreCookedWorkerBootstrap.Exists()).To(BeTrue())
			Expect(bashibleBundleUbuntu.Exists()).To(BeTrue())
			Expect(bashibleBundleUbuntuBootstrap.Exists()).To(BeTrue())
			Expect(bashibleBundleUbuntuWorker.Exists()).To(BeTrue())
			Expect(bashibleBundleUbuntuWorkerBootstrap.Exists()).To(BeTrue())
		})
	})

})