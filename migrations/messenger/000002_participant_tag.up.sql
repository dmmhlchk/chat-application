-- 1. create participant role-based tag
create type participant_tag as enum ('owner', 'member');

-- 2. add new column to 'participants' table
alter table participants add column tag participant_tag not null default 'member';