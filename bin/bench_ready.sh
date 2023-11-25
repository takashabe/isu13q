#!/bin/sh

:> /var/log/nginx/access.log
:> /var/log/mysql/slow.log

## require change to connect mysql
echo "set global slow_query_log_file = '/var/log/mysql/slow.log';set global long_query_time=0;set global slow_query_log = ON;" | mysql -uisucon -pisucon isupipe

restart_app.sh
