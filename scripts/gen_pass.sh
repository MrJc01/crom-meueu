#!/bin/bash
# Script to generate SHA256 hash for Admin Password
# Usage: ./gen_pass.sh "my_password"

if [ -z "$1" ]; then
  echo "Usage: $0 <password>"
  exit 1
fi

echo -n "$1" | sha256sum | awk '{print $1}'
