-- 1. drop tables that depend on other tables first (Foreign Keys)
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS participants;

-- 2. drop parent tables
DROP TABLE IF EXISTS chats;
DROP TABLE IF EXISTS users;

-- 3. drop custom enum types (only possible after deleting tables that use them)
DROP TYPE IF EXISTS message_type;
DROP TYPE IF EXISTS chat_type;