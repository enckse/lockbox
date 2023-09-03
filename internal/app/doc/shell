_{{ $.Executable }}() {
{{- if eq (len $.Profiles) 1 }}
  {{ $.DefaultProfile.Name }}
{{- else}}
  case "{{ $.CompletionEnv }}" in
{{- range $idx, $profile := $.Profiles }}
{{- if not $profile.IsDefault }}
    "{{ $profile.Display }}")
      {{ $profile.Name }}
      ;;
{{- end}}
{{- end}}
    *)
      {{ $.DefaultProfile.Name }} 
      ;;
  esac
{{- end}}
}
