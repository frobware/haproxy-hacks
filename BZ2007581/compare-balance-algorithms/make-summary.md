#!/usr/bin/env bash

set -eu

cat <<EOF > README.md
# HAProxy Load times
EOF

for i in benchmark-result-*.md; do
    a=${i/benchmark-result-/}
    cat <<EOF >> README.md

## Configuration: $(basename $a .md)
EOF
    cat $i >> README.md
done



    
