-- 20240916_add_title.sql
BEGIN;
ALTER TABLE public.single_release_escrow ADD COLUMN IF NOT EXISTS title text;
ALTER TABLE public.multi_release_escrow  ADD COLUMN IF NOT EXISTS title text;
COMMIT;
