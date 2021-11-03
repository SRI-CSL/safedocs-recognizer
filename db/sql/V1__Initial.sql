-- public.consensus definition

-- Drop table

-- DROP TABLE public.consensus;

CREATE TABLE public.consensus (
	"parser" text NULL,
	doc text NULL,
	baseline bool NULL,
	digest text NULL,
	status text NULL,
	"stdout" text NULL,
	stderr text NULL,
	callgrind text NULL,
	cfg text NULL,
	cfg_image bytea NULL,
	"date" timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_doc ON public.consensus USING btree (doc);
CREATE INDEX idx_parser ON public.consensus USING btree (parser);
CREATE INDEX idx_parser_doc ON public.consensus USING btree (parser, doc);
