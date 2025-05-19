--;;
create table if not exists system_prompt (
  id uuid not null primary key,
  name varchar(255) not null unique,
  description text,
  content text not null,
  created_at timestamp not null
);
--;;
CREATE INDEX IF NOT EXISTS idx_system_prompt_name ON system_prompt(name);
--;;
