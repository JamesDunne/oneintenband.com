--
-- PostgreSQL database dump
--

SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;

SET search_path = public, pg_catalog;

--
-- Data for Name: album; Type: TABLE DATA; Schema: public; Owner: band
--

COPY album (album_id, date, title, description, best_album_mix_id) FROM stdin;
1	2014-08-09	Demo Reel	\N	1
\.


--
-- Name: album_id; Type: SEQUENCE SET; Schema: public; Owner: band
--

SELECT pg_catalog.setval('album_id', 1, false);


--
-- Data for Name: album_mix; Type: TABLE DATA; Schema: public; Owner: band
--

COPY album_mix (album_mix_id, album_id, mix_name) FROM stdin;
1	1	8a
2	1	9c
\.


--
-- Name: album_mix_id; Type: SEQUENCE SET; Schema: public; Owner: band
--

SELECT pg_catalog.setval('album_mix_id', 1, false);


--
-- Data for Name: news; Type: TABLE DATA; Schema: public; Owner: band
--

COPY news (date, contents) FROM stdin;
2014-08-20	Come see our first show!
\.


--
-- Data for Name: page; Type: TABLE DATA; Schema: public; Owner: band
--

COPY page (name, title, headorder, disabled, urlpath) FROM stdin;
index	About	1	f	/
audio	Audio	2	f	/audio
video	Video	3	t	/video
gallery	Gallery	4	t	/gallery
contacts	Contacts	6	t	/contacts
dates	Show Dates	5	f	/dates
\.


--
-- Data for Name: show; Type: TABLE DATA; Schema: public; Owner: band
--

COPY show (date, venue, notes, city) FROM stdin;
2014-08-22 19:30:00	Penny Road Pub	Our first show!	Barrington, IL
\.


--
-- Data for Name: song; Type: TABLE DATA; Schema: public; Owner: band
--

COPY song (song_id, album_mix_id, title, href, track, artist) FROM stdin;
2	1	Song 2	\N	1	Blur
3	1	Brainstew	\N	2	Green Day
4	1	Zero	\N	3	Smashing Pumpkins
5	1	Hash Pipe	\N	4	Weezer
6	1	Plush	\N	5	STP
7	1	Everything Zen	\N	6	Bush
8	2	Song 2	\N	1	Blur
9	2	Brainstew	\N	2	Green Day
10	2	Zero	\N	3	Smashing Pumpkins
11	2	Hash Pipe	\N	4	Weezer
12	2	Plush	\N	5	STP
13	2	Everything Zen	\N	6	Bush
\.


--
-- Name: song_id; Type: SEQUENCE SET; Schema: public; Owner: band
--

SELECT pg_catalog.setval('song_id', 13, true);


--
-- PostgreSQL database dump complete
--

