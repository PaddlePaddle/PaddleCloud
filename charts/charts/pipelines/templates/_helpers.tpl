{{/*
Create the name of the service account to use
*/}}
{{- define "pipeline.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "pipeline.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}
