#!/bin/bash

# List sqlite DB backups at AWS S3
: '
~/.n8n-shortlink/backup/backup-list.sh
'

BUCKET_NAME=$(grep bucket_name ~/.aws/config | cut -d '=' -f2 | tr -d ' ')

if [ -z "$BUCKET_NAME" ]; then
  echo "Error: Bucket name not found in ~/.aws/config"
  exit 1
fi

human_readable_size() {
  local size=$1
  local units=("B" "KiB" "MiB" "GiB" "TiB")
  local unit=0

  while (( $(echo "$size > 1024" | bc -l) )); do
    size=$(echo "scale=2; $size / 1024" | bc -l)
    ((unit++))
  done

  printf "%.2f %s" $size "${units[$unit]}"
}

aws s3 ls "s3://$BUCKET_NAME/" | sort -r | while read -r line; do
  date=$(echo $line | awk '{print $1}')
  time=$(echo $line | awk '{print $2}')
  size=$(echo $line | awk '{print $3}')
  filename=$(echo $line | awk '{print $4}')

  hr_size=$(human_readable_size $size)

  printf "%-12s %-12s %-12s %s\n" "$date" "$time" "$hr_size" "$filename"
done
