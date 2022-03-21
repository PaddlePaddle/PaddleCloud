{{/* vim: set filetype=mustache: */}}

{{/*
Return the proper image name
{{ include "serving.image" ( dict "imageRoot" .Values.path.to.the.image ) }}
*/}}
{{- define "serving.image" -}}
{{- $registryName := .imageRoot.registry -}}
{{- $repositoryName := .imageRoot.repository -}}
{{- if .Values.global.imageRegistry }}
  {{- $registryName = .Values.global.imageRegistry -}}
{{- end -}}
{{- if $registryName }}
{{- printf "%s/%s:%s" $registryName $repositoryName .Chart.AppVersion -}}
{{- else -}}
{{- printf "%s:%s" $repositoryName .Chart.AppVersion -}}
{{- end -}}
{{- end -}}
