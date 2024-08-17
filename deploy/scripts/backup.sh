#!/bin/bash

# Back up sqlite DB to AWS S3: deploy/scripts/backup.sh (via cronjob)

CONFIG_FILEPATH="$HOME/deploy/.config"

if [ ! -f "$CONFIG_FILEPATH" ]; then
  echo "Error: Config file $CONFIG_FILEPATH not found."
  exit 1
fi

DB_PATH=".n8n-shortlink/n8n-shortlink.sqlite"
BUCKET_NAME=$(grep BUCKET_NAME $CONFIG_FILEPATH | cut -d'=' -f2 | tr -d '"*')
PLAINTEXT_BACKUP_FILENAME="$(date +%Y-%m-%d-%H:%M:%S%z).sql.gz"
ENCRYPTED_BACKUP_FILENAME="$PLAINTEXT_BACKUP_FILENAME.enc"
ENCRYPTION_KEY_PATH="$HOME/.keys/n8n-shortlink-backup-secret.key"
LOG_FILE="$HOME/deploy/backups.log"
BUCKET_PATH="s3://$BUCKET_NAME/n8n-shortlink-backups/$ENCRYPTED_BACKUP_FILENAME"

if [ ! -f "$DB_PATH" ]; then
  echo "❌ Backup failed to start because of missing DB file at $DB_PATH" | tee -a $LOG_FILE
  exit 1
fi

# ==================================
#        compress + encrypt
# ==================================

sqlite3 $DB_PATH .dump | gzip > "./$PLAINTEXT_BACKUP_FILENAME"
openssl enc -aes-256-cbc -salt -in "./$PLAINTEXT_BACKUP_FILENAME" -out "./$ENCRYPTED_BACKUP_FILENAME" -pass "file:$ENCRYPTION_KEY_PATH" -pbkdf2

# ==================================
#             upload
# ==================================

log_message() {
  echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" | tee -a "$LOG_FILE"
}

UPLOAD_OUTPUT=$(aws s3 cp "./$ENCRYPTED_BACKUP_FILENAME" $BUCKET_PATH 2>&1)
EXIT_STATUS=$?
if [ $EXIT_STATUS -eq 0 ]; then
  log_message "Backup uploaded: $ENCRYPTED_BACKUP_FILENAME"
else
  log_message "Backup upload failed: $ENCRYPTED_BACKUP_FILENAME. Details: $UPLOAD_OUTPUT"
fi

rm ./$ENCRYPTED_BACKUP_FILENAME
rm ./$PLAINTEXT_BACKUP_FILENAME
