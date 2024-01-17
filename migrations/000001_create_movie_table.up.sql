CREATE TABLE IF NOT EXISTS movies (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    title text NOT NULL,
    year integer NOT NULL,
    runtime integer NOT NULL CHECK (runtime > 0), -- this is a way of adding constraints while defining property,
    genres text[] NOT NULL,
    version integer NOT NULL DEFAULT 1,
    CONSTRAINT version_positive CHECK (version > 0) -- a way of defining additional constraint while defining property
)

