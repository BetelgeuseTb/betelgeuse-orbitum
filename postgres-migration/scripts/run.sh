#!/bin/bash
set -e

echo "Starting migrations..."

for dir in /app/migrations/*/; do
  if [ -d "$dir" ]; then
    echo "Processing $dir"
    for file in "$dir"/*.sql; do
      if [[ "$(basename "$file")" =~ ^0*([0-9]+)_ ]]; then
        num="${BASH_REMATCH[1]}"
        newname="$dir/$num${file#$dir/0*_*}"
        mv "$file" "$newname"
        echo "Renamed $file → $newname"
      fi
    done
  fi
done
        
echo "All migrations completed successfully"
