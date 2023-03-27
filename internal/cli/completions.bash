# {{ $.Executable }} completion

_{{ $.Executable }}() {
    local cur opts
    cur=${COMP_WORDS[COMP_CWORD]}
    if [ "$COMP_CWORD" -eq 1 ]; then
        {{range $idx, $value := $.Options }}
        opts="${opts}{{ $value }} "{{end}}
        # shellcheck disable=SC2207
        COMPREPLY=( $(compgen -W "$opts" -- "$cur") )
    else
        if [ "$COMP_CWORD" -eq 2 ]; then
            case ${COMP_WORDS[1]} in
{{ if not $.ReadOnly }}
                "{{ $.InsertCommand }}")
{{ range $key, $value := .InsertSubCommands }}
                    opts="$opts {{ $value }}"
{{end}}
                    opts="$opts $({{ $.DoList }})"
                    ;;
                "{{ $.HelpCommand }}")
                    opts="{{ $.HelpAdvancedCommand }}"
                    ;;
                "{{ $.MoveCommand }}")
                    opts=$({{ $.DoList }})
                    ;;
{{end}}
{{ if $.CanTOTP }}
                "{{ $.TOTPCommand }}")
                    opts="{{ $.TOTPListCommand }} "
{{ range $key, $value := .TOTPSubCommands }}
                    opts="$opts {{ $value }}"
{{end}}
                    opts="$opts "$({{ $.DoTOTPList }})
                    ;;
{{end}}
                "{{ $.ShowCommand }}" | "{{ $.StatsCommand }}" {{ if not $.ReadOnly }}| "{{ $.RemoveCommand }}" {{end}} {{ if $.CanClip }} | "{{ $.ClipCommand }}" {{end}})
                    opts=$({{ $.DoList }})
                    ;;
            esac
        fi
        if [ "$COMP_CWORD" -eq 3 ]; then
            case "${COMP_WORDS[1]}" in
{{ if not $.ReadOnly }}
                "{{ $.InsertCommand }}")
                    case "${COMP_WORDS[2]}" in
{{ range $key, $value := .InsertSubCommands }}
                      "{{ $value }}")
                      opts=$({{ $.DoList }})
                      ;;
{{end}}
                    esac
                    ;;
                "{{ $.MoveCommand }}")
                    opts=$({{ $.DoList }})
                    ;;
{{end}}
{{ if $.CanTOTP }}
                "{{ $.TOTPCommand }}")
                    case "${COMP_WORDS[2]}" in
{{ range $key, $value := .TOTPSubCommands }}
                      "{{ $value }}")
                        opts=$({{ $.DoTOTPList }})
                        ;;
{{end}}
                    esac
                    ;;
{{end}}
            esac
        fi
        if [ -n "$opts" ]; then
            # shellcheck disable=SC2207
            COMPREPLY=($(compgen -W "$opts" -- "$cur"))
        fi
    fi
}

complete -F _{{ $.Executable }} -o bashdefault -o default {{ $.Executable }}
