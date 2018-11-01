#!/bin/bash
set -e

echo x y cell_n run_n per_ns

grep -o 'size:.*ns/op' bench.out |
sed -e 's/size:(//' -e 's/)-[0-9]*//' -e 's~ ns/op~~' -e 's/[[:space:]][[:space:]]*/ /g' -e 's/,/ /' |
while read -r x y run_n per_ns; do
    echo "$x" "$y" $(( x * y )) "$run_n" "$per_ns"
done
