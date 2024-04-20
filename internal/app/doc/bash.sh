# {{ $.Executable }} completion

{{ $.Shell }}

{{- range $idx, $profile := $.Profiles }}

{{ $profile.Name }}() {
  local cur opts
  cur=${COMP_WORDS[COMP_CWORD]}
  if [ "$COMP_CWORD" -eq 1 ]; then
{{- range $idx, $value := $profile.Options }}
    opts="${opts}{{ $value }} "
{{- end}}
    # shellcheck disable=SC2207
    COMPREPLY=( $(compgen -W "$opts" -- "$cur") )
  else
    if [ "$COMP_CWORD" -eq 2 ]; then
      case ${COMP_WORDS[1]} in
        "{{ $.HelpCommand }}")
          opts="{{ $.HelpAdvancedCommand }}"
          ;;
{{- if not $profile.ReadOnly }}
{{- if $profile.CanList }}
        "{{ $.InsertCommand }}" | "{{ $.MultiLineCommand }}" | "{{ $.MoveCommand }}" | "{{ $.RemoveCommand }}")
          opts="$opts $({{ $.DoList }})"
          ;;
{{- end}}
{{- end}}
{{- if $profile.CanTOTP }}
        "{{ $.TOTPCommand }}")
          opts="{{ $.TOTPListCommand }} "
{{- range $key, $value := .TOTPSubCommands }}
          opts="$opts {{ $value }}"
{{- end}}
          ;;
{{- end}}
{{- if $profile.CanList }}
        "{{ $.ShowCommand }}" | "{{ $.JSONCommand }}"{{ if $profile.CanClip }} | "{{ $.ClipCommand }}" {{end}})
          opts=$({{ $.DoList }})
          ;;
{{- end}}
      esac
{{- if $profile.CanList }}
    else
      if [ "$COMP_CWORD" -eq 3 ]; then
        case "${COMP_WORDS[1]}" in
{{- if not $profile.ReadOnly }}
          "{{ $.MoveCommand }}")
            opts=$({{ $.DoList }})
            ;;
{{- end }}
{{- if $profile.CanTOTP }}
          "{{ $.TOTPCommand }}")
            case "${COMP_WORDS[2]}" in
{{- range $key, $value := $profile.TOTPSubCommands }}
              "{{ $value }}")
                opts=$({{ $.DoTOTPList }})
                ;;
{{- end}}
            esac
            ;;
{{- end}}
        esac
      fi
{{- end}}
    fi
    if [ -n "$opts" ]; then
      # shellcheck disable=SC2207
      COMPREPLY=($(compgen -W "$opts" -- "$cur"))
    fi
  fi
}
{{- end}}

complete -F _{{ $.Executable }} -o bashdefault {{ $.Executable }}
