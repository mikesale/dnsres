#compdef dnsres dnsres-tui

# zsh completion for dnsres and dnsres-tui

_dnsres_flags() {
    local -a flags
    flags=(
        '-config[Path to configuration file]:config file:_files -g "*.json"'
        '-host[Hostname to resolve (overrides config)]:hostname:'
        '-help[Show help message]'
        '-version[Show version information]'
    )
    
    # Add -report flag only for dnsres (not dnsres-tui)
    if [[ ${words[1]} == "dnsres" ]]; then
        flags+=('-report[Print statistics report and exit]')
    fi
    
    _arguments -s -S $flags '*:hostname:'
}

_dnsres() {
    _dnsres_flags
}

_dnsres_tui() {
    _dnsres_flags
}

# Register both completions
compdef _dnsres dnsres
compdef _dnsres_tui dnsres-tui
