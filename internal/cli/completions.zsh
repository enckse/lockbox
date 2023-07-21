#compdef _{{ $.Executable }} {{ $.Executable }}
 
_{{ $.Executable }}() { 
  local curcontext="$curcontext" state len
  typeset -A opt_args

  _arguments \
    '1: :->main'\
    '*: :->args'

  len=${#words[@]}
  case $state in
    main)
      _arguments '1:main:({{ range $idx, $value := $.Options }}{{ if gt $idx 0}} {{ end }}{{ $value }}{{ end }})'
    ;;
    *)
      case $words[2] in
        "{{ $.HelpCommand }}")
          if [ "$len" -eq 3 ]; then
            compadd "$@" "{{ $.HelpAdvancedCommand }}"
          fi
        ;;
{{- if not $.ReadOnly }}
        "{{ $.InsertCommand }}" | "{{ $.MultiLineCommand }}")
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
{{- if $.CanTOTP }}
        "{{ $.TOTPCommand }}")
          case "$len" in
            3)
              compadd "$@" {{ $.TOTPListCommand }}{{ range $key, $value := .TOTPSubCommands }} {{ $value }}{{ end }}
            ;;
            4)
              case $words[3] in
{{- range $key, $value := .TOTPSubCommands }}
                "{{ $value }}")
                  compadd "$@" $({{ $.DoTOTPList }})
                ;;
{{- end}}
              esac
          esac
        ;;
{{- end}}
        "{{ $.ShowCommand }}" | "{{ $.JSONCommand }}" {{ if not $.ReadOnly }}| "{{ $.RemoveCommand }}" {{end}} {{ if $.CanClip }} | "{{ $.ClipCommand }}" {{end}})
          if [ "$len" -eq 3 ]; then
            compadd "$@" $({{ $.DoList }})
          fi
        ;;
      esac
  esac
}
