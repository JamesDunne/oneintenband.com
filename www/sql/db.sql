--
-- PostgreSQL database dump
--

SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;

--
-- Name: band; Type: DATABASE; Schema: -; Owner: postgres
--

CREATE DATABASE band WITH TEMPLATE = template0 ENCODING = 'UTF8' LC_COLLATE = 'C' LC_CTYPE = 'C';


ALTER DATABASE band OWNER TO postgres;

\connect band

SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;

--
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


SET search_path = public, pg_catalog;

--
-- Name: album_id; Type: SEQUENCE; Schema: public; Owner: band
--

CREATE SEQUENCE album_id
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.album_id OWNER TO band;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: album; Type: TABLE; Schema: public; Owner: band; Tablespace: 
--

CREATE TABLE album (
    album_id integer DEFAULT nextval('album_id'::regclass) NOT NULL,
    date date,
    title character varying,
    description character varying,
    best_album_mix_id integer
);


ALTER TABLE public.album OWNER TO band;

--
-- Name: album_mix_id; Type: SEQUENCE; Schema: public; Owner: band
--

CREATE SEQUENCE album_mix_id
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.album_mix_id OWNER TO band;

--
-- Name: album_mix; Type: TABLE; Schema: public; Owner: band; Tablespace: 
--

CREATE TABLE album_mix (
    album_mix_id integer DEFAULT nextval('album_mix_id'::regclass) NOT NULL,
    album_id integer NOT NULL,
    mix_name character varying
);


ALTER TABLE public.album_mix OWNER TO band;

--
-- Name: news; Type: TABLE; Schema: public; Owner: band; Tablespace: 
--

CREATE TABLE news (
    date date,
    contents character varying
);


ALTER TABLE public.news OWNER TO band;

--
-- Name: page; Type: TABLE; Schema: public; Owner: band; Tablespace: 
--

CREATE TABLE page (
    name character varying,
    title character varying,
    headorder integer,
    disabled boolean DEFAULT false NOT NULL,
    urlpath character varying
);


ALTER TABLE public.page OWNER TO band;

--
-- Name: show; Type: TABLE; Schema: public; Owner: band; Tablespace: 
--

CREATE TABLE show (
    date timestamp without time zone,
    venue character varying,
    notes character varying,
    city character varying
);


ALTER TABLE public.show OWNER TO band;

--
-- Name: song_id; Type: SEQUENCE; Schema: public; Owner: band
--

CREATE SEQUENCE song_id
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.song_id OWNER TO band;

--
-- Name: song; Type: TABLE; Schema: public; Owner: band; Tablespace: 
--

CREATE TABLE song (
    song_id integer DEFAULT nextval('song_id'::regclass) NOT NULL,
    album_mix_id integer NOT NULL,
    title character varying,
    href character varying,
    track integer,
    artist character varying
);


ALTER TABLE public.song OWNER TO band;

--
-- Name: album_mix_pkey; Type: CONSTRAINT; Schema: public; Owner: band; Tablespace: 
--

ALTER TABLE ONLY album_mix
    ADD CONSTRAINT album_mix_pkey PRIMARY KEY (album_mix_id);


--
-- Name: album_pkey; Type: CONSTRAINT; Schema: public; Owner: band; Tablespace: 
--

ALTER TABLE ONLY album
    ADD CONSTRAINT album_pkey PRIMARY KEY (album_id);


--
-- Name: page_headorder_key; Type: CONSTRAINT; Schema: public; Owner: band; Tablespace: 
--

ALTER TABLE ONLY page
    ADD CONSTRAINT page_headorder_key UNIQUE (headorder);


--
-- Name: song_pkey; Type: CONSTRAINT; Schema: public; Owner: band; Tablespace: 
--

ALTER TABLE ONLY song
    ADD CONSTRAINT song_pkey PRIMARY KEY (song_id);


--
-- Name: album_mix_album_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: band
--

ALTER TABLE ONLY album_mix
    ADD CONSTRAINT album_mix_album_id_fkey FOREIGN KEY (album_id) REFERENCES album(album_id);


--
-- Name: song_album_mix_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: band
--

ALTER TABLE ONLY song
    ADD CONSTRAINT song_album_mix_id_fkey FOREIGN KEY (album_mix_id) REFERENCES album_mix(album_mix_id);


--
-- Name: public; Type: ACL; Schema: -; Owner: postgres
--

REVOKE ALL ON SCHEMA public FROM PUBLIC;
REVOKE ALL ON SCHEMA public FROM postgres;
GRANT ALL ON SCHEMA public TO postgres;
GRANT ALL ON SCHEMA public TO PUBLIC;


--
-- Name: news; Type: ACL; Schema: public; Owner: band
--

REVOKE ALL ON TABLE news FROM PUBLIC;
REVOKE ALL ON TABLE news FROM band;
GRANT ALL ON TABLE news TO band;


--
-- PostgreSQL database dump complete
--

