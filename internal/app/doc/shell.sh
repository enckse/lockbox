_{{ $.Executable }}() {
  if [ -z "{{ $.DefaultCompletion }}" ] || [ "{{ $.DefaultCompletion }}" != "{{ $.IsYes }}" ]; then
    {{- range $idx, $prof := $.Profiles }}
    {{- if not $prof.IsDefault }}
      if {{ $prof.Conditional }}; then
        {{ $prof.Name }}
        return
      fi
    {{- end }}
    {{- end }}
  fi
  {{ $.DefaultProfile.Name }}
}
