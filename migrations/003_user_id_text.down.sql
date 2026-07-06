ALTER TABLE url_mappings
    ALTER COLUMN user_id TYPE UUID USING user_id::uuid;
