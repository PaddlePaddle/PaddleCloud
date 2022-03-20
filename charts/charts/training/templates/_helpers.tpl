{{/*
Create the name of the service account to use
*/}}
{{- define "training.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "training.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}
