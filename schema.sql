--CREATE IF DOES NOT EXIST
CREATE TABLE IF NOT EXISTS packets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    data TEXT,
    created_at DATETIME,
    updated_at DATETIME,
    device_id INTEGER,
    FOREIGN KEY (device_id) REFERENCES devices (id)
);

CREATE TABLE IF NOT EXISTS devices (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT,
    description TEXT
);

CREATE INDEX IF NOT EXISTS idx_packets_device_id ON packets (device_id);
