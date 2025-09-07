#compdef _{{ $.Executable }} {{ $.Executable }}

_{{ $.Executable }}() {
  local curcontext="$curcontext" state len chosen found args
  typeset -A opt_args
  _arguments \
    '1: :->main'\
    '*: :->args'

  len=${#words[@]}
  case $state in
    main)
      args=""
{{- range $idx, $value := $.Options }}
      if [ -n "$args" ]; then
        args="$args "
      fi
      args="${args}{{ $value }}"
{{- end }}
      _arguments "1:main:($args)"
    ;;
    *)
      if [ "$len" -lt 2 ]; then
        return
      fi
      chosen=$words[2]
      found=0
{{- range $idx, $value := $.Options }}
      if [[ "$chosen" == "{{ $value }}" ]]; then
        found=1
      fi
{{- end }}
      if [ "$found" -eq 0 ]; then
        return
      fi
      case $chosen in
        "{{ $.HelpCommand }}")
          if [ "$len" -eq 3 ]; then
            compadd "$@" "{{ $.HelpAdvancedCommand }}"
            compadd "$@" "{{ $.HelpConfigCommand }}"
          fi
        ;;
        "{{ $.RemoveCommand }}")
          if [ "$len" -eq 3 ]; then
            compadd "$@" $({{ $.DoGroups }})
          fi
        ;;
        "{{ $.InsertCommand }}")
          if [ "$len" -eq 3 ]; then
            compadd "$@" $({{ $.DoFields }})
          fi
        ;;
        "{{ $.UnsetCommand }}")
          if [ "$len" -eq 3 ]; then
            compadd "$@" $({{ $.DoList }})
          fi
        ;;
        "{{ $.MoveCommand }}")
          case "$len" in
            3 | 4)
              compadd "$@" $({{ $.DoGroups }})
            ;;
          esac
        ;;
        "{{ $.TOTPCommand }}")
          case "$len" in
            3)
              compadd "$@" {{ $.TOTPListCommand }}
{{- range $key, $value := .TOTPSubCommands }}
              compadd "$@" {{ $value }}
{{ end }}
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
        "{{ $.ShowCommand }}" | "{{ $.JSONCommand }}" | "{{ $.ClipCommand }}")
          if [ "$len" -eq 3 ]; then
            compadd "$@" $({{ $.DoList }})
          fi
        ;;
      esac
  esac
}

compdef _{{ $.Executable }} {{ $.Executable }}
