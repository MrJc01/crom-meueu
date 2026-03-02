#!/bin/bash
set -e

# Creates a JSON map of all JS files in frontend/js/ mapping to their SHA256 hashes
MANIFEST_FILE="manifest.json"

echo "{" > "$MANIFEST_FILE"
echo '  "files": {' >> "$MANIFEST_FILE"
first=true

# Find all JS files in frontend/ and index.html
find frontend -type f \( -name "*.js" -o -name "*.html" \) | sort | while read -r filepath; do
    hash=$(sha256sum "$filepath" | awk '{print $1}')
    
    # Strip the leading 'frontend' part to make it match checker.js expectations
    # e.g. frontend/js/api.js -> /js/api.js
    relative_path="/${filepath#frontend/}"
    
    if [ "$first" = true ]; then
        first=false
    else
        echo "," >> "$MANIFEST_FILE"
    fi
    printf '    "%s": "%s"' "$relative_path" "$hash" >> "$MANIFEST_FILE"
done

echo "" >> "$MANIFEST_FILE"
echo "  }" >> "$MANIFEST_FILE"
echo "}" >> "$MANIFEST_FILE"

echo "Manifest $MANIFEST_FILE generated successfully!"
