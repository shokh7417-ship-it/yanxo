-- Locations: canonical names and aliases for flexible city search (Turso/SQLite)

CREATE TABLE IF NOT EXISTS locations (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name_canonical TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS location_aliases (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  location_id INTEGER NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
  alias_normalized TEXT NOT NULL,
  UNIQUE(alias_normalized)
);

CREATE INDEX IF NOT EXISTS idx_location_aliases_normalized ON location_aliases(alias_normalized);
CREATE INDEX IF NOT EXISTS idx_location_aliases_location_id ON location_aliases(location_id);
