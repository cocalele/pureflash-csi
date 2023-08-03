#!/bin/bash
set -m

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
echo "Welcome to PureFlash(https://github.com/cocalele/PureFlash) mariadb!"
echo "Start MariaDB..."
mysql_install_db --user=mysql --ldata=/var/lib/mysql
/usr/bin/mysqld_safe --user=mysql --datadir='/var/lib/mysql' --log-error=/var/log/mysql/error.log  &
echo "Waiting mysql start ..."
sleep 3
if  ! mysql -e "use s5" ; then
  echo "initialize database s5 ..."
  mysql -e "source /opt/pureflash/mariadb/init_s5metadb.sql"
  mysql -e "GRANT ALL PRIVILEGES ON *.* TO 'pureflash'@'%' IDENTIFIED BY '123456'" 
fi
fg 1

