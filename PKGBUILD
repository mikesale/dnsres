# Maintainer: Mike Sale <mike.sale@gmail.com>
pkgname=dnsres-bin
pkgver=1.1.8
pkgrel=1
pkgdesc="DNS resolution monitoring tool with health checks, metrics, and TUI"
arch=('x86_64' 'aarch64')
url="https://github.com/mikesale/dnsres"
license=('GPL3')
provides=('dnsres')
conflicts=('dnsres')
source_x86_64=("https://github.com/mikesale/dnsres/releases/download/v${pkgver}/dnsres_${pkgver}_Linux_x86_64.tar.gz")
source_aarch64=("https://github.com/mikesale/dnsres/releases/download/v${pkgver}/dnsres_${pkgver}_Linux_arm64.tar.gz")
sha256sums_x86_64=('SKIP')  # Updated by maintainer for each release
sha256sums_aarch64=('SKIP')  # Updated by maintainer for each release

package() {
    # Install binaries
    install -Dm755 "${srcdir}/dnsres" "${pkgdir}/usr/local/bin/dnsres"
    install -Dm755 "${srcdir}/dnsres-tui" "${pkgdir}/usr/local/bin/dnsres-tui"
    
    # Install example configuration
    install -Dm644 "${srcdir}/examples/config.json" "${pkgdir}/etc/dnsres/config.json.example"
    
    # Install documentation
    install -Dm644 "${srcdir}/README.md" "${pkgdir}/usr/share/doc/${pkgname}/README.md"
    install -Dm644 "${srcdir}/INSTALL.md" "${pkgdir}/usr/share/doc/${pkgname}/INSTALL.md"
    install -Dm644 "${srcdir}/LICENSE" "${pkgdir}/usr/share/licenses/${pkgname}/LICENSE"
    
    # Install shell completions
    install -Dm644 "${srcdir}/completions/dnsres.bash" "${pkgdir}/usr/share/bash-completion/completions/dnsres"
    install -Dm644 "${srcdir}/completions/dnsres.bash" "${pkgdir}/usr/share/bash-completion/completions/dnsres-tui"
    install -Dm644 "${srcdir}/completions/dnsres.zsh" "${pkgdir}/usr/share/zsh/site-functions/_dnsres"
    install -Dm644 "${srcdir}/completions/dnsres.zsh" "${pkgdir}/usr/share/zsh/site-functions/_dnsres-tui"
    install -Dm644 "${srcdir}/completions/dnsres.fish" "${pkgdir}/usr/share/fish/vendor_completions.d/dnsres.fish"
    install -Dm644 "${srcdir}/completions/dnsres.fish" "${pkgdir}/usr/share/fish/vendor_completions.d/dnsres-tui.fish"
}
