#!/bin/bash

# Restore a sqlite BB backup from AWS S3: deploy/scripts/backup-restore.sh <backup_name>

if [ $# -eq 0 ]; then
  echo -e "Error: No backup name provided.\nUsage: $0 <backup_name>\nExample: $0 2020-01-03-17:34:35+0200.sql.gz.enc"
  exit 1
fi

CONFIG_FILEPATH="$HOME/deploy/.config"

if [ ! -f "$CONFIG_FILEPATH" ]; then
  echo "Error: Config file $CONFIG_FILEPATH not found."
  exit 1
fi

BUCKET_NAME=$(grep BUCKET_NAME $CONFIG_FILEPATH | cut -d'=' -f2 | tr -d '"*')
BUCKET_PREFIX="n8n-shortlink-backups/"
ENCRYPTION_KEY_PATH="$HOME/.keys/n8n-shortlink-backup-secret.key"
RESTORED_DB=".n8n-shortlink/restored.sqlite"
LOG_FILE="$HOME/deploy/restore.log"
BACKUP_NAME="$1"
BUCKET_PATH="s3://$BUCKET_NAME/$BUCKET_PREFIX$BACKUP_NAME"

bold='\033[1m'
unbold='\033[0m'

rm -f $RESTORED_DB

echo -e "Selected backup: ${bold}$BACKUP_NAME${unbold}"
echo "Downloading backup..."
aws s3 cp "$BUCKET_PATH" "./$BACKUP_NAME" > /dev/null 2>&1
if [ $? -ne 0 ]; then
  echo "Failed to download backup from S3"
  exit 1
fi

echo "Decrypting and restoring backup..."
openssl enc -d -aes-256-cbc -in "./$BACKUP_NAME" -pass "file:$ENCRYPTION_KEY_PATH" -pbkdf2 | gunzip | sqlite3 "$RESTORED_DB"
if [ $? -ne 0 ]; then
  echo "Failed to decrypt and restore backup"
  rm -f "./$BACKUP_NAME"
  exit 1
fi

rm -f "./$BACKUP_NAME"

echo -e "Backup ${bold}$BACKUP_NAME${unbold} restored as ${bold}$RESTORED_DB${unbold}"
echo -e "To use this backup, rename it to ${bold}.n8n-shortlink/n8n-shortlink.sqlite${unbold}"
