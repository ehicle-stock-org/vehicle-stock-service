{{- define "vehicle-stock-service.name" -}}
vehicle-stock-service
{{- end -}}

{{- define "vehicle-stock-service.fullname" -}}
{{- printf "%s" (include "vehicle-stock-service.name" .) -}}
{{- end -}}
