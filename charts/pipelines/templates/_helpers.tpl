{{/*
Generate fullname of minio sub-chart
*/}}
{{- define "pipeline.minio.fullname" -}}
{{ include "common.names.dependency.fullname" (dict "chartName" "minio" "chartValues" .Values.minio "context" $) }}
{{- end -}}

{{/*
Generate endpoint of minio service
*/}}
{{- define "pipeline.minio.endpoint" -}}
{{- $port := .Values.minio.service.ports.api | toString -}}
{{- printf "%s.%s:%s" (include "pipeline.minio.fullname" .) (include "common.names.namespace" .) $port  -}}
{{- end }}

{{/*
Generate fullname of mysql sub-chart
*/}}
{{- define "pipeline.mysql.fullname" -}}
{{ include "common.names.dependency.fullname" (dict "chartName" "mysql" "chartValues" .Values.mysql "context" $) }}
{{- end -}}

{{/*
Generate service name of mysql which used as host in pipeline's installation configmap
*/}}
{{- define "pipeline.mysql.host" -}}
{{- if eq .Values.mysql.architecture "replication" }}
{{- printf "%s-%s" (include "pipeline.mysql.fullname" .) "primary" | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- include "pipeline.mysql.fullname" . -}}
{{- end -}}
{{- end -}}
