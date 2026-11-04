CREATE TYPE source_name AS ENUM ('sales', 'service');

CREATE TABLE snapshot (
    vin        varchar(17) NOT NULL,
    source     source_name NOT NULL,
    data       jsonb       NOT NULL,
    fetched_at timestamptz NOT NULL,
    PRIMARY KEY (vin, source)
);
