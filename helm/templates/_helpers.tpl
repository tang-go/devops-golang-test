{{/*namespace*/}}
{{- define "nm.namespaceOverride" -}}
  {{- if .Values.namespaceOverride -}}
    {{- .Values.namespaceOverride -}}
  {{- else -}}
    {{- .Release.Namespace -}}
  {{- end -}}
{{- end -}}

{{- define "global.imageRegistry" -}}
{{- $registry := default .Values.global.imageRegistry .Values.imageRegistryOverride }}
  {{- if $registry -}}
    {{- printf "%s/" $registry -}}
  {{- end -}}
{{- end -}}
