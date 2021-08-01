# bash completion for lb                        -*- shell-script -*-

_lb() {
    local cur opts
    cur=${COMP_WORDS[COMP_CWORD]}
    if [ $COMP_CWORD -eq 1 ]; then
        opts="version ls clip show -c insert rm rekey totp list pwgen stats find"
        COMPREPLY=( $(compgen -W "$opts" -- $cur) )
    else
        if [ $COMP_CWORD -eq 2 ]; then
            case ${COMP_WORDS[1]} in
                "insert")
                    opts="-m $(lb ls)"
                    ;;
                "totp")
                    opts="-c clip $(lb totp ls)"
                    ;;
                "pwgen")
                    opts="-length -transform -special"
                    ;;
                "-c" | "show" | "rm" | "clip" | "stats")
                    opts=$(lb ls)
                    ;;
            esac
        fi
        if [ $COMP_CWORD -eq 3 ]; then
            case "${COMP_WORDS[1]}" in
                "insert")
                    if [ "${COMP_WORDS[2]}" == "-m" ]; then
                        opts=$(lb ls)
                    else
                        opts="-m"
                    fi
                    ;;
                "totp")
                    if [ "${COMP_WORDS[2]}" == "-c" ] || [ "${COMP_CWORDS[2]}" == "clip" ]; then
                        opts=$(lb totp ls)
                    fi
                    ;;
            esac
        fi
        if [ ! -z "$opts" ]; then
            COMPREPLY=( $(compgen -W "$opts" -- $cur) )
        fi
    fi
}

complete -F _lb -o bashdefault -o default lb
