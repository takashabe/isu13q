#!/bin/bash

# リモートサーバーのアドレス
REMOTE_SERVER="isucon@57.180.238.61"

# リモートサーバーでサービスを停止
echo "Stopping isupipe-go service on remote server..."
ssh -t $REMOTE_SERVER "sudo systemctl stop isupipe-go"

# ファイルをアップロード
echo "Uploading file to remote server..."
scp ./isupipe $REMOTE_SERVER:/home/isucon/webapp/go/

# リモートサーバーでサービスを再開
echo "Starting isupipe-go service on remote server..."
ssh -t $REMOTE_SERVER "sudo systemctl start isupipe-go"

echo "Script completed."
