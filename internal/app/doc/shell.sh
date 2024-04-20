{{- if not $.IsFish }}
_{{ $.Executable }}() {
{{- end }}
  if [ -z "{{ $.DefaultCompletion }}" ] || [ "{{ $.DefaultCompletion }}" != "{{ $.IsYes }}" ]{{ if not $.IsFish }}; then{{ end }}
    {{- range $idx, $prof := $.Profiles }}
    {{- if not $prof.IsDefault }}
    if {{ $prof.Conditional }}{{ if not $.IsFish }}; then {{ end }}
      {{ $prof.Name }}
      return
    {{ if $.IsFish }}end{{ else }}fi{{ end }}
    {{- end }}
    {{- end }}
  {{ if $.IsFish }}end{{ else }}fi{{ end }}
  {{ $.DefaultProfile.Name }}
{{- if not $.IsFish }}
}
{{- end }}
