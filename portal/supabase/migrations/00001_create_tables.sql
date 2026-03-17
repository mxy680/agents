-- integrations table: stores encrypted OAuth tokens per user per provider
-- Users are managed by Supabase Auth (auth.users)

create table public.integrations (
  id            uuid primary key default gen_random_uuid(),
  user_id       uuid not null references auth.users(id) on delete cascade,
  provider      text not null,
  status        text not null default 'active',
  access_token  text,
  refresh_token text,
  token_expiry  timestamptz,
  metadata      jsonb not null default '{}',
  created_at    timestamptz not null default now(),
  updated_at    timestamptz not null default now(),

  unique(user_id, provider)
);

alter table public.integrations enable row level security;

create policy "Users can view own integrations"
  on public.integrations for select
  using (auth.uid() = user_id);

create policy "Users can insert own integrations"
  on public.integrations for insert
  with check (auth.uid() = user_id);

create policy "Users can update own integrations"
  on public.integrations for update
  using (auth.uid() = user_id);

create policy "Users can delete own integrations"
  on public.integrations for delete
  using (auth.uid() = user_id);

create or replace function public.update_updated_at()
returns trigger as $$
begin
  new.updated_at = now();
  return new;
end;
$$ language plpgsql;

create trigger integrations_updated_at
  before update on public.integrations
  for each row execute function public.update_updated_at();
