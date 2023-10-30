CREATE UNIQUE INDEX idx_id_hash
    ON fs_entities USING btree
    (id_hash ASC NULLS LAST);
