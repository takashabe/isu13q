#!/bin/sh

## nginx
export ALP_FILENAME=/tmp/alp-`date +%H%M%S`
# alp json -m でidとかまとめられる
# https://zenn.dev/tkuchiki/articles/how-to-use-alp
cat /var/log/nginx/access.log | alp ltsv --sort avg -r -m '/api/user/.+/icon','api/user/.+/theme','api/user/.+/statistics','api/livestream/.+/reaction','api/livestream/.+/livecomment','api/livestream/.+/enter','/api/livestream/.+/ngwords','/api/livestream/.+/report','/api/livestream/.+/exit','/api/livestream/.+/moderate','/api/livestream/.+/statistics','api/user/.+/livestream' > $ALP_FILENAME
head -n 30 $ALP_FILENAME | notify_slack

## mysql
echo "set global slow_query_log = OFF;" | mysql -h192.168.0.13 -uisucon -pisucon isupipe
scp isucon@192.168.0.13:/var/log/mysql/slow.log /var/log/mysql/slow.log
export QUERY_FILENAME=/tmp/query-`date +%H%M%S`
pt-query-digest /var/log/mysql/slow.log > $QUERY_FILENAME
head -n 30 $QUERY_FILENAME | notify_slack
