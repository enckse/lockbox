#compdef _{{ $.Executable }} {{ $.Executable }}

{{ $.Shell }}

{{- range $idx, $profile := $.Profiles }}

{{ $profile.Name }}() {
  local curcontext="$curcontext" state len
  typeset -A opt_args

  _arguments \
    '1: :->main'\
    '*: :->args'

  len=${#words[@]}
  case $state in
    main)
      _arguments '1:main:({{ range $idx, $value := $profile.Options }}{{ if gt $idx 0}} {{ end }}{{ $value }}{{ end }})'
    ;;
    *)
      case $words[2] in
        "{{ $.HelpCommand }}")
          if [ "$len" -eq 3 ]; then
            compadd "$@" "{{ $.HelpAdvancedCommand }}"
          fi
        ;;
{{- if not $profile.ReadOnly }}
{{- if $profile.CanList }}
        "{{ $.InsertCommand }}" | "{{ $.MultiLineCommand }}" | "{{ $.RemoveCommand }}")
          if [ "$len" -eq 3 ]; then
            compadd "$@" $({{ $.DoList }})
          fi
        ;;
        "{{ $.MoveCommand }}")
          case "$len" in
            3 | 4)
              compadd "$@" $({{ $.DoList }})
            ;;
          esac
        ;;
{{- end}}
{{- end}}
{{- if $profile.CanTOTP }}
        "{{ $.TOTPCommand }}")
          case "$len" in
            3)
              compadd "$@" {{ $.TOTPListCommand }}{{ range $key, $value := .TOTPSubCommands }} {{ $value }}{{ end }}
            ;;
{{- if $profile.CanList }}
            4)
              case $words[3] in
{{- range $key, $value := .TOTPSubCommands }}
                "{{ $value }}")
                  compadd "$@" $({{ $.DoTOTPList }})
                ;;
{{- end}}
              esac
{{- end}}
          esac
        ;;
{{- end}}
{{- if $profile.CanList }}
        "{{ $.ShowCommand }}" | "{{ $.JSONCommand }}"{{ if $profile.CanClip }} | "{{ $.ClipCommand }}" {{end}})
          if [ "$len" -eq 3 ]; then
            compadd "$@" $({{ $.DoList }})
          fi
        ;;
{{- end}}
      esac
  esac
}
{{- end}}
