drop database IF EXISTS cache1;
create database cache1;
use cache1;
drop table IF EXISTS t1;
drop table IF EXISTS ver;

create table t1(
  id BIGINT(11) PRIMARY KEY AUTO_INCREMENT,
  value VARCHAR(100)
);

create table ver (
  id BIGINT(11) PRIMARY KEY AUTO_INCREMENT,
  value VARCHAR(11),
  version int(100)
);