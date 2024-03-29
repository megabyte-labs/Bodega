#!/bin/sh

# @description This script will install Bodega. The script heavily borrows from
#   [dep's install.sh script](https://github.com/golang/dep/blob/master/install.sh) but is adapted
#   to serve as a generic GitHub release installer instead of a Go GitHub release installer. It
#   identifies most platforms and throws an error if the platform is not supported.
#
#   ## Environment variables:
#
#   * INSTALL_DIRECTORY (optional): defaults to $HOME/.local/bin
#   * DEP_RELEASE_TAG (optional): defaults to fetching the latest release
#
#   ## Install Command
#
#   ```
#   curl -sSL https://raw.githubusercontent.com/megabyte-labs/bodega/master/local/install.sh | sh
#   ```

RELEASES_URL="https://github.com/megabyte-labs/Bodega/releases"

downloadJSON() {
    url="$2"
    echo "Fetching $url.."
    if test -x "$(command -v curl)"; then
        response=$(curl -s -L -w 'HTTPSTATUS:%{http_code}' -H 'Accept: application/json' "$url")
        body=$(echo "$response" | sed -e 's/HTTPSTATUS\:.*//g')
        code=$(echo "$response" | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
    elif test -x "$(command -v wget)"; then
        temp=$(mktemp)
        body=$(wget -q --header='Accept: application/json' -O - --server-response "$url" 2> "$temp")
        code=$(awk '/^  HTTP/{print $2}' < "$temp" | tail -1)
        rm "$temp"
    else
        echo "Neither curl nor wget was available to perform http requests."
        exit 1
    fi
    if [ "$code" != 200 ]; then
        echo "Request failed with code $code"
        exit 1
    fi
    eval "$1='$body'"
}

downloadFile() {
    url="$1"
    destination="$2"
    echo "Fetching $url.."
    if test -x "$(command -v curl)"; then
        code=$(curl -s -w '%{http_code}' -L "$url" -o "$destination")
    elif test -x "$(command -v wget)"; then
        code=$(wget -q -O "$destination" --server-response "$url" 2>&1 | awk '/^  HTTP/{print $2}' | tail -1)
    else
        echo "Neither curl nor wget was available to perform http requests."
        exit 1
    fi
    if [ "$code" != 200 ]; then
        echo "Request failed with code $code"
        exit 1
    fi
}

findGoBinDirectory() {
    EFFECTIVE_GOPATH=$(go env GOPATH)
    # CYGWIN: Convert Windows-style path into sh-compatible path
    if [ "$OS_CYGWIN" = "1" ]; then
	EFFECTIVE_GOPATH=$(cygpath "$EFFECTIVE_GOPATH")
    fi
    if [ -z "$EFFECTIVE_GOPATH" ]; then
        echo "Installation could not determine your \$GOPATH."
        exit 1
    fi
    if [ -z "$GOBIN" ]; then
        GOBIN=$(echo "${EFFECTIVE_GOPATH%%:*}/bin" | sed s#//*#/#g)
    fi
    if [ ! -d "$GOBIN" ]; then
        echo "Installation requires your GOBIN directory $GOBIN to exist. Please create it."
        exit 1
    fi
    eval "$1='$GOBIN'"
}

initArch() {
    ARCH=$(uname -m)
    if [ -n "$DEP_ARCH" ]; then
        echo "Using DEP_ARCH"
        ARCH="$DEP_ARCH"
    fi
    case $ARCH in
        amd64) ARCH="amd64";;
        x86_64) ARCH="amd64";;
        i386) ARCH="386";;
        ppc64) ARCH="ppc64";;
        ppc64le) ARCH="ppc64le";;
        s390x) ARCH="s390x";;
        armv6*) ARCH="arm";;
        armv7*) ARCH="arm";;
        aarch64) ARCH="arm64";;
        *) echo "Architecture ${ARCH} is not supported by this installation script"; exit 1;;
    esac
    echo "ARCH = $ARCH"
}

initOS() {
    OS=$(uname | tr '[:upper:]' '[:lower:]')
    OS_CYGWIN=0
    if [ -n "$DEP_OS" ]; then
        echo "Using DEP_OS"
        OS="$DEP_OS"
    fi
    case "$OS" in
        darwin) OS='darwin';;
        linux) OS='linux';;
        freebsd) OS='freebsd';;
        mingw*) OS='windows';;
        msys*) OS='windows';;
	cygwin*)
	    OS='windows'
	    OS_CYGWIN=1
	    ;;
        *) echo "OS ${OS} is not supported by this installation script"; exit 1;;
    esac
    echo "OS = $OS"
}

echo "Acquiring platform details"
initArch
initOS

if [ -z "$INSTALL_DIRECTORY" ]; then
    INSTALL_DIRECTORY="$HOME/.local/bin"
    mkdir -p "$HOME/.local/bin"
fi
echo "Setting install directory to ---> $INSTALL_DIRECTORY"

echo "Ensuring platform is supported"
if [ "${OS}" != "linux" ] && { [ "${ARCH}" = "ppc64" ] || [ "${ARCH}" = "ppc64le" ];}; then
  echo "${OS}-${ARCH} is not supported by this instalation script" && exit 1
else
  BINARY="task-${OS}-${ARCH}"
  if [ "$OS" = "windows" ]; then
    BINARY="$BINARY.exe"
  fi
fi

echo "Ensuring desired tag is downloaded"
if [ -z "$DEP_RELEASE_TAG" ]; then
    downloadJSON LATEST_RELEASE "$RELEASES_URL/latest"
    DEP_RELEASE_TAG=$(echo "${LATEST_RELEASE}" | tr -s '\n' ' ' | sed 's/.*"tag_name":"//' | sed 's/".*//' )
fi

echo "Ensuring release exists"
downloadJSON RELEASE_DATA "$RELEASES_URL/tag/$DEP_RELEASE_TAG"

echo "Downloading binary"
BINARY_URL="$RELEASES_URL/download/$DEP_RELEASE_TAG/$BINARY"
DOWNLOAD_FILE=$(mktemp)
downloadFile "$BINARY_URL" "$DOWNLOAD_FILE"

echo "Setting executable permissions"
chmod +x "$DOWNLOAD_FILE"

echo "Moving executable to $INSTALL_DIRECTORY/$INSTALL_NAME"
INSTALL_NAME="task"
if [ "$OS" = "windows" ]; then
    INSTALL_NAME="$INSTALL_NAME.exe"
fi
mv "$DOWNLOAD_FILE" "$INSTALL_DIRECTORY/$INSTALL_NAME"

echo "All done! Make sure the executable is included in your PATH environment variable."
