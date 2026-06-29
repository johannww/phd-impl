{{- define "exp-app.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "exp-app.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- include "exp-app.name" . -}}
{{- end -}}
{{- end -}}

{{- define "exp-app.labels" -}}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}
app.kubernetes.io/name: {{ include "exp-app.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app: exp-app
{{- end -}}

{{- define "exp-app.namespace" -}}
{{- $global := default (dict) .Values.global -}}
{{- default .Release.Namespace (get $global "namespace") -}}
{{- end -}}

{{- define "exp-app.orgFullname" -}}
{{- $root := .root -}}
{{- $org := .org -}}
{{- printf "%s-%s" (include "exp-app.fullname" $root) $org | trunc 63 | trimSuffix "-" -}}
{{- end -}}
