# bash completion for lb                        -*- shell-script -*-

_is_clip() {
    if [ "$1" == "${2}clip" ]; then
        echo 1
    else
        echo 0
    fi
}

_lb() {
    local cur opts clip_enabled needs readwrite
    clip_enabled=" clip"
    if [ -n "$LOCKBOX_NOCLIP" ]; then
        if [ "$LOCKBOX_NOCLIP" == "yes" ]; then
            clip_enabled=""
        fi
    fi
    readwrite=" insert rm mv"
    if [ -n "$LOCKBOX_READONLY" ]; then
        if [ "$LOCKBOX_READONLY" == "yes" ]; then
            readwrite=""
        fi
    fi
    cur=${COMP_WORDS[COMP_CWORD]}
    if [ "$COMP_CWORD" -eq 1 ]; then
        opts="version help ls show env totp$readwrite find$clip_enabled"
        # shellcheck disable=SC2207
        COMPREPLY=( $(compgen -W "$opts" -- "$cur") )
    else
        if [ "$COMP_CWORD" -eq 2 ]; then
            case ${COMP_WORDS[1]} in
                "insert")
                    opts="-multi $(lb ls)"
                    ;;
                "mv")
                    opts=$(lb ls)
                    ;;
                "totp")
                    opts="-once -short "$(lb totp -list)
                    if [ -n "$clip_enabled" ]; then
                        opts="$opts -clip"
                    fi
                    ;;
                "show" | "rm" | "clip")
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
                    if [ "${COMP_WORDS[2]}" == "-multi" ]; then
                        opts=$(lb ls)
                    fi
                    ;;
                "mv")
                    opts=$(lb ls)
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
                        opts=$(lb totp -list)
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
