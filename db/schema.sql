--
-- PostgreSQL database dump
--

-- Dumped from database version 17.5 (Postgres.app)
-- Dumped by pg_dump version 17.5 (Postgres.app)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: public; Type: SCHEMA; Schema: -; Owner: -
--

-- *not* creating schema, since initdb creates it


--
-- Name: citext; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS citext WITH SCHEMA public;


--
-- Name: EXTENSION citext; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION citext IS 'data type for case-insensitive character strings';


--
-- Name: pg_trgm; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pg_trgm WITH SCHEMA public;


--
-- Name: EXTENSION pg_trgm; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION pg_trgm IS 'text similarity measurement and index searching based on trigrams';


--
-- Name: pgcrypto; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;


--
-- Name: EXTENSION pgcrypto; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION pgcrypto IS 'cryptographic functions';


--
-- Name: feature_kind; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.feature_kind AS ENUM (
    'text',
    'number',
    'boolean',
    'enum'
);


--
-- Name: guitar_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.guitar_type AS ENUM (
    'electric',
    'acoustic',
    'classical',
    'bass',
    'ukulele'
);


--
-- Name: set_updated_at(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.set_updated_at() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  NEW.updated_at := now();
  RETURN NEW;
END$$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: brands; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.brands (
    slug public.citext NOT NULL,
    name text NOT NULL,
    about text,
    website_url text,
    wikipedia_url text,
    logo_url text,
    country_code text,
    founded_year integer,
    headquarters text,
    meta jsonb DEFAULT '{}'::jsonb NOT NULL,
    CONSTRAINT brands_founded_year_check CHECK (((founded_year IS NULL) OR ((founded_year >= 1700) AND (founded_year <= 2100)))),
    CONSTRAINT brands_logo_url_check CHECK (((logo_url IS NULL) OR (logo_url ~ '^https?://'::text))),
    CONSTRAINT brands_slug_check CHECK ((slug OPERATOR(public.~) '^[a-z0-9-]+$'::public.citext)),
    CONSTRAINT brands_website_url_check CHECK (((website_url IS NULL) OR (website_url ~ '^https?://'::text))),
    CONSTRAINT brands_wikipedia_url_check CHECK (((wikipedia_url IS NULL) OR (wikipedia_url ~ '^https?://'::text)))
);


--
-- Name: feature_allowed_values; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.feature_allowed_values (
    id bigint NOT NULL,
    feature_id bigint NOT NULL,
    value text NOT NULL,
    label text NOT NULL
);


--
-- Name: feature_allowed_values_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.feature_allowed_values_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: feature_allowed_values_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.feature_allowed_values_id_seq OWNED BY public.feature_allowed_values.id;


--
-- Name: features; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.features (
    id bigint NOT NULL,
    key text NOT NULL,
    label text NOT NULL,
    kind public.feature_kind NOT NULL,
    unit text,
    description text
);


--
-- Name: features_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.features_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: features_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.features_id_seq OWNED BY public.features.id;


--
-- Name: guitar_features; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.guitar_features (
    guitar_id uuid NOT NULL,
    feature_id bigint NOT NULL,
    allowed_value_id bigint,
    value_text text,
    value_number numeric,
    value_boolean boolean
);


--
-- Name: guitar_features_resolved; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.guitar_features_resolved AS
 SELECT gf.guitar_id,
    f.key AS feature_key,
    f.label AS feature_label,
    f.kind AS feature_kind,
    COALESCE(fav.label, gf.value_text) AS value_label,
    gf.value_number,
    gf.value_boolean,
    f.unit
   FROM ((public.guitar_features gf
     JOIN public.features f ON ((f.id = gf.feature_id)))
     LEFT JOIN public.feature_allowed_values fav ON ((fav.id = gf.allowed_value_id)));


--
-- Name: guitars; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.guitars (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    slug public.citext NOT NULL,
    type public.guitar_type NOT NULL,
    model text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    shape_slug public.citext NOT NULL,
    brand_slug public.citext NOT NULL
);


--
-- Name: shapes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.shapes (
    slug public.citext NOT NULL,
    name text NOT NULL,
    description text,
    meta jsonb DEFAULT '{}'::jsonb NOT NULL,
    CONSTRAINT shapes_slug_check CHECK ((slug OPERATOR(public.~) '^[a-z0-9-]+$'::public.citext))
);


--
-- Name: feature_allowed_values id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.feature_allowed_values ALTER COLUMN id SET DEFAULT nextval('public.feature_allowed_values_id_seq'::regclass);


--
-- Name: features id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.features ALTER COLUMN id SET DEFAULT nextval('public.features_id_seq'::regclass);


--
-- Name: brands brands_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.brands
    ADD CONSTRAINT brands_pkey PRIMARY KEY (slug);


--
-- Name: feature_allowed_values feature_allowed_values_feature_id_value_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.feature_allowed_values
    ADD CONSTRAINT feature_allowed_values_feature_id_value_key UNIQUE (feature_id, value);


--
-- Name: feature_allowed_values feature_allowed_values_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.feature_allowed_values
    ADD CONSTRAINT feature_allowed_values_pkey PRIMARY KEY (id);


--
-- Name: features features_key_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.features
    ADD CONSTRAINT features_key_key UNIQUE (key);


--
-- Name: features features_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.features
    ADD CONSTRAINT features_pkey PRIMARY KEY (id);


--
-- Name: guitar_features guitar_features_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.guitar_features
    ADD CONSTRAINT guitar_features_pkey PRIMARY KEY (guitar_id, feature_id);


--
-- Name: guitars guitars_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.guitars
    ADD CONSTRAINT guitars_pkey PRIMARY KEY (id);


--
-- Name: guitars guitars_slug_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.guitars
    ADD CONSTRAINT guitars_slug_key UNIQUE (slug);


--
-- Name: shapes shapes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.shapes
    ADD CONSTRAINT shapes_pkey PRIMARY KEY (slug);


--
-- Name: idx_brands_meta_gin; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_brands_meta_gin ON public.brands USING gin (meta);


--
-- Name: idx_brands_name_trgm; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_brands_name_trgm ON public.brands USING gin (name public.gin_trgm_ops);


--
-- Name: idx_guitar_features_allowed; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_guitar_features_allowed ON public.guitar_features USING btree (allowed_value_id);


--
-- Name: idx_guitar_features_bool; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_guitar_features_bool ON public.guitar_features USING btree (value_boolean);


--
-- Name: idx_guitar_features_feature; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_guitar_features_feature ON public.guitar_features USING btree (feature_id);


--
-- Name: idx_guitar_features_num; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_guitar_features_num ON public.guitar_features USING btree (value_number);


--
-- Name: idx_guitars_brand_slug; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_guitars_brand_slug ON public.guitars USING btree (brand_slug);


--
-- Name: idx_guitars_model_trgm; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_guitars_model_trgm ON public.guitars USING gin (model public.gin_trgm_ops);


--
-- Name: idx_guitars_shape_slug; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_guitars_shape_slug ON public.guitars USING btree (shape_slug);


--
-- Name: idx_guitars_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_guitars_type ON public.guitars USING btree (type);


--
-- Name: idx_shapes_meta_gin; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_shapes_meta_gin ON public.shapes USING gin (meta);


--
-- Name: idx_shapes_name_trgm; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_shapes_name_trgm ON public.shapes USING gin (name public.gin_trgm_ops);


--
-- Name: guitars trg_guitars_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_guitars_updated_at BEFORE UPDATE ON public.guitars FOR EACH ROW EXECUTE FUNCTION public.set_updated_at();


--
-- Name: feature_allowed_values feature_allowed_values_feature_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.feature_allowed_values
    ADD CONSTRAINT feature_allowed_values_feature_id_fkey FOREIGN KEY (feature_id) REFERENCES public.features(id) ON DELETE CASCADE;


--
-- Name: guitars fk_guitars_brand_slug; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.guitars
    ADD CONSTRAINT fk_guitars_brand_slug FOREIGN KEY (brand_slug) REFERENCES public.brands(slug) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: guitars fk_guitars_shape_slug; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.guitars
    ADD CONSTRAINT fk_guitars_shape_slug FOREIGN KEY (shape_slug) REFERENCES public.shapes(slug) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: guitar_features guitar_features_allowed_value_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.guitar_features
    ADD CONSTRAINT guitar_features_allowed_value_id_fkey FOREIGN KEY (allowed_value_id) REFERENCES public.feature_allowed_values(id) ON DELETE SET NULL;


--
-- Name: guitar_features guitar_features_feature_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.guitar_features
    ADD CONSTRAINT guitar_features_feature_id_fkey FOREIGN KEY (feature_id) REFERENCES public.features(id) ON DELETE CASCADE;


--
-- Name: guitar_features guitar_features_guitar_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.guitar_features
    ADD CONSTRAINT guitar_features_guitar_id_fkey FOREIGN KEY (guitar_id) REFERENCES public.guitars(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

