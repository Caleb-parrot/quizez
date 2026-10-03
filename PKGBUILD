# Maintainer: Caleb-parrot <288791519+Caleb-parrot@users.noreply.github.com>
pkgname=quizez
pkgver=0.1.0
pkgrel=2
pkgdesc='Grokipedia streak quiz for the terminal'
arch=('x86_64')
url='https://github.com/Caleb-parrot/quizez'
license=('LicenseRef-quizez')
depends=('glibc' 'xdg-terminal-exec')
makedepends=('go')
options=('!debug')

prepare() {
  # pkg/ and src/ sit beside the PKGBUILD. Copy the module so `go test ./...`
  # does not walk them.
  rm -rf "$srcdir/build"
  mkdir -p "$srcdir/build"
  tar -C "$startdir" \
    --exclude=./pkg --exclude=./src --exclude='./*.pkg.tar.*' \
    -cf - . | tar -C "$srcdir/build" -xf -
}

build() {
  cd "$srcdir/build"
  export CGO_CPPFLAGS="${CPPFLAGS}"
  export CGO_CFLAGS="${CFLAGS}"
  export CGO_CXXFLAGS="${CXXFLAGS}"
  export CGO_LDFLAGS="${LDFLAGS}"
  export GOFLAGS="-buildmode=pie -trimpath -ldflags=-linkmode=external -mod=readonly -modcacherw"
  go build -o quizez .
}

check() {
  cd "$srcdir/build"
  export GOFLAGS="-mod=readonly"
  go test ./...
}

package() {
  install -Dm755 "$srcdir/build/quizez" "$pkgdir/usr/bin/quizez"
  install -Dm644 "$startdir/quizez.desktop" "$pkgdir/usr/share/applications/quizez.desktop"
  install -Dm644 "$startdir/LICENSE" "$pkgdir/usr/share/licenses/$pkgname/LICENSE"
}
