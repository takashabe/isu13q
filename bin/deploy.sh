#!/bin/bash
set -ex

HOST=todo
BRANCH=$1
USERNAME=$USER
ssh isucon@$HOST "echo $USERNAME 'deploying... | notify_slack' && deploy_in_remote && echo $USERNAME 'deploy done' | notify_slack"
