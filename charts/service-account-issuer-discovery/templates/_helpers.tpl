{{- define "name" -}}
service-account-issuer-discovery
{{- end -}}

{{- define "tlsEnabled" -}}
{{- if or .Values.gardenerManagedDNS.enabled .Values.gardenerManagedCertificate.enabled ( eq .Values.serviceType "LoadBalancer" ) .Values.tlsConfig.enabled -}}
true
{{- end -}}
{{- end -}}
