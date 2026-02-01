alter table topics enable row level security;
alter table sources enable row level security;
alter table articles enable row level security;
alter table runs enable row level security;

-- No policies are defined by default to keep tables private.
