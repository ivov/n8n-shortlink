#!/bin/bash

# Back up sqlite DB to AWS S3: ./backup-run.sh

set -euo pipefail

required_env_vars=(
  "APP_DIR"
  "BACKUP_DIR"
  "APP_DB" 
  "BACKUP_ENCRYPTION_KEY"
  "BUCKET_NAME"
  "BUCKET_PREFIX"
)

for var in "${required_env_vars[@]}"; do
  if [ -z "${!var}" ]; then
    echo "Required environment variable $var is not set"
    exit 1
  fi
done

PLAINTEXT_BACKUP_NAME="$(date +%Y-%m-%d-%H:%M:%S%z).sql.gz"
ENCRYPTED_BACKUP_NAME="$PLAINTEXT_BACKUP_NAME.enc"
BACKUP_LOG_FILE="$BACKUP_DIR/backup.log"
BUCKET_URI="s3://$BUCKET_NAME/$BUCKET_PREFIX/$BACKUP_NAME"
TEMP_DIR=$(mktemp -d)

trap 'rm -rf "$TEMP_DIR"' EXIT

# ==================================
#        compress + encrypt
# ==================================

sqlite3 $APP_DB .dump | gzip > "$TEMP_DIR/$PLAINTEXT_BACKUP_NAME"

if [ $? -ne 0 ]; then
  log_message "Failed to dump and compress database"
  exit 1
fi

openssl enc -aes-256-cbc -salt -in "$TEMP_DIR/$PLAINTEXT_BACKUP_NAME" -out "$TEMP_DIR/$ENCRYPTED_BACKUP_NAME" -pass "file:$BACKUP_ENCRYPTION_KEY" -pbkdf2

if [ $? -ne 0 ]; then
  log_message "Failed to encrypt backup"
  exit 1
fi

# ==================================
#             upload
# ==================================

log_message() {
  echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" | tee -a "$BACKUP_LOG_FILE"
}

UPLOAD_OUTPUT=$(aws s3 cp "./$ENCRYPTED_BACKUP_NAME" $BUCKET_URI 2>&1)
EXIT_STATUS=$?
if [ $EXIT_STATUS -eq 0 ]; then
  log_message "Backup uploaded: $ENCRYPTED_BACKUP_NAME"
else
  log_message "Backup upload failed: $ENCRYPTED_BACKUP_NAME. Details: $UPLOAD_OUTPUT"
fi
