# {{ $.Executable }} completion

_{{ $.Executable }}() {
  local cur opts chosen found
  source <({{ $.ExportCommand }}) 
  cur=${COMP_WORDS[COMP_CWORD]}
  if [ "$COMP_CWORD" -eq 1 ]; then
{{- range $idx, $value := $.Options }}
    if {{ $value.Conditional }}; then
      opts="${opts}{{ $value.Key }} "
    fi
{{- end}}
    # shellcheck disable=SC2207
    COMPREPLY=( $(compgen -W "$opts" -- "$cur") )
  else
    if [ "$COMP_CWORD" -lt 2 ]; then
      return
    fi
    chosen=${COMP_WORDS[1]}
    found=0
{{- range $idx, $value := $.Options }}
    if {{ $value.Conditional }}; then
      if [ "$chosen" == "{{ $value.Key }}" ]; then
        found=1
      fi
    fi
{{- end}}
    if [ "$found" -eq 0 ]; then
      return
    fi
    if [ "$COMP_CWORD" -eq 2 ]; then
      case "$chosen" in
        "{{ $.HelpCommand }}")
          opts="{{ $.HelpAdvancedCommand }} {{ $.HelpConfigCommand }}"
          ;;
        "{{ $.InsertCommand }}" | "{{ $.MultiLineCommand }}" | "{{ $.MoveCommand }}" | "{{ $.RemoveCommand }}")
          if {{ $.Conditionals.Not.AskMode }}; then
            opts="$opts $({{ $.DoList }})"
          fi
          ;;
        "{{ $.TOTPCommand }}")
          opts="{{ $.TOTPListCommand }} "
          opts="{{ $.TOTPFindCommand }} "
{{- range $key, $value := .TOTPSubCommands }}
          if {{ $value.Conditional }}; then
            opts="$opts {{ $value.Key }}"
          fi
{{- end}}
          ;;
        "{{ $.ShowCommand }}" | "{{ $.JSONCommand }}" | "{{ $.ClipCommand }}")
          if {{ $.Conditionals.Not.AskMode }}; then
            opts=$({{ $.DoList }})
          fi
          ;;
      esac
    else
      if [ "$COMP_CWORD" -eq 3 ]; then
        case "$chosen" in
          "{{ $.MoveCommand }}")
            if {{ $.Conditionals.Not.AskMode }}; then
              opts=$({{ $.DoList }})
            fi
            ;;
          "{{ $.TOTPCommand }}")
            case "${COMP_WORDS[2]}" in
{{- range $key, $value := $.TOTPSubCommands }}
              "{{ $value.Key }}")
                if {{ $value.Conditional }}; then
                  if {{ $.Conditionals.Not.AskMode }}; then
                    opts=$({{ $.DoTOTPList }})
                  fi
                fi
                ;;
{{- end}}
            esac
            ;;
        esac
      fi
    fi
    if [ -n "$opts" ]; then
      # shellcheck disable=SC2207
      COMPREPLY=($(compgen -W "$opts" -- "$cur"))
    fi
  fi
}

complete -F _{{ $.Executable }} -o bashdefault {{ $.Executable }}
