{{- define "node_group_static_or_hybrid_script" -}}
  {{- $context := index . 0 -}}
  {{- $ng := index . 1 -}}
  {{- $bootstrap_token := index . 2 -}}
  {{- $adopt := index . 3 -}}
#!/bin/bash

mkdir -p /var/lib/bashible

  {{- if hasKey $context.Values.nodeManager.internal "cloudProvider" }}
    {{- $bootstrap_scripts_from_bundles := list }}
    {{- range $path, $_ := $context.Files.Glob (printf "candi/cloud-providers/%s/bashible/bundles/*/bootstrap-networks.sh.tpl" $context.Values.nodeManager.internal.cloudProvider.type) }}
      {{- $bootstrap_scripts_from_bundles = append $bootstrap_scripts_from_bundles $path }}
    {{- end }}

    {{- $bootstrap_scripts_common := list }}
    {{- range $path, $_ := $context.Files.Glob (printf "candi/cloud-providers/%s/common-steps/bootstrap-networks.sh.tpl" $context.Values.nodeManager.internal.cloudProvider.type) }}
      {{- $bootstrap_scripts_common = append $bootstrap_scripts_common $path }}
    {{- end }}

    {{- $bootstrap_scripts := list  }}
    {{- if gt (len $bootstrap_scripts_common) 0 }}
      {{- $bootstrap_scripts = $bootstrap_scripts_common }}
    {{- else }}
      {{- $bootstrap_scripts = $bootstrap_scripts_from_bundles }}
    {{- end }}

    {{- range $path := $bootstrap_scripts }}
      {{- $bundle := (dir $path | base) }}
      {{- $tpl_context := dict }}
      {{- $_ := set $tpl_context "Release" $context.Release }}
      {{- $_ := set $tpl_context "Chart" $context.Chart }}
      {{- $_ := set $tpl_context "Files" $context.Files }}
      {{- $_ := set $tpl_context "Capabilities" $context.Capabilities }}
      {{- $_ := set $tpl_context "Template" $context.Template }}
      {{- $_ := set $tpl_context "Values" $context.Values }}
      {{- $_ := set $tpl_context "nodeGroup" $ng }}
cat > /var/lib/bashible/cloud-provider-bootstrap-networks-{{ $bundle }}.sh <<"EOF"
{{ tpl ($context.Files.Get $path) $tpl_context }}
EOF
chmod +x /var/lib/bashible/cloud-provider-bootstrap-networks-{{ $bundle }}.sh
    {{- end }}
  {{- end }}

  {{- $bashible_bootstrap_script_tpl_context := dict }}
  {{- $_ := set $bashible_bootstrap_script_tpl_context "normal" (dict "apiserverEndpoints" $context.Values.nodeManager.internal.clusterMasterAddresses) }}
  {{- $_ := set $bashible_bootstrap_script_tpl_context "nodeGroup" $ng }}
  {{- $_ := set $bashible_bootstrap_script_tpl_context "Template" $context.Template }}
  {{- $_ := set $bashible_bootstrap_script_tpl_context "Files" $context.Files }}
  {{- $_ := set $bashible_bootstrap_script_tpl_context "allowedBundles" $context.Values.nodeManager.allowedBundles }}
  {{- $_ := set $bashible_bootstrap_script_tpl_context "registry" (dict "address" $context.Values.global.modulesImages.registryAddress "path" $context.Values.global.modulesImages.registryPath "scheme" $context.Values.global.modulesImages.registryScheme "ca" $context.Values.global.modulesImages.registryCA "dockerCfg" $context.Values.global.modulesImages.registryDockercfg) }}
cat > /var/lib/bashible/bootstrap.sh <<"END"
{{ if $adopt }}
  {{- include "node_group_bashible_bootstrap_script_noninteractive" $bashible_bootstrap_script_tpl_context }}
{{ else }}
  {{- include "node_group_bashible_bootstrap_script" $bashible_bootstrap_script_tpl_context }}
{{ end }}
END
chmod +x /var/lib/bashible/bootstrap.sh

cat > /var/lib/bashible/ca.crt <<"EOF"
{{ $context.Values.nodeManager.internal.kubernetesCA }}
EOF

cat > /var/lib/bashible/bootstrap-token <<"EOF"
{{ $bootstrap_token }}
EOF
chmod 0600 /var/lib/bashible/bootstrap-token

{{- if not $adopt }}
touch /var/lib/bashible/first_run
{{- end }}

/var/lib/bashible/bootstrap.sh
{{ end }}
