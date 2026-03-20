-- Atomic increment for acquisition_count to avoid read-modify-write race.
create or replace function increment_acquisition_count(p_template_id uuid)
returns void
language sql
as $$
  update agent_templates
  set acquisition_count = coalesce(acquisition_count, 0) + 1
  where id = p_template_id;
$$;
