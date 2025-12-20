create extension if not exists "uuid-ossp";

create table if not exists users (
  id uuid primary key,
  email text not null unique,
  created_at timestamptz not null
);

create table if not exists orders (
  id uuid primary key,
  user_id uuid not null references users(id),
  product_id uuid not null,
  qty integer not null,
  status text not null,
  created_at timestamptz not null,
  updated_at timestamptz not null
);
