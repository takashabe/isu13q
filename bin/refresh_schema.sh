#!/bin/bash

set -euo pipefail

sudo mycli -e "DROP DATABASE IF EXISTS isupipe"
sudo mycli -e "CREATE DATABASE isupipe"
cat $HOME/webapp/sql/initdb.d/10_schema.sql | sudo mycli
