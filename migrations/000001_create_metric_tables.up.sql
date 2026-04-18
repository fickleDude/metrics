CREATE TABLE IF NOT EXISTS gauge (
    id     varchar,
    value    numeric,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS counter (
    id     varchar,
    delta    int,
    PRIMARY KEY (id)
);