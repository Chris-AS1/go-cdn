CREATE TABLE fs_entities
(
    id serial,
    id_hash character varying,
    filename character varying,
    content bytea,
    PRIMARY KEY (id)
);
