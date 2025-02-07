CREATE EXTENSION if not exists vector;
--;;
create table if not exists context (
  id uuid not null primary key,
  name varchar(255) not null unique,
  description text,
  created_at timestamp not null
);
--;;
CREATE INDEX IF NOT EXISTS idx_context_name ON context(name);
--;;
create table if not exists context_source (
  ordering bigserial,
  context_id uuid not null,
  source_context_id uuid not null,
  constraint fk_context foreign key(context_id) references context(id),
  constraint fk_source_context foreign key(source_context_id) references context(id),
  unique(context_id, source_context_id)
);
--;;
create table if not exists context_message (
  ordering bigserial,
  id uuid not null primary key,
  role varchar(255) not null,
  content text not null,
  created_at timestamp not null,
  context_id uuid not null,
  constraint fk_context foreign key(context_id) references context(id)
);
--;;
CREATE INDEX IF NOT EXISTS idx_context_message_ordering ON context_message(ordering);
CREATE INDEX IF NOT EXISTS idx_context_message_context_id ON context_message(context_id);
--;;
CREATE TABLE if not exists document (
id uuid not null primary key,
name varchar(255) not null unique,
description text,
created_at timestamp not null
);
--;;
CREATE TABLE if not exists document_chunk (
id uuid not null primary key,
document_id uuid not null,
fragment text,
embedding vector(1024),
created_at timestamp not null,
constraint fk_document_id foreign key(document_id) references document(id)
);
--;;
CREATE INDEX IF NOT EXISTS idx_document_chunk_document_id ON document_chunk(document_id);
--;;
