-- Turso / libSQL (SQLite-compatible)

CREATE TABLE IF NOT EXISTS ads (
  id TEXT PRIMARY KEY,
  user_id INTEGER NOT NULL,
  category TEXT NOT NULL CHECK(category IN ('road','service')),
  status TEXT NOT NULL CHECK(status IN ('active','full','expired','replaced','deleted')),

  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,

  -- taxi fields
  from_city TEXT,
  to_city TEXT,
  ride_date TEXT,        -- YYYY-MM-DD
  departure_time TEXT,   -- HH:MM (24h)
  car_type TEXT,
  total_seats INTEGER,
  occupied_seats INTEGER,

  -- service fields
  service_type TEXT,
  area TEXT,
  note TEXT,

  -- common
  contact TEXT,
  channel_message_id INTEGER
);

CREATE INDEX IF NOT EXISTS idx_ads_category ON ads(category);
CREATE INDEX IF NOT EXISTS idx_ads_status ON ads(status);
CREATE INDEX IF NOT EXISTS idx_ads_route ON ads(from_city, to_city);
CREATE INDEX IF NOT EXISTS idx_ads_ride_date ON ads(ride_date);
CREATE INDEX IF NOT EXISTS idx_ads_user_id ON ads(user_id);

