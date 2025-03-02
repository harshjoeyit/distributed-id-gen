CREATE TABLE `Tickets64` (
  `id` bigint(20) unsigned NOT NULL auto_increment,
  `stub` char(1) NOT NULL default '',
  PRIMARY KEY  (`id`),
  UNIQUE KEY `stub` (`stub`)
);

-- Offset for one server is 1 and for other server is 2
INSERT INTO ticket (stub) VALUES ('a') ON DUPLICATE KEY UPDATE id = id + 1;