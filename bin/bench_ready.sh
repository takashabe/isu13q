#!/bin/sh

:> /var/log/nginx/access.log

ssh isucon@192.168.0.13 "sudo sh -c 'echo > /var/log/mysql/mysql-slow.log'"
# :> /var/log/mysql/slow.log

## require change to connect mysql
# echo "set global slow_query_log_file = '/var/log/mysql/slow.log';set global long_query_time=0;set global slow_query_log = ON;" | mysql -h192.168.0.13 -uisucon -pisucon isupipe

restart_app.sh
