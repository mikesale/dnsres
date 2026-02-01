# fish completion for dnsres and dnsres-tui

# dnsres completions
complete -c dnsres -s c -l config -d 'Path to configuration file' -r -F
complete -c dnsres -s h -l host -d 'Hostname to resolve (overrides config)' -r
complete -c dnsres -s r -l report -d 'Print statistics report and exit'
complete -c dnsres -l help -d 'Show help message'
complete -c dnsres -l version -d 'Show version information'

# Short flag versions (matching Go flag package behavior)
complete -c dnsres -o config -d 'Path to configuration file' -r -F
complete -c dnsres -o host -d 'Hostname to resolve (overrides config)' -r
complete -c dnsres -o report -d 'Print statistics report and exit'
complete -c dnsres -o help -d 'Show help message'
complete -c dnsres -o version -d 'Show version information'

# dnsres-tui completions
complete -c dnsres-tui -s c -l config -d 'Path to configuration file' -r -F
complete -c dnsres-tui -s h -l host -d 'Hostname to resolve (overrides config)' -r
complete -c dnsres-tui -l help -d 'Show help message'
complete -c dnsres-tui -l version -d 'Show version information'

# Short flag versions (matching Go flag package behavior)
complete -c dnsres-tui -o config -d 'Path to configuration file' -r -F
complete -c dnsres-tui -o host -d 'Hostname to resolve (overrides config)' -r
complete -c dnsres-tui -o help -d 'Show help message'
complete -c dnsres-tui -o version -d 'Show version information'
