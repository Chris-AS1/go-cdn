CREATE TABLE fs_entities
(
    id serial,
    id_hash bytea,
    filename character varying,
    content bytea,
    PRIMARY KEY (id)
);
