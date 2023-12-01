#!/bin/sh

# Start MySQL in the background
mysqld --user=mysql --datadir=/var/lib/mysql --skip-networking &

# Start phpMyAdmin
phpmyadmin