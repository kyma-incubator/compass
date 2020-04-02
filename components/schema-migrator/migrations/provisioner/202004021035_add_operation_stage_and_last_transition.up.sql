ALTER TABLE operation ADD COLUMN stage varchar(256);
ALTER TABLE operation ADD COLUMN last_transition timestamp without time zone;
