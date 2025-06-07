#compdef _{{ $.Executable }} {{ $.Executable }}

_{{ $.Executable }}() {
  local curcontext="$curcontext" state len chosen found args
  typeset -A opt_args
  source <({{ $.ExportCommand }}) 
  _arguments \
    '1: :->main'\
    '*: :->args'

  len=${#words[@]}
  case $state in
    main)
      args=""
{{- range $idx, $value := $.Options }}
      if {{ $value.Conditional }}; then
        if [ -n "$args" ]; then
          args="$args "
        fi
        args="${args}{{ $value.Key }}"
      fi
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
      if {{ $value.Conditional }}; then
        if [[ "$chosen" == "{{ $value.Key }}" ]]; then
          found=1
        fi
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
            if {{ $.Conditionals.Not.AskMode }}; then
              compadd "$@" $({{ $.DoGroups }})
            fi
          fi
        ;;
        "{{ $.InsertCommand }}" | "{{ $.UnsetCommand }}")
          if [ "$len" -eq 3 ]; then
            if {{ $.Conditionals.Not.AskMode }}; then
              compadd "$@" $({{ $.DoList }})
            fi
          fi
        ;;
        "{{ $.MoveCommand }}")
          case "$len" in
            3 | 4)
              if {{ $.Conditionals.Not.AskMode }}; then
                compadd "$@" $({{ $.DoGroups }})
              fi
            ;;
          esac
        ;;
        "{{ $.TOTPCommand }}")
          case "$len" in
            3)
              compadd "$@" {{ $.TOTPListCommand }}
              compadd "$@" {{ $.TOTPFindCommand }}
{{- range $key, $value := .TOTPSubCommands }}
              if {{ $value.Conditional }}; then
                compadd "$@" {{ $value.Key }}
              fi
{{ end }}
            ;;
            4)
              case $words[3] in
{{- range $key, $value := .TOTPSubCommands }}
                "{{ $value.Key }}")
                  if {{ $value.Conditional }}; then
                    if {{ $.Conditionals.Not.AskMode }}; then
                      compadd "$@" $({{ $.DoTOTPList }})
                    fi
                  fi
                ;;
{{- end}}
              esac
          esac
        ;;
        "{{ $.ShowCommand }}" | "{{ $.JSONCommand }}" | "{{ $.ClipCommand }}")
          if [ "$len" -eq 3 ]; then
            if {{ $.Conditionals.Not.AskMode }}; then
              compadd "$@" $({{ $.DoList }})
            fi
          fi
        ;;
      esac
  esac
}

compdef _{{ $.Executable }} {{ $.Executable }}
