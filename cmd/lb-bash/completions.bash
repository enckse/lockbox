# bash completion for lb                        -*- shell-script -*-

_is_clip() {
    if [ "$1" == "-c" ] || [ "$1" == "{2}clip" ]; then
        echo 1
    else
        echo 0
    fi
}

_lb() {
    local cur opts clip_enabled needs
    clip_enabled=" -c clip"
    if [ -n "$LOCKBOX_CLIPMODE" ]; then
        if [ "$LOCKBOX_CLIPMODE" == "off" ]; then
            clip_enabled=""
        fi
    fi
    cur=${COMP_WORDS[COMP_CWORD]}
    if [ "$COMP_CWORD" -eq 1 ]; then
        opts="version ls show insert rm rekey totp list pwgen find$clip_enabled"
        # shellcheck disable=SC2207
        COMPREPLY=( $(compgen -W "$opts" -- "$cur") )
    else
        if [ "$COMP_CWORD" -eq 2 ]; then
            case ${COMP_WORDS[1]} in
                "insert")
                    opts="-m $(lb ls)"
                    ;;
                "totp")
                    opts="-once -short "$(lb totp -ls)
                    if [ -n "$clip_enabled" ]; then
                        opts="$opts -c -clip"
                    fi
                    ;;
                "pwgen")
                    opts="-length -transform -special"
                    ;;
                "-c" | "show" | "rm" | "clip")
                    opts=$(lb ls)
                    if [ $(_is_clip "${COMP_WORDS[1]}" "") == 1 ]; then 
                        if [ -z "$clip_enabled" ]; then
                            opts=""
                        fi
                    fi
                    ;;
            esac
        fi
        if [ "$COMP_CWORD" -eq 3 ]; then
            case "${COMP_WORDS[1]}" in
                "insert")
                    if [ "${COMP_WORDS[2]}" == "-m" ]; then
                        opts=$(lb ls)
                    else
                        opts="-m"
                    fi
                    ;;
                "totp")
                    needs=0
                    if [ "${COMP_WORDS[2]}" == "-once" ] || [ "${COMP_WORDS[2]}" == "-short" ]; then
                        needs=1
                    else
                        if [ -n "$clip_enabled" ]; then
                            if [ $(_is_clip "${COMP_WORDS[2]}" "-") == 1 ]; then 
                                needs=1
                            fi
                        fi
                    fi
                    if [ $needs -eq 1 ]; then
                        opts=$(lb totp -ls)
                    fi
                    ;;
            esac
        fi
        if [ -n "$opts" ]; then
            # shellcheck disable=SC2207
            COMPREPLY=($(compgen -W "$opts" -- "$cur"))
        fi
    fi
}

complete -F _lb -o bashdefault -o default lb
