#!/bin/bash
set -ex

BRANCH=$1
cd $HOME/<todo>
git pull origin $BRANCH
make

restart_app.sh
bench_ready.sh
