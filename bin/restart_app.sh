#!/bin/sh

## Require change to system name
sudo systemctl restart isu-go
curl http://localhost
