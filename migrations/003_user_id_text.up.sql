ALTER TABLE url_mappings
    ALTER COLUMN user_id TYPE TEXT USING user_id::text;
