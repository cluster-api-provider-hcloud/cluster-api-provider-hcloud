{{- define "cluster.HcloudMachineDeployments" }}
{{- range $deployment := .Values.hcloudWorkers -}}
apiVersion: cluster.x-k8s.io/v1alpha4
kind: MachineDeployment
metadata:
  name: {{ $.Values.cluster.name }}-hc-workers-{{ $deployment.name }}
  namespace: {{ $.Values.cluster.namespace }}
  labels:
    nodepool: {{ $deployment.server.nodepool }}
spec:
  replicas: {{ $deployment.server.replicas }}
  revisionHistoryLimit: {{ $deployment.revisionHistoryLimit }}
  progressDeadlineSeconds: {{ $deployment.progressDeadlineSeconds }}
  clusterName: {{ $.Values.cluster.name }}
  selector:
    matchLabels:
      nodepool: {{ $deployment.server.nodepool }}
  template:
    metadata:
      labels:
        nodepool: {{ $deployment.server.nodepool }}
    spec:
      clusterName: {{ $.Values.cluster.name }}
      version: {{ $.Values.cluster.kubernetesVersion }}
      bootstrap:
        configRef:
          apiVersion: bootstrap.cluster.x-k8s.io/v1alpha4
          kind: KubeadmConfigTemplate
          name: {{ $.Values.cluster.name }}-hc-workers-{{ $deployment.name }}-ka-{{ $deployment.kubeadmTemplate.version }}
      infrastructureRef:
        apiVersion: cluster-api-provider-hcloud.capihc.com/v1alpha4
        kind: HcloudMachineTemplate
        name: {{ $.Values.cluster.name }}-hc-workers-{{ $deployment.name }}-mt-{{ $deployment.machineTemplate.version }}
      failureDomain: {{ $deployment.server.location }}
      nodeDrainTimeout: {{ $deployment.server.nodeDrainTimeout | quote }}
---
{{- end }}
{{- end }}

{{- define "cluster.HcloudMachineTemplates" }}
{{- range $mt := .Values.hcloudWorkers -}}
apiVersion: cluster-api-provider-hcloud.capihc.com/v1alpha4
kind: HcloudMachineTemplate
metadata:
  name: {{ $.Values.cluster.name }}-hc-workers-{{ $mt.name }}-mt-{{ $mt.machineTemplate.version }}
  namespace: {{ $.Values.cluster.namespace }}
spec:
  template:
    spec:
      type: {{ $mt.server.type }}
      image: {{ $mt.server.image }}
---
{{- end }}
{{- end }}

{{- define "cluster.HcloudKubeadmConfigTemplate" }}
{{- range $ka := .Values.hcloudWorkers -}}
apiVersion: bootstrap.cluster.x-k8s.io/v1alpha4
kind: KubeadmConfigTemplate
metadata:
  name: {{ $.Values.cluster.name }}-hc-workers-{{ $ka.name }}-ka-{{ $ka.kubeadmTemplate.version }}
  namespace: {{ $.Values.cluster.namespace }}
spec:
  template:
    spec:
      joinConfiguration:
        nodeRegistration:
          kubeletExtraArgs: 
            cloud-provider: external
            {{- with $ka.server.nodeLabels }}
            node-labels: node-access=ssh,capihc.com/type=bm,capihc.com/cpu={{ .cpu }},{{ .customLabels }}
            {{- end }}
      useExperimentalRetryJoin: {{ $ka.kubeadmTemplate.useExperimentalRetryJoin }}
      verbosity: {{ $ka.kubeadmTemplate.verbosity }}
---
{{- end }}
{{- end }}


{{- define "cluster.HcloudMachineHealthChecks" }}
{{- range $mhc := .Values.hcloudWorkers -}}
apiVersion: cluster.x-k8s.io/v1alpha4
kind: MachineHealthCheck
metadata:
  name: {{ $.Values.cluster.name }}-hc-workers-{{ $mhc.name }}-{{ $mhc.machineHealthCheck.name }}
spec:
  clusterName: {{ $.Values.cluster.name }}
  maxUnhealthy: {{ $mhc.machineHealthCheck.maxUnhealthy }}
  nodeStartupTimeout: {{ $mhc.machineHealthCheck.nodeStartupTimeout }}
  selector:
    matchLabels:
      nodepool: {{ $mhc.server.nodepool }}
  unhealthyConditions:
  - type: Ready
    status: Unknown
    timeout: {{ $mhc.machineHealthCheck.timeout }}
  - type: Ready
    status: "False"
    timeout: {{ $mhc.machineHealthCheck.timeout }}
---
{{- end }}
{{- end }}



{{- define "cluster.BaremetalMachineDeployments" }}
{{- range $deployment := .Values.hetznerBaremetalWorkers -}}
apiVersion: cluster.x-k8s.io/v1alpha4
kind: MachineDeployment
metadata:
  name: {{ $.Values.cluster.name }}-bm-workers-{{ $deployment.name }}
  namespace: {{ $.Values.cluster.namespace }}
  labels:
    nodepool: {{ $deployment.server.nodepool }}
spec:
  replicas: {{ $deployment.server.replicas }}
  revisionHistoryLimit: {{ $deployment.revisionHistoryLimit }}
  progressDeadlineSeconds: {{ $deployment.progressDeadlineSeconds }}
  clusterName: {{ $.Values.cluster.name }}
  selector:
    matchLabels:
      nodepool: {{ $deployment.server.nodepool }}
  template:
    metadata:
      labels:
        nodepool: {{ $deployment.server.nodepool }}
    spec:
      clusterName: {{ $.Values.cluster.name }}
      version: {{ $.Values.cluster.kubernetesVersion }}
      bootstrap:
        configRef:
          apiVersion: bootstrap.cluster.x-k8s.io/v1alpha4
          kind: KubeadmConfigTemplate
          name: {{ $.Values.cluster.name }}-bm-workers-{{ $deployment.name }}-ka-{{ $deployment.kubeadmTemplate.version }}
      infrastructureRef:
        apiVersion: cluster-api-provider-hcloud.capihc.com/v1alpha4
        kind: BareMetalMachineTemplate
        name: {{ $.Values.cluster.name }}-bm-workers-{{ $deployment.name }}-mt-{{ $deployment.machineTemplate.version }}
      # failureDomain: {{ $deployment.server.location }}
      nodeDrainTimeout: {{ $deployment.server.nodeDrainTimeout | quote }}
---
{{- end }}
{{- end }}



{{- define "cluster.BaremetalMachineTemplates" }}
{{- range $mt := .Values.hetznerBaremetalWorkers -}}
apiVersion: cluster-api-provider-hcloud.capihc.com/v1alpha4
kind: BareMetalMachineTemplate
metadata:
  name: {{ $.Values.cluster.name }}-bm-workers-{{ $mt.name }}-mt-{{ $mt.machineTemplate.version }}
  namespace: {{ $.Values.cluster.namespace }}
spec:
  template:
    spec:
      serverType: {{ $mt.server.type }}
      imagePath: {{ $mt.server.image }}
      {{- if $mt.server.customPartition.enabled }}
      partition: {{ $mt.server.customPartition.content }}
      {{- end }}
      sshTokenRef:
        publicKey: {{ $.Values.cluster.tokenRef.baremetal.sshRef.publicKey }}
        privateKey: {{ $.Values.cluster.tokenRef.baremetal.sshRef.privateKey }}
        sshKeyName: {{ $.Values.cluster.tokenRef.baremetal.sshRef.key }}
        tokenName: {{ $.Values.cluster.tokenRef.baremetal.sshRef.name }}
---
{{- end }}
{{- end }}


{{- define "cluster.BaremetalKubeadmConfigTemplate" }}
{{- range $ka := .Values.hetznerBaremetalWorkers -}}
apiVersion: bootstrap.cluster.x-k8s.io/v1alpha4
kind: KubeadmConfigTemplate
metadata:
  name: {{ $.Values.cluster.name }}-bm-workers-{{ $ka.name }}-ka-{{ $ka.kubeadmTemplate.version }}
  namespace: {{ $.Values.cluster.namespace }}
spec:
  template:
    spec:
      joinConfiguration:
        nodeRegistration:
          kubeletExtraArgs: 
            cloud-provider: external
            {{- with $ka.server.nodeLabels }}
            node-labels: node-access=ssh,capihc.com/cpu={{ .cpu }},{{ .customLabels }}
            {{- end }}
      useExperimentalRetryJoin: {{ $ka.kubeadmTemplate.useExperimentalRetryJoin }}
      verbosity: {{ $ka.kubeadmTemplate.verbosity }}
---
{{- end }}
{{- end }}


{{- define "cluster.BaremetalMachineHealthChecks" }}
{{- range $mhc := .Values.hetznerBaremetalWorkers -}}
apiVersion: cluster.x-k8s.io/v1alpha4
kind: MachineHealthCheck
metadata:
  name: {{ $.Values.cluster.name }}-bm-workers-{{ $mhc.name }}-{{ $mhc.machineHealthCheck.name }}
spec:
  clusterName: {{ $.Values.cluster.name }}
  maxUnhealthy: {{ $mhc.machineHealthCheck.maxUnhealthy }}
  nodeStartupTimeout: {{ $mhc.machineHealthCheck.nodeStartupTimeout }}
  selector:
    matchLabels:
      nodepool: {{ $mhc.server.nodepool }}
  unhealthyConditions:
  - type: Ready
    status: Unknown
    timeout: {{ $mhc.machineHealthCheck.timeout }}
  - type: Ready
    status: "False"
    timeout: {{ $mhc.machineHealthCheck.timeout }}
---
{{- end }}
{{- end }}
