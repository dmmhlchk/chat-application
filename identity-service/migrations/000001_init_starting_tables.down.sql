-- 1. drop session table
drop table if exists sessions;

-- 2. drop user table
drop table if exists users;

-- 3. drop enum only when it's not in use (after all removed dependencies)
drop type if exists platform;