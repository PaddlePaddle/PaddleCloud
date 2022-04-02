{{/*
Generate fullname of redis sub-chart
*/}}
{{- define "pipeline.redis.fullname" -}}
{{ include "common.names.dependency.fullname" (dict "chartName" "redis" "chartValues" .Values.redis "context" $) }}
{{- end -}}


{{/*
Generate fullname of minio sub-chart
*/}}
{{- define "pipeline.minio.fullname" -}}
{{ include "common.names.dependency.fullname" (dict "chartName" "minio" "chartValues" .Values.pipelines.minio "context" $) }}
{{- end -}}


{{/*
Generate the MetaUrl of JuiceFS csi driver
*/}}
{{- define "paddlecloud.juicefs.metaurl" -}}
{{- $svcName := printf "%s-master" (include "pipeline.redis.fullname" .) -}}
{{- $host := printf "%s.%s" $svcName .Release.Namespace -}}
{{- $svcPort := .Values.redis.master.service.ports.redis | toString -}}
{{- if .Values.global.redis.password -}}
{{ printf "redis://default:%s@%s:%s/0" .Values.global.redis.password $host $svcPort }}
{{- else if and .Values.redis.auth.enabled .Values.redis.auth.password -}}
{{- $pwd := .Values.redis.auth.password -}}
{{ printf "redis://default:%s@%s:%s/0" .Values.redis.auth.password $host $svcPort }}
{{- end -}}
{{ printf "redis://%s:%s/0" $host $svcPort }}
{{- end -}}


{{/*
Generate the data center bucket of JuiceFS csi driver
*/}}
{{- define "paddlecloud.juicefs.metaurl" -}}
{{- $svcName := include "pipeline.minio.fullname" . -}}
{{- $port := .Values.minio.service.ports.api | toString -}}
{{- $endpoint := printf "%s.%s:%s" $svcName .Release.Namespace $port -}}
{{ printf "http://%s/data-center" $endpoint }}
{{- end -}}