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
-- Name: oneintenband; Type: DATABASE; Schema: -; Owner: postgres
--

CREATE DATABASE oneintenband WITH TEMPLATE = template0 ENCODING = 'UTF8' LC_COLLATE = 'C' LC_CTYPE = 'C';


ALTER DATABASE oneintenband OWNER TO postgres;

\connect oneintenband

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

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: album; Type: TABLE; Schema: public; Owner: www; Tablespace: 
--

CREATE TABLE album (
    album_id integer NOT NULL,
    date date,
    title character varying,
    mix_version integer
);


ALTER TABLE public.album OWNER TO www;

--
-- Name: news; Type: TABLE; Schema: public; Owner: www; Tablespace: 
--

CREATE TABLE news (
    date date,
    contents character varying
);


ALTER TABLE public.news OWNER TO www;

--
-- Name: shows; Type: TABLE; Schema: public; Owner: www; Tablespace: 
--

CREATE TABLE shows (
    date timestamp without time zone,
    venue character varying,
    notes character varying,
    city character varying
);


ALTER TABLE public.shows OWNER TO www;

--
-- Name: song; Type: TABLE; Schema: public; Owner: www; Tablespace: 
--

CREATE TABLE song (
    song_id integer NOT NULL,
    album_id integer,
    track integer,
    title character varying,
    path character varying
);


ALTER TABLE public.song OWNER TO www;

--
-- Name: album_pkey; Type: CONSTRAINT; Schema: public; Owner: www; Tablespace: 
--

ALTER TABLE ONLY album
    ADD CONSTRAINT album_pkey PRIMARY KEY (album_id);


--
-- Name: song_pkey; Type: CONSTRAINT; Schema: public; Owner: www; Tablespace: 
--

ALTER TABLE ONLY song
    ADD CONSTRAINT song_pkey PRIMARY KEY (song_id);


--
-- Name: song_album_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: www
--

ALTER TABLE ONLY song
    ADD CONSTRAINT song_album_id_fkey FOREIGN KEY (album_id) REFERENCES album(album_id);


--
-- Name: public; Type: ACL; Schema: -; Owner: postgres
--

REVOKE ALL ON SCHEMA public FROM PUBLIC;
REVOKE ALL ON SCHEMA public FROM postgres;
GRANT ALL ON SCHEMA public TO postgres;
GRANT ALL ON SCHEMA public TO PUBLIC;


--
-- Name: news; Type: ACL; Schema: public; Owner: www
--

REVOKE ALL ON TABLE news FROM PUBLIC;
REVOKE ALL ON TABLE news FROM www;
GRANT ALL ON TABLE news TO www;


--
-- PostgreSQL database dump complete
--

