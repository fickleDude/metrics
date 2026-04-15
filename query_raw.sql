	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
drop table gauge;
drop table counter;
CREATE TABLE gauge (
    id     varchar,
    value    numeric,
    PRIMARY KEY (id)
);

select current_database();

select * from counter;
INSERT INTO gauge (id, value) 
VALUES ('mem', 2.8)
ON CONFLICT (id) DO UPDATE 
  SET value = excluded.value;
select delta from counter where id = 'mem';
CREATE TABLE counter (
    id     varchar,
    delta    int,
    PRIMARY KEY (id)
);
