-- public.youraccount definition

-- Drop table

-- DROP TABLE public.youraccount;

CREATE TABLE public.youraccount (
	exchange varchar NOT NULL,
	currency varchar NOT NULL,
	amount float8 NOT NULL
);
CREATE UNIQUE INDEX youraccount_exchange_idx ON public.youraccount USING btree (exchange, currency);


-- public.yourcandle definition

-- Drop table

-- DROP TABLE public.yourcandle;

CREATE TABLE public.yourcandle (
	exchange varchar(30) NOT NULL,
	pair varchar(30) NOT NULL,
	"timestamp" timestamptz NOT NULL,
	"open" float8 NOT NULL,
	high float8 NOT NULL,
	low float8 NOT NULL,
	"close" float8 NOT NULL,
	volume float8 NOT NULL,
	asset varchar(255) NOT NULL,
	"interval" varchar(30) NULL
);
CREATE UNIQUE INDEX yourcandle_exchange_name_idx ON public.yourcandle USING btree (exchange, pair, "timestamp", asset, "interval");

-- public.yourlimits definition

-- Drop table

-- DROP TABLE public.yourlimits;

CREATE TABLE public.yourlimits (
	exchange varchar NOT NULL,
	pair varchar(30) NOT NULL,
	min float8 NULL,
	avg float8 NULL,
	max float8 NULL,
	count float8 NULL,
	potwin float8 NULL,
	"current" float8 NULL,
	limitbuy float8 NULL,
	limitsell float8 NULL,
	trend1 float8 NULL,
	trend2 float8 NULL,
	trend3 float8 NULL
);
CREATE UNIQUE INDEX yourlimits_exchange_idx ON public.yourlimits USING btree (exchange, pair);

-- public.yourorder definition

-- Drop table

-- DROP TABLE public.yourorder;

CREATE TABLE public.yourorder (
	exchange varchar(30) NOT NULL,
	id varchar(20) NOT NULL,
	pair varchar(30) NOT NULL,
	asset varchar(10) NOT NULL,
	price float8 NOT NULL,
	amount float8 NOT NULL,
	side varchar(10) NULL,
	"timestamp" timestamp NOT NULL,
	order_type varchar NOT NULL,
	status varchar NOT NULL
);
CREATE UNIQUE INDEX yourorder_exchange_idx ON public.yourorder USING btree (exchange, id);

-- public.yourparameter definition

-- Drop table

-- DROP TABLE public.yourparameter;

CREATE TABLE public.yourparameter (
	"key" varchar(50) NOT NULL,
	"int" int8 NULL,
	"float" float8 NULL,
	string varchar(255) NULL,
	"date" date NULL,
	"time" time(0) NULL,
	"timestamp" timestamp(0) NULL
);
CREATE UNIQUE INDEX yourparameter_key_idx ON public.yourparameter USING btree (key);

-- public.yourposition definition

-- Drop table

-- DROP TABLE public.yourposition;

CREATE TABLE public.yourposition (
	exchange varchar NOT NULL,
	pair varchar NOT NULL,
	trtype varchar NOT NULL,
	"timestamp" timestamp NOT NULL,
	rate float8 NOT NULL,
	amount float8 NOT NULL,
	rid serial NOT NULL,
	active bool NULL DEFAULT true,
	notrade bool NULL DEFAULT false
);


