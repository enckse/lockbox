# bash completion for lb                        -*- shell-script -*-

_lb() {
    local cur opts
    cur=${COMP_WORDS[COMP_CWORD]}
    if [ $COMP_CWORD -eq 1 ]; then
        opts="ls clip show -c insert rm rekey totp list pwgen stats"
        COMPREPLY=( $(compgen -W "$opts" -- $cur) )
    else
        if [ $COMP_CWORD -eq 2 ]; then
            case ${COMP_WORDS[1]} in
                "insert")
                    opts="-m $(lb ls)"
                    ;;
                "totp")
                    opts=$(lb totp ls)
                    ;;
                "pwgen")
                    opts="-length -transform -special"
                    ;;
                "-c" | "show" | "rm" | "clip" | "stats")
                    opts=$(lb ls)
                    ;;
            esac
        fi
        if [ $COMP_CWORD -eq 3 ] && [ "${COMP_WORDS[1]}" == "insert" ]; then
            if [ "${COMP_WORDS[2]}" == "-m" ]; then
                opts=$(lb ls)
            else
                opts="-m"
            fi
        fi
        if [ ! -z "$opts" ]; then
            COMPREPLY=( $(compgen -W "$opts" -- $cur) )
        fi
    fi
}

complete -F _lb -o bashdefault -o default lb
