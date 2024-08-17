#!/bin/bash

# List sqlite DB backups at AWS S3: deploy/scripts/backup-list.sh

CONFIG_FILEPATH="$HOME/deploy/.config"

if [ ! -f "$CONFIG_FILEPATH" ]; then
  echo "Error: Config file $CONFIG_FILEPATH not found."
  exit 1
fi

BUCKET_NAME=$(grep BUCKET_NAME $CONFIG_FILEPATH | cut -d'=' -f2 | tr -d '"*')
BACKUP_PREFIX="n8n-shortlink-backups/"

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

aws s3 ls "s3://$BUCKET_NAME/$BACKUP_PREFIX" | sort -r | while read -r line; do
  date=$(echo $line | awk '{print $1}')
  time=$(echo $line | awk '{print $2}')
  size=$(echo $line | awk '{print $3}')
  filename=$(echo $line | awk '{print $4}')

  hr_size=$(human_readable_size $size)

  short_filename=${filename#$BACKUP_PREFIX} # remove prefix

  printf "%-12s %-12s %-12s %s\n" "$date" "$time" "$hr_size" "$short_filename"
done
