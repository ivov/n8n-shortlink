#!/bin/bash

# Restore a sqlite DB backup from AWS S3: 
: '
~/.n8n-shortlink/backup/backup-restore.sh 2024-11-23-17:24:39+0100.sql.gz.enc
'

if [ $# -eq 0 ]; then
  echo -e "Error: No backup name provided.\nUsage: $0 <backup_name>\nExample: $0 2020-01-03-17:34:35+0200.sql.gz.enc"
  exit 1
fi

APP_DIR="$HOME/.n8n-shortlink"
BACKUP_DIR="$APP_DIR/backup"
BACKUP_ENCRYPTION_KEY="$BACKUP_DIR/n8n-shortlink-backup-encryption.key"
BACKUP_LOG_FILE="$BACKUP_DIR/backup.log"
RESTORED_DB="$BACKUP_DIR/restored.sqlite"

BUCKET_NAME=$(grep bucket_name ~/.aws/config | cut -d '=' -f2 | tr -d ' ')
if [ -z "$BUCKET_NAME" ]; then
  echo "Error: Bucket name not found in ~/.aws/config"
  exit 1
fi

BACKUP_NAME="$1"
BUCKET_URI="s3://$BUCKET_NAME/$BACKUP_NAME"

bold='\033[1m'
unbold='\033[0m'

rm -f $RESTORED_DB

echo -e "Selected backup: ${bold}$BACKUP_NAME${unbold}"
echo "Downloading backup..."
aws s3 cp "$BUCKET_URI" "./$BACKUP_NAME" > /dev/null 2>&1
if [ $? -ne 0 ]; then
  echo "Failed to download backup from S3"
  exit 1
fi

echo "Decrypting and restoring backup..."
openssl enc -d -aes-256-cbc -in "./$BACKUP_NAME" -pass "file:$BACKUP_ENCRYPTION_KEY" -pbkdf2 | gunzip | sqlite3 "$RESTORED_DB"
if [ $? -ne 0 ]; then
  echo "Failed to decrypt and restore backup"
  rm -f "./$BACKUP_NAME"
  exit 1
fi

rm -f "./$BACKUP_NAME"

echo -e "Backup ${bold}$BACKUP_NAME${unbold} restored as ${bold}$RESTORED_DB${unbold}"
echo -e "To replace the current DB with the restored DB, run the following command:"
echo -e "${bold}mv $RESTORED_DB $APP_DIR/n8n-shortlink.sqlite${unbold}"
