#!/usr/bin/env bash
set -e

input=$1
pkg=$2

from=$(grep -m1 '^package' "$input" | awk '{print $2}')

perl -ne 'print "$1\n" if /^const \(/.../^\)/ and /^\s+([A-Z]\w+)/' . "$input" | sort | {
    echo "package $pkg"
    echo "// @generated"
    echo
    echo "// Constants copied from $input"
    echo "const ("
    while read -r name; do
        echo "  $name = $from.$name"
    done
    echo ")"
}
