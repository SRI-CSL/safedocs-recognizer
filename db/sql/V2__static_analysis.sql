CREATE TABLE public.static_analysis (
	parser TEXT,
	doc TEXT,
	digest TEXT,
	baseline BOOL,
	trace TEXT,
	"date" timestamptz NOT NULL DEFAULT now()
);
