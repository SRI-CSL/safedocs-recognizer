CREATE INDEX idx_digest ON public.consensus USING btree (digest);
CREATE INDEX idx_parser_digest ON public.consensus USING btree (parser, digest);
CREATE INDEX idx_parser_digest_doc ON public.consensus USING btree (parser, digest, doc);
