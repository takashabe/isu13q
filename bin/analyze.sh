#!/bin/sh

## nginx
export ALP_FILENAME=/tmp/alp-`date +%H%M%S`
# alp json -m でidとかまとめられる
# https://zenn.dev/tkuchiki/articles/how-to-use-alp
cat /var/log/nginx/access.log | alp json --sort avg -r > $ALP_FILENAME
head -n 30 $ALP_FILENAME | notify_slack

## mysql
echo "set global slow_query_log = OFF;" | mysql -uisucon -pisucon isupipe
export QUERY_FILENAME=/tmp/query-`date +%H%M%S`
pt-query-digest /var/log/mysql/slow.log > $QUERY_FILENAME
head -n 30 $QUERY_FILENAME | notify_slack
