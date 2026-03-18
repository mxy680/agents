-- Agent templates: metadata about deployable agent types
create table if not exists agent_templates (
  id uuid primary key default gen_random_uuid(),
  name text unique not null,
  display_name text not null,
  description text not null default '',
  git_path text not null,
  docker_image text not null,
  required_integrations text[] not null default '{}',
  default_config jsonb not null default '{}',
  status text not null default 'draft' check (status in ('active', 'deprecated', 'draft')),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

-- Agent instances: running/stopped agent deployments per user
create table if not exists agent_instances (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null references auth.users(id) on delete cascade,
  template_id uuid not null references agent_templates(id),
  status text not null default 'pending' check (status in ('pending', 'creating', 'running', 'stopping', 'stopped', 'failed', 'completed')),
  k8s_pod_name text,
  k8s_namespace text not null default 'agents',
  config_overrides jsonb not null default '{}',
  error_message text,
  started_at timestamptz,
  stopped_at timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

-- RLS policies
alter table agent_templates enable row level security;
alter table agent_instances enable row level security;

-- Templates are readable by all authenticated users
create policy "templates_select" on agent_templates
  for select to authenticated using (true);

-- Instances: users can only see/manage their own
create policy "instances_select" on agent_instances
  for select to authenticated using (auth.uid() = user_id);

create policy "instances_insert" on agent_instances
  for insert to authenticated with check (auth.uid() = user_id);

create policy "instances_update" on agent_instances
  for update to authenticated using (auth.uid() = user_id);

create policy "instances_delete" on agent_instances
  for delete to authenticated using (auth.uid() = user_id);

-- Indexes
create index idx_agent_instances_user_id on agent_instances(user_id);
create index idx_agent_instances_status on agent_instances(status);
create index idx_agent_templates_status on agent_templates(status);
