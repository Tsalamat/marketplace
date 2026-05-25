{{/*
Expand the name of the chart.
*/}}
{{- define "my_project.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "my_project.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "my_project.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "my_project.labels" -}}
helm.sh/chart: {{ include "my_project.chart" . }}
{{ include "my_project.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "my_project.selectorLabels" -}}
app.kubernetes.io/name: {{ include "my_project.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create a component-scoped Kubernetes name.
*/}}
{{- define "my_project.componentFullname" -}}
{{- printf "%s-%s" (include "my_project.fullname" .root) .component | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Selector labels for one component.
*/}}
{{- define "my_project.componentSelectorLabels" -}}
{{ include "my_project.selectorLabels" .root }}
app.kubernetes.io/component: {{ .component }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "my_project.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "my_project.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Resolve a secret value while preserving the live value on upgrades.
*/}}
{{- define "my_project.secretValue" -}}
{{- if and .secret .secret.data (hasKey .secret.data .key) -}}
{{- index .secret.data .key | b64dec -}}
{{- else -}}
{{- .value -}}
{{- end -}}
{{- end }}
