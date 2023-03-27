# {{ $.Executable }} completion

_{{ $.Executable }}() {
    local cur opts needs
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
                    opts="{{ $.InsertMultiCommand }}{{ if $.CanTOTP }} {{ $.InsertTOTPCommand }}{{end}} $({{ $.DoList }})"
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
                    opts="{{ $.TOTPShortCommand }} {{ $.TOTPOnceCommand }} {{ $.TOTPListCommand }} "$({{ $.DoTOTPList }})
{{ if $.CanClip }}
                    opts="$opts {{ $.TOTPClipCommand }}"
{{end}}
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
                    if [ "${COMP_WORDS[2]}" == "{{ $.InsertMultiCommand }}" ] {{ if $.CanTOTP }}|| [ "${COMP_WORDS[2]}" == "{{ $.InsertTOTPCommand }}" ] {{end}}; then
                        opts=$({{ $.DoList }})
                    fi
                    ;;
                "{{ $.MoveCommand }}")
                    opts=$({{ $.DoList }})
                    ;;
{{end}}
{{ if $.CanTOTP }}
                "{{ $.TOTPCommand }}")
                    needs=0
                    if [ "${COMP_WORDS[2]}" == "{{ $.TOTPOnceCommand }}" ] || [ "${COMP_WORDS[2]}" == "{{ $.TOTPShortCommand }}" ]; then
                        needs=1
{{ if $.CanClip }}
                    else
                        if [ "${COMP_WORDS[2]}" == "{{ $.TOTPClipCommand }}" ]; then
                            needs=1
                        fi
{{end}}
                    fi
                    if [ $needs -eq 1 ]; then
                        opts=$({{ $.DoTOTPList }})
                    fi
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
