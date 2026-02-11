{{/*
Expand the name of the chart.
*/}}
{{- define "observability-stack.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "observability-stack.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "observability-stack.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "observability-stack.labels" -}}
helm.sh/chart: {{ include "observability-stack.chart" . }}
{{ include "observability-stack.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: observability-stack
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "observability-stack.selectorLabels" -}}
app.kubernetes.io/name: {{ include "observability-stack.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "observability-stack.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
{{- default (include "observability-stack.fullname" .) .Values.serviceAccount.name -}}
{{- else -}}
{{- default "default" .Values.serviceAccount.name -}}
{{- end -}}
{{- end -}}

{{/*
Common labels for a specific component
Usage: {{ include "observability-stack.componentLabels" (dict "component" "prometheus" "type" "monitoring" "context" $) }}
*/}}
{{- define "observability-stack.componentLabels" -}}
app: {{ .component }}
component: {{ .type | default "observability" }}
{{ include "observability-stack.labels" .context }}
{{- end -}}

{{/*
Selector labels for a specific component
Usage: {{ include "observability-stack.componentSelectorLabels" (dict "component" "prometheus" "context" $) }}
*/}}
{{- define "observability-stack.componentSelectorLabels" -}}
app: {{ .component }}
{{ include "observability-stack.selectorLabels" .context }}
{{- end -}}

{{/*
Generate the namespace
*/}}
{{- define "observability-stack.namespace" -}}
{{- .Values.global.namespace | default .Release.Namespace -}}
{{- end -}}

{{/*
Return the appropriate apiVersion for deployment
*/}}
{{- define "observability-stack.deployment.apiVersion" -}}
{{- if semverCompare ">=1.14-0" .Capabilities.KubeVersion.GitVersion -}}
apps/v1
{{- else -}}
apps/v1beta2
{{- end -}}
{{- end -}}

{{/*
Return the appropriate apiVersion for ingress
*/}}
{{- define "observability-stack.ingress.apiVersion" -}}
{{- if semverCompare ">=1.19-0" .Capabilities.KubeVersion.GitVersion -}}
networking.k8s.io/v1
{{- else if semverCompare ">=1.14-0" .Capabilities.KubeVersion.GitVersion -}}
networking.k8s.io/v1beta1
{{- else -}}
extensions/v1beta1
{{- end -}}
{{- end -}}

{{/*
Return the storage class name
*/}}
{{- define "observability-stack.storageClass" -}}
{{- if .Values.global.storageClass -}}
{{- .Values.global.storageClass -}}
{{- else -}}
{{- "standard" -}}
{{- end -}}
{{- end -}}

{{/*
Return the image pull policy
*/}}
{{- define "observability-stack.imagePullPolicy" -}}
{{- .Values.global.imagePullPolicy | default "IfNotPresent" -}}
{{- end -}}

{{/*
Create a default service name for a component
Usage: {{ include "observability-stack.componentServiceName" (dict "component" "prometheus" "context" $) }}
*/}}
{{- define "observability-stack.componentServiceName" -}}
{{- .component -}}
{{- end -}}

{{/*
Return the proper image name
Usage: {{ include "observability-stack.image" (dict "image" .Values.prometheus.image "context" $) }}
*/}}
{{- define "observability-stack.image" -}}
{{- $registryName := .image.registry -}}
{{- $repositoryName := .image.repository -}}
{{- $tag := .image.tag | toString -}}
{{- if $registryName -}}
{{- printf "%s/%s:%s" $registryName $repositoryName $tag -}}
{{- else -}}
{{- printf "%s:%s" $repositoryName $tag -}}
{{- end -}}
{{- end -}}

{{/*
Renders a value that contains template.
Usage: {{ include "observability-stack.tplvalues.render" (dict "value" .Values.path.to.value "context" $) }}
*/}}
{{- define "observability-stack.tplvalues.render" -}}
{{- if typeIs "string" .value -}}
{{- tpl .value .context -}}
{{- else -}}
{{- tpl (.value | toYaml) .context -}}
{{- end -}}
{{- end -}}

{{/*
Validate values - Check if required values are set
*/}}
{{- define "observability-stack.validateValues" -}}
{{- $messages := list -}}
{{- if not .Values.global.namespace -}}
{{- $messages = append $messages "global.namespace is required" -}}
{{- end -}}
{{- if $messages -}}
{{- fail (join ", " $messages) -}}
{{- end -}}
{{- end -}}
