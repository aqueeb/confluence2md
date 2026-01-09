#!/bin/bash
# Download Pandoc binaries for all supported platforms
# Usage: ./scripts/download-pandoc.sh [VERSION]

set -e

VERSION="${1:-3.6.4}"
DEST_DIR="internal/pandoc/bin"
TEMP_DIR=$(mktemp -d)

cleanup() {
    rm -rf "$TEMP_DIR"
}
trap cleanup EXIT

echo "Downloading Pandoc $VERSION for all platforms..."

# Create destination directory
mkdir -p "$DEST_DIR"

# Platform configurations: name, url_suffix, archive_type, binary_path_in_archive
declare -A PLATFORMS=(
    ["linux-amd64"]="linux-amd64.tar.gz|tar|pandoc-$VERSION/bin/pandoc"
    ["darwin-amd64"]="x86_64-macOS.zip|zip|pandoc-$VERSION-x86_64/bin/pandoc"
    ["darwin-arm64"]="arm64-macOS.zip|zip|pandoc-$VERSION-arm64/bin/pandoc"
    ["windows-amd64"]="windows-x86_64.zip|zip|pandoc-$VERSION/pandoc.exe"
)

BASE_URL="https://github.com/jgm/pandoc/releases/download/$VERSION"

for platform in "${!PLATFORMS[@]}"; do
    IFS='|' read -r url_suffix archive_type binary_path <<< "${PLATFORMS[$platform]}"

    # Determine output filename
    if [[ "$platform" == "windows-amd64" ]]; then
        output_file="$DEST_DIR/pandoc-$platform.exe"
    else
        output_file="$DEST_DIR/pandoc-$platform"
    fi

    # Skip if already exists
    if [[ -f "$output_file" ]]; then
        echo "[$platform] Already exists: $output_file (skipping)"
        continue
    fi

    url="$BASE_URL/pandoc-$VERSION-$url_suffix"
    archive_file="$TEMP_DIR/pandoc-$platform.$archive_type"

    echo "[$platform] Downloading from $url..."
    curl -fsSL "$url" -o "$archive_file"

    echo "[$platform] Extracting..."
    extract_dir="$TEMP_DIR/extract-$platform"
    mkdir -p "$extract_dir"

    if [[ "$archive_type" == "tar" ]]; then
        tar -xzf "$archive_file" -C "$extract_dir"
    else
        unzip -q "$archive_file" -d "$extract_dir"
    fi

    # Copy binary to destination
    cp "$extract_dir/$binary_path" "$output_file"
    chmod +x "$output_file"

    # Verify
    size=$(stat -c%s "$output_file" 2>/dev/null || stat -f%z "$output_file" 2>/dev/null)
    echo "[$platform] Done: $output_file ($(numfmt --to=iec-i --suffix=B $size 2>/dev/null || echo "${size} bytes"))"
done

echo ""
echo "All Pandoc binaries downloaded to $DEST_DIR/"
ls -lh "$DEST_DIR/"
