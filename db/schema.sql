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

CREATE SCHEMA public;


--
-- Name: SCHEMA public; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON SCHEMA public IS 'standard public schema';


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
-- Name: guitars_unique_slug(text); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.guitars_unique_slug(base text) RETURNS text
    LANGUAGE plpgsql
    AS $$
DECLARE
  candidate text := base;
  n integer := 1;
BEGIN
  IF candidate IS NULL OR btrim(candidate) = '' THEN
    RAISE EXCEPTION 'guitars_unique_slug: empty base slug';
  END IF;

  WHILE EXISTS (SELECT 1 FROM public.guitars WHERE slug = candidate) LOOP
    n := n + 1;
    candidate := base || '-' || n::text;
  END LOOP;

  RETURN candidate;
END
$$;


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


--
-- Name: slugify(text); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.slugify(txt text) RETURNS text
    LANGUAGE plpgsql IMMUTABLE
    AS $_$
DECLARE
  s text := coalesce(txt,'');
BEGIN
  s := unaccent(s);
  s := replace(s, '&', ' and ');
  s := lower(s);
  s := regexp_replace(s, '[^a-z0-9]+', '-', 'g');
  s := regexp_replace(s, '(^-+|-+$)', '', 'g');
  s := regexp_replace(s, '-{2,}', '-', 'g');
  RETURN s;
END
$_$;


--
-- Name: trg_guitars_set_slug(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.trg_guitars_set_slug() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
  base text;
BEGIN
  IF NEW.slug IS NULL OR btrim(NEW.slug) = '' THEN
    base := public.slugify(NEW.brand_slug) || '-' || public.slugify(NEW.model);
    NEW.slug := public.guitars_unique_slug(base);
  ELSE
    NEW.slug := public.guitars_unique_slug(public.slugify(NEW.slug));
  END IF;
  RETURN NEW;
END
$$;


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
    value text NOT NULL,
    feature_id uuid NOT NULL,
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    description text
);


--
-- Name: features; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.features (
    key text NOT NULL,
    label text NOT NULL,
    kind public.feature_kind NOT NULL,
    unit text,
    description text,
    id uuid DEFAULT gen_random_uuid() NOT NULL
);


--
-- Name: guitar_features; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.guitar_features (
    guitar_id uuid NOT NULL,
    value_text text,
    value_number numeric,
    value_boolean boolean,
    feature_id uuid NOT NULL,
    allowed_value_id uuid
);


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
-- Name: brands brands_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.brands
    ADD CONSTRAINT brands_pkey PRIMARY KEY (slug);


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
-- Name: uq_gf_guitar_feature; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uq_gf_guitar_feature ON public.guitar_features USING btree (guitar_id, feature_id);


--
-- Name: ux_guitars_slug; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX ux_guitars_slug ON public.guitars USING btree (slug);


--
-- Name: guitars trg_guitars_set_slug; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_guitars_set_slug BEFORE INSERT ON public.guitars FOR EACH ROW EXECUTE FUNCTION public.trg_guitars_set_slug();


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
-- Name: SCHEMA public; Type: ACL; Schema: -; Owner: -
--

GRANT USAGE ON SCHEMA public TO guitar_specs_ro;
GRANT USAGE ON SCHEMA public TO guitar_specs_web;


--
-- Name: TABLE brands; Type: ACL; Schema: public; Owner: -
--

GRANT SELECT ON TABLE public.brands TO guitar_specs_ro;


--
-- Name: TABLE feature_allowed_values; Type: ACL; Schema: public; Owner: -
--

GRANT SELECT ON TABLE public.feature_allowed_values TO guitar_specs_ro;


--
-- Name: TABLE features; Type: ACL; Schema: public; Owner: -
--

GRANT SELECT ON TABLE public.features TO guitar_specs_ro;


--
-- Name: TABLE guitar_features; Type: ACL; Schema: public; Owner: -
--

GRANT SELECT ON TABLE public.guitar_features TO guitar_specs_ro;


--
-- Name: TABLE guitars; Type: ACL; Schema: public; Owner: -
--

GRANT SELECT ON TABLE public.guitars TO guitar_specs_ro;


--
-- Name: TABLE shapes; Type: ACL; Schema: public; Owner: -
--

GRANT SELECT ON TABLE public.shapes TO guitar_specs_ro;


--
-- Name: DEFAULT PRIVILEGES FOR SEQUENCES; Type: DEFAULT ACL; Schema: public; Owner: -
--

ALTER DEFAULT PRIVILEGES FOR ROLE guitar_specs_owner IN SCHEMA public GRANT SELECT ON SEQUENCES TO guitar_specs_ro;


--
-- Name: DEFAULT PRIVILEGES FOR TABLES; Type: DEFAULT ACL; Schema: public; Owner: -
--

ALTER DEFAULT PRIVILEGES FOR ROLE guitar_specs_owner IN SCHEMA public GRANT SELECT ON TABLES TO guitar_specs_ro;


--
-- PostgreSQL database dump complete
--

