create extension if not exists "uuid-ossp";

create table if not exists prices (
  product_id uuid primary key,
  current_price double precision not null,
  updated_at timestamptz not null
);
