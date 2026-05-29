CREATE TABLE IF NOT EXISTS gauge (
    id     varchar,
    value    numeric,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS counter (
    id     varchar,
    delta    bigint,
    PRIMARY KEY (id)
);