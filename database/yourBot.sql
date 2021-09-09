--
-- PostgreSQL database dump
--

-- Dumped from database version 11.12 (Raspbian 11.12-0+deb10u1)
-- Dumped by pg_dump version 11.12 (Raspbian 11.12-0+deb10u1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: youraccount; Type: TABLE; Schema: public; Owner: pi
--

CREATE TABLE public.youraccount (
    exchange character varying NOT NULL,
    currency character varying NOT NULL,
    amount double precision NOT NULL
);


ALTER TABLE public.youraccount OWNER TO pi;

--
-- Name: yourcandle; Type: TABLE; Schema: public; Owner: pi
--

CREATE TABLE public.yourcandle (
    exchange character varying(30) NOT NULL,
    pair character varying(30) NOT NULL,
    "timestamp" timestamp with time zone NOT NULL,
    open double precision NOT NULL,
    high double precision NOT NULL,
    low double precision NOT NULL,
    close double precision NOT NULL,
    volume double precision NOT NULL,
    asset character varying(255) NOT NULL,
    "interval" character varying(30)
);


ALTER TABLE public.yourcandle OWNER TO pi;

--
-- Name: yourlimits; Type: TABLE; Schema: public; Owner: pi
--

CREATE TABLE public.yourlimits (
    exchange character varying NOT NULL,
    pair character varying(30) NOT NULL,
    min double precision,
    avg double precision,
    max double precision,
    count double precision,
    potwin double precision,
    current double precision,
    limitbuy double precision,
    limitsell double precision,
    trend1 double precision,
    trend2 double precision,
    trend3 double precision,
    lastcandle double precision
);


ALTER TABLE public.yourlimits OWNER TO pi;

--
-- Name: yourorder; Type: TABLE; Schema: public; Owner: pi
--

CREATE TABLE public.yourorder (
    exchange character varying(30) NOT NULL,
    id character varying NOT NULL,
    pair character varying(30) NOT NULL,
    asset character varying(10) NOT NULL,
    price double precision NOT NULL,
    amount double precision NOT NULL,
    side character varying(10),
    "timestamp" timestamp without time zone NOT NULL,
    order_type character varying NOT NULL,
    status character varying NOT NULL
);


ALTER TABLE public.yourorder OWNER TO pi;

--
-- Name: yourparameter; Type: TABLE; Schema: public; Owner: pi
--

CREATE TABLE public.yourparameter (
    key character varying(50) NOT NULL,
    "int" bigint,
    "float" double precision,
    string character varying(255),
    date date,
    "time" time(0) without time zone,
    "timestamp" timestamp(0) without time zone
);


ALTER TABLE public.yourparameter OWNER TO pi;

--
-- Name: yourposition; Type: TABLE; Schema: public; Owner: pi
--

CREATE TABLE public.yourposition (
    exchange character varying NOT NULL,
    pair character varying NOT NULL,
    trtype character varying NOT NULL,
    "timestamp" timestamp without time zone NOT NULL,
    rate double precision NOT NULL,
    amount double precision NOT NULL,
    rid integer NOT NULL,
    active boolean DEFAULT true,
    notrade boolean DEFAULT false
);


ALTER TABLE public.yourposition OWNER TO pi;

--
-- Name: yourposition_rid_seq; Type: SEQUENCE; Schema: public; Owner: pi
--

CREATE SEQUENCE public.yourposition_rid_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.yourposition_rid_seq OWNER TO pi;

--
-- Name: yourposition_rid_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: pi
--

ALTER SEQUENCE public.yourposition_rid_seq OWNED BY public.yourposition.rid;


--
-- Name: yoursum; Type: TABLE; Schema: public; Owner: pi
--

CREATE TABLE public.yoursum (
    exchange character varying NOT NULL,
    valdate date NOT NULL,
    sum double precision
);


ALTER TABLE public.yoursum OWNER TO pi;

--
-- Name: yourtrace; Type: TABLE; Schema: public; Owner: pi
--

CREATE TABLE public.yourtrace (
    timest timestamp without time zone NOT NULL,
    exchange character varying NOT NULL,
    log text
);


ALTER TABLE public.yourtrace OWNER TO pi;

--
-- Name: yourtransaction; Type: TABLE; Schema: public; Owner: pi
--

CREATE TABLE public.yourtransaction (
    exchange character varying,
    "timestamp" timestamp without time zone,
    pair character varying,
    amount double precision,
    price double precision,
    amount_quote double precision,
    fee double precision,
    id character varying
);


ALTER TABLE public.yourtransaction OWNER TO pi;

--
-- Name: yposition; Type: VIEW; Schema: public; Owner: pi
--

CREATE VIEW public.yposition AS
 SELECT yourposition.exchange,
    yourposition.pair,
    yourposition.rate,
    yourposition.amount
   FROM public.yourposition
  ORDER BY yourposition.pair;


ALTER TABLE public.yposition OWNER TO pi;

--
-- Name: yourposition rid; Type: DEFAULT; Schema: public; Owner: pi
--

ALTER TABLE ONLY public.yourposition ALTER COLUMN rid SET DEFAULT nextval('public.yourposition_rid_seq'::regclass);


--
-- Name: youraccount_exchange_idx; Type: INDEX; Schema: public; Owner: pi
--

CREATE UNIQUE INDEX youraccount_exchange_idx ON public.youraccount USING btree (exchange, currency);


--
-- Name: yourcandle_exchange_name_idx; Type: INDEX; Schema: public; Owner: pi
--

CREATE UNIQUE INDEX yourcandle_exchange_name_idx ON public.yourcandle USING btree (exchange, pair, "timestamp", asset, "interval");


--
-- Name: yourlimits_exchange_idx; Type: INDEX; Schema: public; Owner: pi
--

CREATE UNIQUE INDEX yourlimits_exchange_idx ON public.yourlimits USING btree (exchange, pair);


--
-- Name: yourorder_exchange_idx; Type: INDEX; Schema: public; Owner: pi
--

CREATE UNIQUE INDEX yourorder_exchange_idx ON public.yourorder USING btree (exchange, id);


--
-- Name: yourparameter_key_idx; Type: INDEX; Schema: public; Owner: pi
--

CREATE UNIQUE INDEX yourparameter_key_idx ON public.yourparameter USING btree (key);


--
-- Name: yoursum_exchange_idx; Type: INDEX; Schema: public; Owner: pi
--

CREATE UNIQUE INDEX yoursum_exchange_idx ON public.yoursum USING btree (exchange, valdate);


--
-- Name: yourtransaction_exchange_idx; Type: INDEX; Schema: public; Owner: pi
--

CREATE INDEX yourtransaction_exchange_idx ON public.yourtransaction USING btree (exchange, "timestamp", pair, id);


--
-- PostgreSQL database dump complete
--

