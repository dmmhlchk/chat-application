-- 1. drop tables that depend on other tables first (foreign keys)
drop table if exists messages;
drop table if exists participants;

-- 2. drop parent tables
drop table if exists chats;
drop table if exists users;

-- 3. drop custom enum types (only possible after deleting tables that use them)
drop type if exists message_type;
drop type if exists chat_type;