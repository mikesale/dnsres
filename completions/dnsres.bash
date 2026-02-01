# bash completion for dnsres                               -*- shell-script -*-

_dnsres()
{
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    opts="-config -host -report -help -version"

    case "${prev}" in
        -config)
            # Complete file paths for config flag
            COMPREPLY=( $(compgen -f -X '!*.json' -- "${cur}") )
            return 0
            ;;
        -host)
            # No automatic completion for hostname
            return 0
            ;;
        *)
            ;;
    esac

    if [[ ${cur} == -* ]] ; then
        COMPREPLY=( $(compgen -W "${opts}" -- "${cur}") )
        return 0
    fi

    # Complete hostnames as positional arguments
    return 0
}

complete -F _dnsres dnsres

# bash completion for dnsres-tui
_dnsres_tui()
{
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    opts="-config -host -help -version"

    case "${prev}" in
        -config)
            # Complete file paths for config flag
            COMPREPLY=( $(compgen -f -X '!*.json' -- "${cur}") )
            return 0
            ;;
        -host)
            # No automatic completion for hostname
            return 0
            ;;
        *)
            ;;
    esac

    if [[ ${cur} == -* ]] ; then
        COMPREPLY=( $(compgen -W "${opts}" -- "${cur}") )
        return 0
    fi

    # Complete hostnames as positional arguments
    return 0
}

complete -F _dnsres_tui dnsres-tui
