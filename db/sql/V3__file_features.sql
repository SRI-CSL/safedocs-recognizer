CREATE TABLE public.file_features (
	doc TEXT,
	digest TEXT,
	baseline BOOL,
	features JSONB,
	"date" timestamptz NOT NULL DEFAULT now()
);
