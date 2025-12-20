create extension if not exists "uuid-ossp";

create table if not exists products (
  id uuid primary key,
  name text not null,
  base_price double precision not null,
  stock integer not null,
  updated_at timestamptz not null
);
