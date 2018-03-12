CREATE SCHEMA passwdpot;
use passwdpot;

CREATE TABLE geo (
 id BIGINT NOT NULL AUTO_INCREMENT,
 ip varchar(16) NOT NULL,
 country_code char(2) NOT NULL,
 region_code varchar(256),
 region_name varchar(256),
 city varchar(256),
 time_zone varchar(64),
 latitude float,
 longitude float,
 metro_code int,
 last_update timestamp(6) NOT NULL,
 PRIMARY KEY(id)
);


CREATE TABLE event (
  id BIGINT NOT NULL AUTO_INCREMENT,
  dt timestamp(6)  NOT NULL,
  username varchar(256) NOT NULL,
  passwd varchar(512) NOT NULL,
  remote_addr varchar(16) NOT NULL,
  remote_geo_id bigint NULL REFERENCES geo(id),
  remote_port bigint NULL,
  remote_name varchar(256),
  remote_version varchar(1024),
  origin_addr varchar(16) NOT NULL,
  origin_geo_id bigint NULL REFERENCES geo(id),
  application varchar(32),
  protocol varchar(32),
  PRIMARY KEY(id)
);

CREATE OR REPLACE VIEW event_geo AS
SELECT a.id,
       a.dt,
       a.username,
       a.passwd,
       a.remote_addr,
       a.remote_name,
       a.remote_version,
       a.remote_port,
       coalesce(b.country_code, '')  as remote_country,
       coalesce(b.city, '') as remote_city,
       a.origin_addr,
       coalesce(c.country_code, '') as origin_country,
       coalesce(c.city, '')  as origin_city,
       coalesce(b.latitude,0) as remote_latitude,
       coalesce(b.longitude,0) as remote_longitude,
       coalesce(c.latitude, 0) as origin_latitude,
       coalesce(c.longitude, 0) as origin_longitude
       FROM event a
       LEFT JOIN geo b on a.remote_geo_id = b.id
       LEFT JOIN geo c on a.origin_geo_id = c.id;

CREATE TABLE country_stats (
       country_code char(2),
       latitude float,
       longitude float,
       hits int
);
