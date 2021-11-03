ALTER TABLE consensus ADD COLUMN fname TEXT;

CREATE INDEX idx_fname ON public.consensus USING btree (fname);

UPDATE consensus SET fname = substring(doc from '(?:.+/)(.+)');