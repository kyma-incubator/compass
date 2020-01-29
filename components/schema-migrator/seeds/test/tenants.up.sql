--
-- PostgreSQL database dump
--

-- Dumped from database version 11.5 (Debian 11.5-1.pgdg90+1)
-- Dumped by pg_dump version 11.5

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

--
-- Data for Name: integration_systems; Type: TABLE DATA; Schema: public; Owner: usr
--

COPY public.integration_systems (id, name, description) FROM stdin;
\.


--
-- Data for Name: applications; Type: TABLE DATA; Schema: public; Owner: usr
--

COPY public.applications (id, tenant_id, name, description, status_condition, status_timestamp, healthcheck_url, integration_system_id, provider_name) FROM stdin;
\.


--
-- Data for Name: api_definitions; Type: TABLE DATA; Schema: public; Owner: usr
--

COPY public.api_definitions (id, tenant_id, app_id, name, description, group_name, target_url, spec_data, spec_format, spec_type, default_auth, version_value, version_deprecated, version_deprecated_since, version_for_removal) FROM stdin;
\.


--
-- Data for Name: runtimes; Type: TABLE DATA; Schema: public; Owner: usr
--

COPY public.runtimes (id, tenant_id, name, description, status_condition, status_timestamp, creation_timestamp) FROM stdin;
\.


--
-- Data for Name: api_runtime_auths; Type: TABLE DATA; Schema: public; Owner: usr
--

COPY public.api_runtime_auths (id, tenant_id, runtime_id, api_def_id, value) FROM stdin;
\.


--
-- Data for Name: app_templates; Type: TABLE DATA; Schema: public; Owner: usr
--

COPY public.app_templates (id, name, description, application_input, placeholders, access_level) FROM stdin;
\.


--
-- Data for Name: business_tenant_mappings; Type: TABLE DATA; Schema: public; Owner: usr
--

COPY public.business_tenant_mappings (id, external_name, external_tenant, provider_name, status) FROM stdin;
2bf03de1-23b1-4063-9d3e-67096800accc	bazbar	2bf03de1-23b1-4063-9d3e-67096800accc	Compass	Active
3e64ebae-38b5-46a0-b1ed-9ccee153a0ae	foobaz	3e64ebae-38b5-46a0-b1ed-9ccee153a0ae	Compass	Active
\.


--
-- Data for Name: documents; Type: TABLE DATA; Schema: public; Owner: usr
--

COPY public.documents (id, tenant_id, app_id, title, display_name, description, format, kind, data) FROM stdin;
\.


--
-- Data for Name: event_api_definitions; Type: TABLE DATA; Schema: public; Owner: usr
--

COPY public.event_api_definitions (id, tenant_id, app_id, name, description, group_name, spec_data, spec_format, spec_type, version_value, version_deprecated, version_deprecated_since, version_for_removal) FROM stdin;
\.


--
-- Data for Name: fetch_requests; Type: TABLE DATA; Schema: public; Owner: usr
--

COPY public.fetch_requests (id, tenant_id, api_def_id, event_api_def_id, document_id, url, auth, mode, filter, status_condition, status_timestamp) FROM stdin;
\.


--
-- Data for Name: label_definitions; Type: TABLE DATA; Schema: public; Owner: usr
--

COPY public.label_definitions (id, tenant_id, key, schema) FROM stdin;
\.


--
-- Data for Name: labels; Type: TABLE DATA; Schema: public; Owner: usr
--

COPY public.labels (id, tenant_id, app_id, runtime_id, key, value) FROM stdin;
\.


--
-- Data for Name: system_auths; Type: TABLE DATA; Schema: public; Owner: usr
--

COPY public.system_auths (id, tenant_id, app_id, runtime_id, integration_system_id, value) FROM stdin;
\.


--
-- Data for Name: webhooks; Type: TABLE DATA; Schema: public; Owner: usr
--

COPY public.webhooks (id, tenant_id, app_id, url, type, auth) FROM stdin;
\.


--
-- PostgreSQL database dump complete
--

