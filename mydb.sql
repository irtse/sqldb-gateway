--
-- PostgreSQL database dump
--

-- Dumped from database version 16.1
-- Dumped by pg_dump version 16.1

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
-- Name: dbaction_kind; Type: TYPE; Schema: public; Owner: test
--

CREATE TYPE public.dbaction_kind AS ENUM (
    'LINK_SELECT',
    'BUTTON',
    'LINK_ADD'
);


ALTER TYPE public.dbaction_kind OWNER TO test;

--
-- Name: dbaction_method; Type: TYPE; Schema: public; Owner: test
--

CREATE TYPE public.dbaction_method AS ENUM (
    'POST',
    'GET',
    'PUT',
    'DELETE'
);


ALTER TYPE public.dbaction_method OWNER TO test;

--
-- Name: dbaction_type; Type: TYPE; Schema: public; Owner: test
--

CREATE TYPE public.dbaction_type AS ENUM (
    'LINK_SELECT',
    'BUTTON',
    'LINK_ADD'
);


ALTER TYPE public.dbaction_type OWNER TO test;

--
-- Name: dbpermission_read; Type: TYPE; Schema: public; Owner: test
--

CREATE TYPE public.dbpermission_read AS ENUM (
    'admin',
    'moderator',
    'responsible',
    'normal'
);


ALTER TYPE public.dbpermission_read OWNER TO test;

--
-- Name: dbrequest_state; Type: TYPE; Schema: public; Owner: test
--

CREATE TYPE public.dbrequest_state AS ENUM (
    'pending',
    'in progress',
    'dismiss',
    'completed',
    'progressing'
);


ALTER TYPE public.dbrequest_state OWNER TO test;

--
-- Name: dbschema_column_kind; Type: TYPE; Schema: public; Owner: test
--

CREATE TYPE public.dbschema_column_kind AS ENUM (
    'LINK_SELECT',
    'INPUT'
);


ALTER TYPE public.dbschema_column_kind OWNER TO test;

--
-- Name: dbschema_column_read_level; Type: TYPE; Schema: public; Owner: test
--

CREATE TYPE public.dbschema_column_read_level AS ENUM (
    'admin',
    'moderator',
    'responsible',
    'normal'
);


ALTER TYPE public.dbschema_column_read_level OWNER TO test;

--
-- Name: dbtask_assignee_state; Type: TYPE; Schema: public; Owner: test
--

CREATE TYPE public.dbtask_assignee_state AS ENUM (
    'in progress',
    'pending',
    'completed'
);


ALTER TYPE public.dbtask_assignee_state OWNER TO test;

--
-- Name: dbtask_priority; Type: TYPE; Schema: public; Owner: test
--

CREATE TYPE public.dbtask_priority AS ENUM (
    'low',
    'medium',
    'high'
);


ALTER TYPE public.dbtask_priority OWNER TO test;

--
-- Name: dbtask_state; Type: TYPE; Schema: public; Owner: test
--

CREATE TYPE public.dbtask_state AS ENUM (
    'completed',
    'in progress',
    'pending',
    'close',
    'dismiss'
);


ALTER TYPE public.dbtask_state OWNER TO test;

--
-- Name: dbtask_urgency; Type: TYPE; Schema: public; Owner: test
--

CREATE TYPE public.dbtask_urgency AS ENUM (
    'low',
    'medium',
    'high'
);


ALTER TYPE public.dbtask_urgency OWNER TO test;

--
-- Name: dbtask_verifyer_state; Type: TYPE; Schema: public; Owner: test
--

CREATE TYPE public.dbtask_verifyer_state AS ENUM (
    'pending',
    'dismiss',
    'complete'
);


ALTER TYPE public.dbtask_verifyer_state OWNER TO test;

--
-- Name: dbworkflow_schema_priority; Type: TYPE; Schema: public; Owner: test
--

CREATE TYPE public.dbworkflow_schema_priority AS ENUM (
    'low',
    'medium',
    'high'
);


ALTER TYPE public.dbworkflow_schema_priority OWNER TO test;

--
-- Name: dbworkflow_schema_urgency; Type: TYPE; Schema: public; Owner: test
--

CREATE TYPE public.dbworkflow_schema_urgency AS ENUM (
    'low',
    'medium',
    'high'
);


ALTER TYPE public.dbworkflow_schema_urgency OWNER TO test;

--
-- Name: priority; Type: TYPE; Schema: public; Owner: test
--

CREATE TYPE public.priority AS ENUM (
    'low',
    'medium',
    'high'
);


ALTER TYPE public.priority OWNER TO test;

--
-- Name: state; Type: TYPE; Schema: public; Owner: test
--

CREATE TYPE public.state AS ENUM (
    'completed',
    'in progress',
    'pending',
    'close'
);


ALTER TYPE public.state OWNER TO test;

--
-- Name: urgency; Type: TYPE; Schema: public; Owner: test
--

CREATE TYPE public.urgency AS ENUM (
    'low',
    'medium',
    'high'
);


ALTER TYPE public.urgency OWNER TO test;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: confidentiality_level; Type: TABLE; Schema: public; Owner: test
--

CREATE TABLE public.confidentiality_level (
    id integer NOT NULL,
    name character varying(255) NOT NULL
);


ALTER TABLE public.confidentiality_level OWNER TO test;

--
-- Name: confidentiality_level_id_seq; Type: SEQUENCE; Schema: public; Owner: test
--

CREATE SEQUENCE public.confidentiality_level_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.confidentiality_level_id_seq OWNER TO test;

--
-- Name: confidentiality_level_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: test
--

ALTER SEQUENCE public.confidentiality_level_id_seq OWNED BY public.confidentiality_level.id;


--
-- Name: dbentity; Type: TABLE; Schema: public; Owner: test
--

CREATE TABLE public.dbentity (
    id integer NOT NULL,
    name character varying(255),
    parent_id character varying(255),
    description text
);


ALTER TABLE public.dbentity OWNER TO test;

--
-- Name: dbentity_id_seq; Type: SEQUENCE; Schema: public; Owner: test
--

CREATE SEQUENCE public.dbentity_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.dbentity_id_seq OWNER TO test;

--
-- Name: dbentity_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: test
--

ALTER SEQUENCE public.dbentity_id_seq OWNED BY public.dbentity.id;


--
-- Name: dbentity_user; Type: TABLE; Schema: public; Owner: test
--

CREATE TABLE public.dbentity_user (
    id integer NOT NULL,
    dbuser_id integer,
    dbentity_id integer,
    start_date timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    end_date timestamp without time zone
);


ALTER TABLE public.dbentity_user OWNER TO test;

--
-- Name: dbentity_user_id_seq; Type: SEQUENCE; Schema: public; Owner: test
--

CREATE SEQUENCE public.dbentity_user_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.dbentity_user_id_seq OWNER TO test;

--
-- Name: dbentity_user_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: test
--

ALTER SEQUENCE public.dbentity_user_id_seq OWNED BY public.dbentity_user.id;


--
-- Name: dbhierarchy; Type: TABLE; Schema: public; Owner: test
--

CREATE TABLE public.dbhierarchy (
    id integer NOT NULL,
    dbuser_id integer,
    dbentity_id integer,
    parent_dbuser_id integer NOT NULL
);


ALTER TABLE public.dbhierarchy OWNER TO test;

--
-- Name: dbhierarchy_id_seq; Type: SEQUENCE; Schema: public; Owner: test
--

CREATE SEQUENCE public.dbhierarchy_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.dbhierarchy_id_seq OWNER TO test;

--
-- Name: dbhierarchy_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: test
--

ALTER SEQUENCE public.dbhierarchy_id_seq OWNED BY public.dbhierarchy.id;


--
-- Name: dbpermission; Type: TABLE; Schema: public; Owner: test
--

CREATE TABLE public.dbpermission (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    write boolean NOT NULL,
    update boolean NOT NULL,
    delete boolean NOT NULL,
    read character varying(255) DEFAULT 'normal'::character varying NOT NULL
);


ALTER TABLE public.dbpermission OWNER TO test;

--
-- Name: dbpermission_id_seq; Type: SEQUENCE; Schema: public; Owner: test
--

CREATE SEQUENCE public.dbpermission_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.dbpermission_id_seq OWNER TO test;

--
-- Name: dbpermission_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: test
--

ALTER SEQUENCE public.dbpermission_id_seq OWNED BY public.dbpermission.id;


--
-- Name: dbrequest; Type: TABLE; Schema: public; Owner: test
--

CREATE TABLE public.dbrequest (
    id integer NOT NULL,
    dbworkflow_id integer NOT NULL,
    name character varying(255) NOT NULL,
    current_index integer DEFAULT 0 NOT NULL,
    created_date timestamp without time zone DEFAULT '2024-02-20 14:13:26.880174'::timestamp without time zone,
    state public.dbrequest_state DEFAULT 'pending'::public.dbrequest_state NOT NULL,
    is_close boolean DEFAULT false NOT NULL,
    dbdest_table_id integer NOT NULL,
    dbschema_id integer NOT NULL,
    dbcreated_by_id integer NOT NULL
);


ALTER TABLE public.dbrequest OWNER TO test;

--
-- Name: dbrequest_id_seq; Type: SEQUENCE; Schema: public; Owner: test
--

CREATE SEQUENCE public.dbrequest_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.dbrequest_id_seq OWNER TO test;

--
-- Name: dbrequest_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: test
--

ALTER SEQUENCE public.dbrequest_id_seq OWNED BY public.dbrequest.id;


--
-- Name: dbrole; Type: TABLE; Schema: public; Owner: test
--

CREATE TABLE public.dbrole (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    description text DEFAULT 'no description...'::text
);


ALTER TABLE public.dbrole OWNER TO test;

--
-- Name: dbrole_attribution; Type: TABLE; Schema: public; Owner: test
--

CREATE TABLE public.dbrole_attribution (
    id integer NOT NULL,
    dbuser_id integer,
    dbentity_id integer,
    dbrole_id integer NOT NULL,
    start_date timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    end_date timestamp without time zone
);


ALTER TABLE public.dbrole_attribution OWNER TO test;

--
-- Name: dbrole_attribution_id_seq; Type: SEQUENCE; Schema: public; Owner: test
--

CREATE SEQUENCE public.dbrole_attribution_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.dbrole_attribution_id_seq OWNER TO test;

--
-- Name: dbrole_attribution_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: test
--

ALTER SEQUENCE public.dbrole_attribution_id_seq OWNED BY public.dbrole_attribution.id;


--
-- Name: dbrole_id_seq; Type: SEQUENCE; Schema: public; Owner: test
--

CREATE SEQUENCE public.dbrole_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.dbrole_id_seq OWNER TO test;

--
-- Name: dbrole_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: test
--

ALTER SEQUENCE public.dbrole_id_seq OWNED BY public.dbrole.id;


--
-- Name: dbrole_permission; Type: TABLE; Schema: public; Owner: test
--

CREATE TABLE public.dbrole_permission (
    id integer NOT NULL,
    dbrole_id integer NOT NULL,
    dbpermission_id integer NOT NULL
);


ALTER TABLE public.dbrole_permission OWNER TO test;

--
-- Name: dbrole_permission_id_seq; Type: SEQUENCE; Schema: public; Owner: test
--

CREATE SEQUENCE public.dbrole_permission_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.dbrole_permission_id_seq OWNER TO test;

--
-- Name: dbrole_permission_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: test
--

ALTER SEQUENCE public.dbrole_permission_id_seq OWNED BY public.dbrole_permission.id;


--
-- Name: dbschema; Type: TABLE; Schema: public; Owner: test
--

CREATE TABLE public.dbschema (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    label character varying(255) DEFAULT 'general'::character varying
);


ALTER TABLE public.dbschema OWNER TO test;

--
-- Name: dbschema_column; Type: TABLE; Schema: public; Owner: test
--

CREATE TABLE public.dbschema_column (
    id integer NOT NULL,
    dbschema_id integer NOT NULL,
    required boolean,
    read_level public.dbschema_column_read_level DEFAULT 'responsible'::public.dbschema_column_read_level,
    readonly boolean NOT NULL,
    name character varying(255) NOT NULL,
    type character varying(255) NOT NULL,
    index integer DEFAULT 1,
    label character varying(255) NOT NULL,
    placeholder character varying(255),
    default_value character varying(255),
    description character varying(255) DEFAULT 'no description...'::character varying,
    link character varying(255),
    link_sql_dir character varying(255),
    link_sql_order character varying(255),
    link_sql_view character varying(255)
);


ALTER TABLE public.dbschema_column OWNER TO test;

--
-- Name: dbschema_column_id_seq; Type: SEQUENCE; Schema: public; Owner: test
--

CREATE SEQUENCE public.dbschema_column_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.dbschema_column_id_seq OWNER TO test;

--
-- Name: dbschema_column_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: test
--

ALTER SEQUENCE public.dbschema_column_id_seq OWNED BY public.dbschema_column.id;


--
-- Name: dbschema_id_seq; Type: SEQUENCE; Schema: public; Owner: test
--

CREATE SEQUENCE public.dbschema_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.dbschema_id_seq OWNER TO test;

--
-- Name: dbschema_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: test
--

ALTER SEQUENCE public.dbschema_id_seq OWNED BY public.dbschema.id;


--
-- Name: dbtask; Type: TABLE; Schema: public; Owner: test
--

CREATE TABLE public.dbtask (
    id integer NOT NULL,
    dbschema_id integer NOT NULL,
    dbrequest_id integer NOT NULL,
    dbuser_id integer,
    dbentity_id integer,
    dbcreated_by_id integer NOT NULL,
    comment text,
    created_date timestamp without time zone DEFAULT '2024-02-20 14:15:08.176263'::timestamp without time zone NOT NULL,
    state public.dbtask_state DEFAULT 'pending'::public.dbtask_state NOT NULL,
    is_close boolean DEFAULT false,
    urgency public.dbtask_urgency DEFAULT 'medium'::public.dbtask_urgency NOT NULL,
    priority public.dbtask_priority DEFAULT 'medium'::public.dbtask_priority NOT NULL,
    name character varying(255) NOT NULL,
    description character varying(255) DEFAULT 'no description...'::character varying NOT NULL,
    dbworkflow_schema_id integer,
    dbdest_table_id integer NOT NULL
);


ALTER TABLE public.dbtask OWNER TO test;

--
-- Name: dbtask_id_seq; Type: SEQUENCE; Schema: public; Owner: test
--

CREATE SEQUENCE public.dbtask_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.dbtask_id_seq OWNER TO test;

--
-- Name: dbtask_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: test
--

ALTER SEQUENCE public.dbtask_id_seq OWNED BY public.dbtask.id;


--
-- Name: dbuser; Type: TABLE; Schema: public; Owner: test
--

CREATE TABLE public.dbuser (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    email character varying(255) NOT NULL,
    password character varying(255) NOT NULL,
    token character varying(255),
    super_admin boolean NOT NULL
);


ALTER TABLE public.dbuser OWNER TO test;

--
-- Name: dbuser_id_seq; Type: SEQUENCE; Schema: public; Owner: test
--

CREATE SEQUENCE public.dbuser_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.dbuser_id_seq OWNER TO test;

--
-- Name: dbuser_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: test
--

ALTER SEQUENCE public.dbuser_id_seq OWNED BY public.dbuser.id;


--
-- Name: dbview; Type: TABLE; Schema: public; Owner: test
--

CREATE TABLE public.dbview (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    is_list boolean NOT NULL,
    indexable boolean NOT NULL,
    description character varying(255) DEFAULT 'no description...'::character varying,
    readonly boolean NOT NULL,
    index integer DEFAULT 1,
    sql_order character varying(255),
    sql_view character varying(255),
    sql_dir character varying(255),
    through_perms integer,
    dbview_id integer,
    dbschema_id integer,
    is_empty boolean,
    sql_restriction character varying
);


ALTER TABLE public.dbview OWNER TO test;

--
-- Name: dbview_id_seq; Type: SEQUENCE; Schema: public; Owner: test
--

CREATE SEQUENCE public.dbview_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.dbview_id_seq OWNER TO test;

--
-- Name: dbview_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: test
--

ALTER SEQUENCE public.dbview_id_seq OWNED BY public.dbview.id;


--
-- Name: dbworkflow; Type: TABLE; Schema: public; Owner: test
--

CREATE TABLE public.dbworkflow (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    description character varying(255),
    dbschema_id integer
);


ALTER TABLE public.dbworkflow OWNER TO test;

--
-- Name: dbworkflow_id_seq; Type: SEQUENCE; Schema: public; Owner: test
--

CREATE SEQUENCE public.dbworkflow_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.dbworkflow_id_seq OWNER TO test;

--
-- Name: dbworkflow_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: test
--

ALTER SEQUENCE public.dbworkflow_id_seq OWNED BY public.dbworkflow.id;


--
-- Name: dbworkflow_schema; Type: TABLE; Schema: public; Owner: test
--

CREATE TABLE public.dbworkflow_schema (
    id integer NOT NULL,
    dbworkflow_id integer NOT NULL,
    dbschema_id integer NOT NULL,
    index integer DEFAULT 1,
    description text,
    name character varying(255) NOT NULL,
    dbuser_id integer,
    dbentity_id integer,
    urgency public.dbworkflow_schema_urgency DEFAULT 'medium'::public.dbworkflow_schema_urgency,
    priority public.dbworkflow_schema_priority DEFAULT 'medium'::public.dbworkflow_schema_priority
);


ALTER TABLE public.dbworkflow_schema OWNER TO test;

--
-- Name: dbworkflow_schema_id_seq; Type: SEQUENCE; Schema: public; Owner: test
--

CREATE SEQUENCE public.dbworkflow_schema_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.dbworkflow_schema_id_seq OWNER TO test;

--
-- Name: dbworkflow_schema_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: test
--

ALTER SEQUENCE public.dbworkflow_schema_id_seq OWNED BY public.dbworkflow_schema.id;


--
-- Name: formalized_data; Type: TABLE; Schema: public; Owner: test
--

CREATE TABLE public.formalized_data (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    ref character varying(255) NOT NULL,
    capitalization_date timestamp without time zone,
    first_evaluation numeric,
    first_evaluation_date timestamp without time zone,
    actualized_evaluation numeric,
    storage_area character varying(255),
    contractual boolean,
    result_family_id integer,
    result_type_id integer,
    support_id integer,
    confidentiality_level_id integer,
    restriction_type_id integer,
    valuation_id integer
);


ALTER TABLE public.formalized_data OWNER TO test;

--
-- Name: formalized_data_id_seq; Type: SEQUENCE; Schema: public; Owner: test
--

CREATE SEQUENCE public.formalized_data_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.formalized_data_id_seq OWNER TO test;

--
-- Name: formalized_data_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: test
--

ALTER SEQUENCE public.formalized_data_id_seq OWNED BY public.formalized_data.id;


--
-- Name: formalized_data_project; Type: TABLE; Schema: public; Owner: test
--

CREATE TABLE public.formalized_data_project (
    id integer NOT NULL,
    formalized_data_id integer NOT NULL,
    project_id integer NOT NULL
);


ALTER TABLE public.formalized_data_project OWNER TO test;

--
-- Name: formalized_data_project_id_seq; Type: SEQUENCE; Schema: public; Owner: test
--

CREATE SEQUENCE public.formalized_data_project_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.formalized_data_project_id_seq OWNER TO test;

--
-- Name: formalized_data_project_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: test
--

ALTER SEQUENCE public.formalized_data_project_id_seq OWNED BY public.formalized_data_project.id;


--
-- Name: formalized_data_storage_type; Type: TABLE; Schema: public; Owner: test
--

CREATE TABLE public.formalized_data_storage_type (
    id integer NOT NULL,
    formalized_data_id integer NOT NULL,
    storage_type_id integer NOT NULL
);


ALTER TABLE public.formalized_data_storage_type OWNER TO test;

--
-- Name: formalized_data_storage_type_id_seq; Type: SEQUENCE; Schema: public; Owner: test
--

CREATE SEQUENCE public.formalized_data_storage_type_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.formalized_data_storage_type_id_seq OWNER TO test;

--
-- Name: formalized_data_storage_type_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: test
--

ALTER SEQUENCE public.formalized_data_storage_type_id_seq OWNED BY public.formalized_data_storage_type.id;


--
-- Name: project; Type: TABLE; Schema: public; Owner: test
--

CREATE TABLE public.project (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    code character varying(255) NOT NULL,
    dbentity_id integer NOT NULL
);


ALTER TABLE public.project OWNER TO test;

--
-- Name: project_id_seq; Type: SEQUENCE; Schema: public; Owner: test
--

CREATE SEQUENCE public.project_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.project_id_seq OWNER TO test;

--
-- Name: project_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: test
--

ALTER SEQUENCE public.project_id_seq OWNED BY public.project.id;


--
-- Name: protection; Type: TABLE; Schema: public; Owner: test
--

CREATE TABLE public.protection (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    end_date timestamp without time zone NOT NULL,
    start_date timestamp without time zone NOT NULL,
    protection_area_id integer NOT NULL,
    protection_type_id integer NOT NULL,
    formalized_data_id integer
);


ALTER TABLE public.protection OWNER TO test;

--
-- Name: protection_area; Type: TABLE; Schema: public; Owner: test
--

CREATE TABLE public.protection_area (
    id integer NOT NULL,
    name character varying(255) NOT NULL
);


ALTER TABLE public.protection_area OWNER TO test;

--
-- Name: protection_area_id_seq; Type: SEQUENCE; Schema: public; Owner: test
--

CREATE SEQUENCE public.protection_area_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.protection_area_id_seq OWNER TO test;

--
-- Name: protection_area_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: test
--

ALTER SEQUENCE public.protection_area_id_seq OWNED BY public.protection_area.id;


--
-- Name: protection_id_seq; Type: SEQUENCE; Schema: public; Owner: test
--

CREATE SEQUENCE public.protection_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.protection_id_seq OWNER TO test;

--
-- Name: protection_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: test
--

ALTER SEQUENCE public.protection_id_seq OWNED BY public.protection.id;


--
-- Name: protection_type; Type: TABLE; Schema: public; Owner: test
--

CREATE TABLE public.protection_type (
    id integer NOT NULL,
    name character varying(255) NOT NULL
);


ALTER TABLE public.protection_type OWNER TO test;

--
-- Name: protection_type_id_seq; Type: SEQUENCE; Schema: public; Owner: test
--

CREATE SEQUENCE public.protection_type_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.protection_type_id_seq OWNER TO test;

--
-- Name: protection_type_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: test
--

ALTER SEQUENCE public.protection_type_id_seq OWNED BY public.protection_type.id;


--
-- Name: restriction_type; Type: TABLE; Schema: public; Owner: test
--

CREATE TABLE public.restriction_type (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    formalized_data_id integer
);


ALTER TABLE public.restriction_type OWNER TO test;

--
-- Name: restriction_type_id_seq; Type: SEQUENCE; Schema: public; Owner: test
--

CREATE SEQUENCE public.restriction_type_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.restriction_type_id_seq OWNER TO test;

--
-- Name: restriction_type_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: test
--

ALTER SEQUENCE public.restriction_type_id_seq OWNED BY public.restriction_type.id;


--
-- Name: result_family; Type: TABLE; Schema: public; Owner: test
--

CREATE TABLE public.result_family (
    id integer NOT NULL,
    name character varying(255) NOT NULL
);


ALTER TABLE public.result_family OWNER TO test;

--
-- Name: result_family_id_seq; Type: SEQUENCE; Schema: public; Owner: test
--

CREATE SEQUENCE public.result_family_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.result_family_id_seq OWNER TO test;

--
-- Name: result_family_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: test
--

ALTER SEQUENCE public.result_family_id_seq OWNED BY public.result_family.id;


--
-- Name: result_type; Type: TABLE; Schema: public; Owner: test
--

CREATE TABLE public.result_type (
    id integer NOT NULL,
    name character varying(255) NOT NULL
);


ALTER TABLE public.result_type OWNER TO test;

--
-- Name: result_type_id_seq; Type: SEQUENCE; Schema: public; Owner: test
--

CREATE SEQUENCE public.result_type_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.result_type_id_seq OWNER TO test;

--
-- Name: result_type_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: test
--

ALTER SEQUENCE public.result_type_id_seq OWNED BY public.result_type.id;


--
-- Name: sq_dbentity; Type: SEQUENCE; Schema: public; Owner: test
--

CREATE SEQUENCE public.sq_dbentity
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.sq_dbentity OWNER TO test;

--
-- Name: sq_dbform; Type: SEQUENCE; Schema: public; Owner: test
--

CREATE SEQUENCE public.sq_dbform
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.sq_dbform OWNER TO test;

--
-- Name: sq_dbformfields; Type: SEQUENCE; Schema: public; Owner: test
--

CREATE SEQUENCE public.sq_dbformfields
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.sq_dbformfields OWNER TO test;

--
-- Name: sq_dbrole; Type: SEQUENCE; Schema: public; Owner: test
--

CREATE SEQUENCE public.sq_dbrole
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.sq_dbrole OWNER TO test;

--
-- Name: sq_dbtableaccess; Type: SEQUENCE; Schema: public; Owner: test
--

CREATE SEQUENCE public.sq_dbtableaccess
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.sq_dbtableaccess OWNER TO test;

--
-- Name: sq_dbtableview; Type: SEQUENCE; Schema: public; Owner: test
--

CREATE SEQUENCE public.sq_dbtableview
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.sq_dbtableview OWNER TO test;

--
-- Name: sq_dbuser; Type: SEQUENCE; Schema: public; Owner: test
--

CREATE SEQUENCE public.sq_dbuser
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.sq_dbuser OWNER TO test;

--
-- Name: sq_dbuserrole; Type: SEQUENCE; Schema: public; Owner: test
--

CREATE SEQUENCE public.sq_dbuserrole
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.sq_dbuserrole OWNER TO test;

--
-- Name: storage_type; Type: TABLE; Schema: public; Owner: test
--

CREATE TABLE public.storage_type (
    id integer NOT NULL,
    name character varying(255) NOT NULL
);


ALTER TABLE public.storage_type OWNER TO test;

--
-- Name: storage_type_id_seq; Type: SEQUENCE; Schema: public; Owner: test
--

CREATE SEQUENCE public.storage_type_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.storage_type_id_seq OWNER TO test;

--
-- Name: storage_type_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: test
--

ALTER SEQUENCE public.storage_type_id_seq OWNED BY public.storage_type.id;


--
-- Name: support; Type: TABLE; Schema: public; Owner: test
--

CREATE TABLE public.support (
    id integer NOT NULL,
    name character varying(255) NOT NULL
);


ALTER TABLE public.support OWNER TO test;

--
-- Name: support_id_seq; Type: SEQUENCE; Schema: public; Owner: test
--

CREATE SEQUENCE public.support_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.support_id_seq OWNER TO test;

--
-- Name: support_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: test
--

ALTER SEQUENCE public.support_id_seq OWNED BY public.support.id;


--
-- Name: valuation; Type: TABLE; Schema: public; Owner: test
--

CREATE TABLE public.valuation (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    valuation_type_id integer NOT NULL,
    valuation_format_id integer NOT NULL,
    formalized_data_id integer NOT NULL
);


ALTER TABLE public.valuation OWNER TO test;

--
-- Name: valuation_format; Type: TABLE; Schema: public; Owner: test
--

CREATE TABLE public.valuation_format (
    id integer NOT NULL,
    name character varying(255) NOT NULL
);


ALTER TABLE public.valuation_format OWNER TO test;

--
-- Name: valuation_format_id_seq; Type: SEQUENCE; Schema: public; Owner: test
--

CREATE SEQUENCE public.valuation_format_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.valuation_format_id_seq OWNER TO test;

--
-- Name: valuation_format_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: test
--

ALTER SEQUENCE public.valuation_format_id_seq OWNED BY public.valuation_format.id;


--
-- Name: valuation_id_seq; Type: SEQUENCE; Schema: public; Owner: test
--

CREATE SEQUENCE public.valuation_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.valuation_id_seq OWNER TO test;

--
-- Name: valuation_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: test
--

ALTER SEQUENCE public.valuation_id_seq OWNED BY public.valuation.id;


--
-- Name: valuation_type; Type: TABLE; Schema: public; Owner: test
--

CREATE TABLE public.valuation_type (
    id integer NOT NULL,
    name character varying(255) NOT NULL
);


ALTER TABLE public.valuation_type OWNER TO test;

--
-- Name: valuation_type_id_seq; Type: SEQUENCE; Schema: public; Owner: test
--

CREATE SEQUENCE public.valuation_type_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.valuation_type_id_seq OWNER TO test;

--
-- Name: valuation_type_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: test
--

ALTER SEQUENCE public.valuation_type_id_seq OWNED BY public.valuation_type.id;


--
-- Name: confidentiality_level id; Type: DEFAULT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.confidentiality_level ALTER COLUMN id SET DEFAULT nextval('public.confidentiality_level_id_seq'::regclass);


--
-- Name: dbentity id; Type: DEFAULT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbentity ALTER COLUMN id SET DEFAULT nextval('public.dbentity_id_seq'::regclass);


--
-- Name: dbentity_user id; Type: DEFAULT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbentity_user ALTER COLUMN id SET DEFAULT nextval('public.dbentity_user_id_seq'::regclass);


--
-- Name: dbhierarchy id; Type: DEFAULT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbhierarchy ALTER COLUMN id SET DEFAULT nextval('public.dbhierarchy_id_seq'::regclass);


--
-- Name: dbpermission id; Type: DEFAULT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbpermission ALTER COLUMN id SET DEFAULT nextval('public.dbpermission_id_seq'::regclass);


--
-- Name: dbrequest id; Type: DEFAULT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbrequest ALTER COLUMN id SET DEFAULT nextval('public.dbrequest_id_seq'::regclass);


--
-- Name: dbrole id; Type: DEFAULT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbrole ALTER COLUMN id SET DEFAULT nextval('public.dbrole_id_seq'::regclass);


--
-- Name: dbrole_attribution id; Type: DEFAULT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbrole_attribution ALTER COLUMN id SET DEFAULT nextval('public.dbrole_attribution_id_seq'::regclass);


--
-- Name: dbrole_permission id; Type: DEFAULT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbrole_permission ALTER COLUMN id SET DEFAULT nextval('public.dbrole_permission_id_seq'::regclass);


--
-- Name: dbschema id; Type: DEFAULT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbschema ALTER COLUMN id SET DEFAULT nextval('public.dbschema_id_seq'::regclass);


--
-- Name: dbschema_column id; Type: DEFAULT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbschema_column ALTER COLUMN id SET DEFAULT nextval('public.dbschema_column_id_seq'::regclass);


--
-- Name: dbtask id; Type: DEFAULT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbtask ALTER COLUMN id SET DEFAULT nextval('public.dbtask_id_seq'::regclass);


--
-- Name: dbuser id; Type: DEFAULT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbuser ALTER COLUMN id SET DEFAULT nextval('public.dbuser_id_seq'::regclass);


--
-- Name: dbview id; Type: DEFAULT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbview ALTER COLUMN id SET DEFAULT nextval('public.dbview_id_seq'::regclass);


--
-- Name: dbworkflow id; Type: DEFAULT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbworkflow ALTER COLUMN id SET DEFAULT nextval('public.dbworkflow_id_seq'::regclass);


--
-- Name: dbworkflow_schema id; Type: DEFAULT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbworkflow_schema ALTER COLUMN id SET DEFAULT nextval('public.dbworkflow_schema_id_seq'::regclass);


--
-- Name: formalized_data id; Type: DEFAULT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.formalized_data ALTER COLUMN id SET DEFAULT nextval('public.formalized_data_id_seq'::regclass);


--
-- Name: formalized_data_project id; Type: DEFAULT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.formalized_data_project ALTER COLUMN id SET DEFAULT nextval('public.formalized_data_project_id_seq'::regclass);


--
-- Name: formalized_data_storage_type id; Type: DEFAULT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.formalized_data_storage_type ALTER COLUMN id SET DEFAULT nextval('public.formalized_data_storage_type_id_seq'::regclass);


--
-- Name: project id; Type: DEFAULT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.project ALTER COLUMN id SET DEFAULT nextval('public.project_id_seq'::regclass);


--
-- Name: protection id; Type: DEFAULT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.protection ALTER COLUMN id SET DEFAULT nextval('public.protection_id_seq'::regclass);


--
-- Name: protection_area id; Type: DEFAULT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.protection_area ALTER COLUMN id SET DEFAULT nextval('public.protection_area_id_seq'::regclass);


--
-- Name: protection_type id; Type: DEFAULT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.protection_type ALTER COLUMN id SET DEFAULT nextval('public.protection_type_id_seq'::regclass);


--
-- Name: restriction_type id; Type: DEFAULT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.restriction_type ALTER COLUMN id SET DEFAULT nextval('public.restriction_type_id_seq'::regclass);


--
-- Name: result_family id; Type: DEFAULT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.result_family ALTER COLUMN id SET DEFAULT nextval('public.result_family_id_seq'::regclass);


--
-- Name: result_type id; Type: DEFAULT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.result_type ALTER COLUMN id SET DEFAULT nextval('public.result_type_id_seq'::regclass);


--
-- Name: storage_type id; Type: DEFAULT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.storage_type ALTER COLUMN id SET DEFAULT nextval('public.storage_type_id_seq'::regclass);


--
-- Name: support id; Type: DEFAULT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.support ALTER COLUMN id SET DEFAULT nextval('public.support_id_seq'::regclass);


--
-- Name: valuation id; Type: DEFAULT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.valuation ALTER COLUMN id SET DEFAULT nextval('public.valuation_id_seq'::regclass);


--
-- Name: valuation_format id; Type: DEFAULT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.valuation_format ALTER COLUMN id SET DEFAULT nextval('public.valuation_format_id_seq'::regclass);


--
-- Name: valuation_type id; Type: DEFAULT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.valuation_type ALTER COLUMN id SET DEFAULT nextval('public.valuation_type_id_seq'::regclass);


--
-- Data for Name: confidentiality_level; Type: TABLE DATA; Schema: public; Owner: test
--

COPY public.confidentiality_level (id, name) FROM stdin;
1	public
2	confidential project
3	IRT confidential
4	restricted diffusion
5	special restricted diffusion
6	classified datas
7	authorized profile
\.


--
-- Data for Name: dbentity; Type: TABLE DATA; Schema: public; Owner: test
--

COPY public.dbentity (id, name, parent_id, description) FROM stdin;
1	researcher	\N	researcher ingeneer && PF researcher common permission
2	CDP	\N	\N
3	CDG	\N	\N
4	RH	\N	\N
5	JURIDIC	\N	\N
\.


--
-- Data for Name: dbentity_user; Type: TABLE DATA; Schema: public; Owner: test
--

COPY public.dbentity_user (id, dbuser_id, dbentity_id, start_date, end_date) FROM stdin;
1	3	2	2024-02-15 08:53:02.562869	\N
2	2	1	2024-02-15 08:53:02.562869	\N
3	4	3	2024-02-15 08:53:02.562869	\N
4	5	4	2024-02-21 11:09:09.258162	\N
5	6	5	2024-02-28 09:50:46.138578	\N
\.


--
-- Data for Name: dbhierarchy; Type: TABLE DATA; Schema: public; Owner: test
--

COPY public.dbhierarchy (id, dbuser_id, dbentity_id, parent_dbuser_id) FROM stdin;
1	2	\N	3
\.


--
-- Data for Name: dbpermission; Type: TABLE DATA; Schema: public; Owner: test
--

COPY public.dbpermission (id, name, write, update, delete, read) FROM stdin;
1964	result_family:creator	t	f	f	normal
1974	storage_type:admin	t	t	t	normal
1995	project:reader	f	f	f	normal
2058	protection_type:reader	f	f	f	normal
2121	formalized_data:project_id:normal:manager	t	t	f	normal
2163	formalized_data:confidentiality_level_id:normal:updater	f	t	f	normal
2205	formalized_data:valuation_id:normal:updater	f	t	f	normal
2259	dbrequest:name:normal:updater	f	t	f	normal
2269	dbhierarchy:admin	t	t	t	normal
2719	formalized_data_project:manager	t	t	f	normal
2311	dbtask:dbrequest_id:<nil>:updater	f	t	f	<nil>
2332	dbtask:dbentity_id:normal:admin	t	t	t	normal
2740	formalized_data_project:formalized_data_id:normal:creator	t	f	f	normal
2761	formalized_data_storage_type:formalized_data_id:normal:updater	f	t	f	normal
1894	formalized_data:first_evaluation:normal:updater	f	t	f	normal
1936	formalized_data:contractual:normal:creator	t	f	f	normal
1196	dbschema_column:type:normal:admin	t	t	t	normal
1197	dbschema_column:type:normal:manager	t	t	f	normal
1198	dbschema_column:type:normal:creator	t	f	f	normal
1199	dbschema_column:index:normal:admin	t	t	t	normal
1200	dbschema_column:index:normal:manager	t	t	f	normal
1201	dbschema_column:index:normal:creator	t	f	f	normal
1202	dbschema_column:index:normal:updater	f	t	f	normal
1203	dbschema_column:index:normal:reader	f	f	f	normal
1411	dbview:creator	t	f	f	normal
1450	dbview:sql_view:normal:manager	t	t	f	normal
1451	dbview:sql_view:normal:creator	t	f	f	normal
1452	dbview:sql_view:normal:updater	f	t	f	normal
1561	dbtask:dbopened_by_id:normal:admin	t	t	t	normal
1651	dbworkflow_schema:admin	t	t	t	normal
1229	dbschema_column:link:responsible:creator	t	f	f	normal
1230	dbschema_column:link:normal:updater	f	t	f	normal
1231	dbschema_column:link:responsible:updater	f	t	f	normal
2799	formalized_data:actualized_evaluation:responsible:manager	t	t	f	responsible
2542	dbrequest:dbschema_id:normal:manager	t	t	f	normal
2626	dbtask:comment:<nil>:manager	t	t	f	<nil>
2668	dbtask:urgency:<nil>:creator	t	f	f	<nil>
2710	dbtask:dbdest_table_id:responsible:updater	f	t	f	responsible
1944	support:admin	t	t	t	normal
1985	confidentiality_level:updater	f	t	f	normal
1372	dbentity:dbentity_user:normal:manager	t	t	f	normal
1547	dbworkflow:dbworkflow_schema:normal:updater	f	t	f	normal
1742	dbrole_permission:dbrole_id:normal:reader	f	f	f	normal
1743	dbrole_permission:dbrole_id:normal:admin	t	t	t	normal
1744	dbrole_permission:dbpermission_id:normal:updater	f	t	f	normal
1745	dbrole_permission:dbpermission_id:normal:reader	f	f	f	normal
1746	dbrole_permission:dbpermission_id:normal:admin	t	t	t	normal
1749	formalized_data:admin	t	t	t	normal
1884	formalized_data:capitalization_date:normal:admin	t	t	t	normal
2225	dbrequest:current_index:normal:manager	t	t	f	normal
2048	protection:start_date:normal:creator	t	f	f	normal
1994	project:updater	f	t	f	normal
1216	dbschema_column:default_value:normal:manager	t	t	f	normal
1217	dbschema_column:default_value:normal:creator	t	f	f	normal
1218	dbschema_column:default_value:normal:updater	f	t	f	normal
1622	dbtask:header:normal:updater	f	t	f	normal
1628	dbtask:dbdest_table_id:normal:updater	f	t	f	normal
1629	dbtask:dbworkflow_id:normal:admin	t	t	t	normal
1352	dbentity:creator	t	f	f	normal
1356	dbentity:name:normal:creator	t	f	f	normal
1357	dbentity:name:normal:updater	f	t	f	normal
1410	dbview:manager	t	t	f	normal
1614	dbtask:description:normal:admin	t	t	t	normal
1316	dbpermission:reader	f	f	f	normal
1432	dbview:description:normal:updater	f	t	f	normal
1433	dbview:description:normal:reader	f	f	f	normal
1434	dbview:readonly:normal:admin	t	t	t	normal
1435	dbview:readonly:normal:manager	t	t	f	normal
1226	dbschema_column:link:normal:manager	t	t	f	normal
1358	dbentity:name:normal:reader	f	f	f	normal
1319	dbpermission:name:normal:admin	t	t	t	normal
1320	dbpermission:name:normal:manager	t	t	f	normal
1321	dbpermission:name:normal:creator	t	f	f	normal
1946	support:creator	t	f	f	normal
2232	dbrequest:dbworkflow_id:normal:admin	t	t	t	normal
1456	dbview:sql_dir:normal:reader	f	f	f	normal
1457	dbview:sql_dir:normal:admin	t	t	t	normal
1458	dbview:sql_dir:normal:manager	t	t	f	normal
2233	dbrequest:dbworkflow_id:normal:manager	t	t	f	normal
1630	dbtask:dbworkflow_id:normal:manager	t	t	f	normal
1715	dbtask_watcher:manager	t	t	f	normal
1716	dbtask_watcher:creator	t	f	f	normal
1225	dbschema_column:link:responsible:admin	t	t	t	normal
1673	dbtask_assignee:reader	f	f	f	normal
1499	dbrole_attribution:reader	f	f	f	normal
2720	formalized_data_project:creator	t	f	f	normal
2270	dbhierarchy:manager	t	t	f	normal
2312	dbtask:dbrequest_id:normal:reader	f	f	f	normal
2333	dbtask:dbentity_id:<nil>:admin	t	t	t	<nil>
2783	dbschema_column:link_is_enum:normal:creator	t	f	f	normal
2800	formalized_data:capitalization_date:responsible:manager	t	t	f	responsible
2438	dbrequest:name:<nil>:reader	f	f	f	<nil>
2501	dbrequest:is_close:responsible:manager	t	t	f	responsible
2543	dbrequest:dbschema_id:responsible:manager	t	t	f	responsible
2648	dbtask:state:<nil>:admin	t	t	t	<nil>
2690	dbtask:description:<nil>:admin	t	t	t	<nil>
1500	dbrole_attribution:admin	t	t	t	normal
1727	dbtask_watcher:dbtask_id:normal:creator	t	f	f	normal
1729	dbtask_watcher:dbentity_id:normal:reader	f	f	f	normal
1730	dbtask_watcher:dbentity_id:normal:admin	t	t	t	normal
1731	dbtask_watcher:dbentity_id:normal:manager	t	t	f	normal
1732	dbtask_watcher:dbentity_id:normal:creator	t	f	f	normal
1733	dbtask_watcher:dbentity_id:normal:updater	f	t	f	normal
1735	dbrole_permission:manager	t	t	f	normal
1736	dbrole_permission:creator	t	f	f	normal
1737	dbrole_permission:updater	f	t	f	normal
1738	dbrole_permission:reader	f	f	f	normal
1739	dbrole_permission:dbrole_id:normal:manager	t	t	f	normal
1623	dbtask:header:normal:reader	f	f	f	normal
1626	dbtask:dbdest_table_id:normal:manager	t	t	f	normal
1627	dbtask:dbdest_table_id:normal:creator	t	f	f	normal
1422	dbview:is_list:normal:reader	f	f	f	normal
1423	dbview:is_list:normal:admin	t	t	t	normal
1424	dbview:indexable:normal:updater	f	t	f	normal
1425	dbview:indexable:normal:reader	f	f	f	normal
1277	dbuser:email:normal:admin	t	t	t	normal
2087	valuation:manager	t	t	f	normal
2129	formalized_data:result_family_id:normal:admin	t	t	t	normal
2171	formalized_data:storage_type_id:normal:admin	t	t	t	normal
1636	dbtask:dbtask_assignee:normal:creator	t	f	f	normal
1637	dbtask:dbtask_assignee:normal:updater	f	t	f	normal
1638	dbtask:dbtask_assignee:normal:reader	f	f	f	normal
1965	result_family:updater	f	t	f	normal
1975	storage_type:manager	t	t	f	normal
1996	project:admin	t	t	t	normal
2017	project:dbentity_id:normal:reader	f	f	f	normal
2038	protection:end_date:normal:creator	t	f	f	normal
2059	protection_area:admin	t	t	t	normal
2080	protection:protection_type_id:normal:updater	f	t	f	normal
2101	valuation:valuation_type:normal:manager	t	t	f	normal
2143	formalized_data:result_type_id:normal:creator	t	f	f	normal
2185	formalized_data:restriction_type_id:normal:updater	f	t	f	normal
1916	formalized_data:actualized_evaluation:normal:manager	t	t	f	normal
1958	support:name:normal:creator	t	f	f	normal
1478	dbentity_user:reader	f	f	f	normal
1479	dbentity_user:dbuser_id:normal:updater	f	t	f	normal
1480	dbentity_user:dbuser_id:normal:reader	f	f	f	normal
1966	result_family:reader	f	f	f	normal
1976	storage_type:creator	t	f	f	normal
2060	protection_area:manager	t	t	f	normal
2123	formalized_data:project_id:normal:creator	t	f	f	normal
2165	formalized_data:confidentiality_level_id:normal:reader	f	f	f	normal
1466	dbview:dbview_id:normal:manager	t	t	f	normal
1467	dbview:dbview_id:normal:creator	t	f	f	normal
1468	dbview:dbview_id:normal:updater	f	t	f	normal
1469	dbview:dbschema_id:normal:updater	f	t	f	normal
2261	dbrequest:name:normal:reader	f	f	f	normal
2271	dbhierarchy:creator	t	f	f	normal
2721	formalized_data_project:updater	f	t	f	normal
2313	dbtask:dbrequest_id:<nil>:reader	f	f	f	<nil>
2742	formalized_data_project:formalized_data_id:normal:updater	f	t	f	normal
2763	formalized_data_storage_type:formalized_data_id:normal:reader	f	f	f	normal
2801	formalized_data:first_evaluation_date:responsible:manager	t	t	f	responsible
2460	dbrequest:created_date:<nil>:manager	t	t	f	<nil>
2544	dbrequest:dbcreated_by_id:normal:admin	t	t	t	normal
2628	dbtask:comment:<nil>:creator	t	f	f	<nil>
2649	dbtask:is_close:normal:creator	t	f	f	normal
2670	dbtask:priority:<nil>:admin	t	t	t	<nil>
2712	dbtask:dbdest_table_id:responsible:reader	f	f	f	responsible
1472	dbview:dbschema_id:normal:manager	t	t	f	normal
1473	dbview:dbschema_id:normal:creator	t	f	f	normal
1475	dbentity_user:manager	t	t	f	normal
1476	dbentity_user:creator	t	f	f	normal
1477	dbentity_user:updater	f	t	f	normal
1501	dbrole_attribution:manager	t	t	f	normal
1902	formalized_data:first_evaluation:normal:creator	t	f	f	normal
1302	dbuser:dbentity_user:normal:manager	t	t	f	normal
1309	dbuser:dbhierarchy:normal:manager	t	t	f	normal
2022	protection:updater	f	t	f	normal
2064	protection:protection_area_id:normal:updater	f	t	f	normal
2226	dbrequest:current_index:normal:creator	t	f	f	normal
1280	dbuser:password:responsible:admin	t	t	t	normal
1281	dbuser:password:normal:manager	t	t	f	normal
1282	dbuser:password:responsible:manager	t	t	f	normal
1283	dbuser:password:normal:creator	t	f	f	normal
1284	dbuser:password:responsible:creator	t	f	f	normal
1285	dbuser:password:normal:updater	f	t	f	normal
1204	dbschema_column:label:normal:updater	f	t	f	normal
1205	dbschema_column:label:normal:reader	f	f	f	normal
1206	dbschema_column:label:normal:admin	t	t	t	normal
1207	dbschema_column:label:normal:manager	t	t	f	normal
1208	dbschema_column:label:normal:creator	t	f	f	normal
1209	dbschema_column:placeholder:normal:updater	f	t	f	normal
1210	dbschema_column:placeholder:normal:reader	f	f	f	normal
1211	dbschema_column:placeholder:normal:admin	t	t	t	normal
1212	dbschema_column:placeholder:normal:manager	t	t	f	normal
1213	dbschema_column:placeholder:normal:creator	t	f	f	normal
1214	dbschema_column:default_value:normal:reader	f	f	f	normal
1597	dbtask:state:normal:creator	t	f	f	normal
1598	dbtask:state:normal:updater	f	t	f	normal
1599	dbtask:urgency:normal:admin	t	t	t	normal
1600	dbtask:urgency:normal:manager	t	t	f	normal
1601	dbtask:urgency:normal:creator	t	f	f	normal
1602	dbtask:urgency:normal:updater	f	t	f	normal
1610	dbtask:name:normal:manager	t	t	f	normal
1611	dbtask:name:normal:creator	t	f	f	normal
1612	dbtask:name:normal:updater	f	t	f	normal
1613	dbtask:name:normal:reader	f	f	f	normal
1415	dbview:name:normal:updater	f	t	f	normal
1416	dbview:name:normal:reader	f	f	f	normal
1453	dbview:sql_view:normal:reader	f	f	f	normal
1454	dbview:sql_dir:normal:creator	t	f	f	normal
2722	formalized_data_project:reader	f	f	f	normal
2272	dbhierarchy:updater	f	t	f	normal
2314	dbtask:dbuser_id:normal:admin	t	t	t	normal
2785	dbschema_column:link_is_enum:normal:updater	f	t	f	normal
2802	formalized_data:restriction_types:responsible:manager	t	t	f	responsible
2440	dbrequest:current_index:<nil>:admin	t	t	t	<nil>
2503	dbrequest:is_close:responsible:creator	t	f	f	responsible
2524	dbrequest:dbdest_table_id:normal:creator	t	f	f	normal
2545	dbrequest:dbcreated_by_id:responsible:admin	t	t	t	responsible
2650	dbtask:is_close:responsible:creator	t	f	f	responsible
2692	dbtask:description:<nil>:manager	t	t	f	<nil>
1455	dbview:sql_dir:normal:updater	f	t	f	normal
1272	dbuser:name:normal:updater	f	t	f	normal
1273	dbuser:name:normal:reader	f	f	f	normal
1274	dbuser:email:normal:creator	t	f	f	normal
1575	dbtask:opened_date:responsible:creator	t	f	f	normal
1576	dbtask:opened_date:normal:updater	f	t	f	normal
1577	dbtask:opened_date:responsible:updater	f	t	f	normal
1578	dbtask:opened_date:normal:reader	f	f	f	normal
1193	dbschema_column:name:normal:reader	f	f	f	normal
1337	dbpermission:delete:normal:admin	t	t	t	normal
2082	protection:protection_type_id:normal:reader	f	f	f	normal
2103	valuation:valuation_type:normal:creator	t	f	f	normal
2145	formalized_data:result_type_id:normal:updater	f	t	f	normal
2187	formalized_data:restriction_type_id:normal:reader	f	f	f	normal
1918	formalized_data:actualized_evaluation:normal:creator	t	f	f	normal
1876	formalized_data:ref:normal:manager	t	t	f	normal
1906	formalized_data:first_evaluation_date:normal:manager	t	t	f	normal
1948	support:reader	f	f	f	normal
1385	dbrole:admin	t	t	t	normal
1386	dbrole:manager	t	t	f	normal
1387	dbrole:creator	t	f	f	normal
2070	protection:protection_area_id:normal:manager	t	t	f	normal
2091	valuation_type:creator	t	f	f	normal
2133	formalized_data:result_family_id:normal:creator	t	f	f	normal
2175	formalized_data:storage_type_id:normal:creator	t	f	f	normal
1502	dbrole_attribution:creator	t	f	f	normal
1503	dbrole_attribution:updater	f	t	f	normal
1504	dbrole_attribution:dbuser_id:normal:admin	t	t	t	normal
1505	dbrole_attribution:dbuser_id:normal:manager	t	t	f	normal
1389	dbrole:name:normal:admin	t	t	t	normal
1390	dbrole:name:normal:manager	t	t	f	normal
1674	dbtask_assignee:dbuser_id:normal:admin	t	t	t	normal
1675	dbtask_assignee:dbuser_id:normal:manager	t	t	f	normal
1417	dbview:name:normal:admin	t	t	t	normal
1418	dbview:name:normal:manager	t	t	f	normal
1227	dbschema_column:link:responsible:manager	t	t	f	normal
1228	dbschema_column:link:normal:creator	t	f	f	normal
1886	formalized_data:capitalization_date:normal:manager	t	t	f	normal
1325	dbpermission:write:normal:admin	t	t	t	normal
1326	dbpermission:write:normal:manager	t	t	f	normal
1327	dbpermission:write:normal:creator	t	f	f	normal
1940	formalized_data:contractual:normal:reader	f	f	f	normal
1546	dbworkflow:dbworkflow_schema:normal:creator	t	f	f	normal
2090	valuation_type:manager	t	t	f	normal
2111	valuation:valuation_format:normal:manager	t	t	f	normal
2153	formalized_data:support_id:normal:admin	t	t	t	normal
2195	formalized_data:protection_id:normal:updater	f	t	f	normal
2263	dbrequest:name:normal:admin	t	t	t	normal
2273	dbhierarchy:reader	f	f	f	normal
2723	formalized_data_project:admin	t	t	t	normal
2315	dbtask:dbuser_id:<nil>:admin	t	t	t	<nil>
2744	formalized_data_project:project_id:normal:admin	t	t	t	normal
2765	formalized_data_storage_type:formalized_data_id:normal:admin	t	t	t	normal
2803	formalized_data:valuations:responsible:manager	t	t	f	responsible
2462	dbrequest:created_date:<nil>:creator	t	f	f	<nil>
2525	dbrequest:dbdest_table_id:responsible:creator	t	f	f	responsible
2546	dbrequest:dbcreated_by_id:normal:manager	t	t	f	normal
2630	dbtask:created_date:<nil>:reader	f	f	f	<nil>
2651	dbtask:is_close:normal:updater	f	t	f	normal
2672	dbtask:priority:<nil>:manager	t	t	f	<nil>
2714	dbtask:dbdest_table_id:responsible:admin	t	t	t	responsible
1408	dbrole:dbrole_attribution:normal:reader	f	f	f	normal
1438	dbview:readonly:normal:reader	f	f	f	normal
1680	dbtask_assignee:dbtask_id:normal:creator	t	f	f	normal
1681	dbtask_assignee:dbtask_id:normal:updater	f	t	f	normal
1682	dbtask_assignee:dbtask_id:normal:reader	f	f	f	normal
1683	dbtask_assignee:dbtask_id:normal:admin	t	t	t	normal
1684	dbtask_assignee:dbentity_id:normal:admin	t	t	t	normal
1685	dbtask_assignee:dbentity_id:normal:manager	t	t	f	normal
1580	dbtask:opened_date:normal:admin	t	t	t	normal
1581	dbtask:opened_date:responsible:admin	t	t	t	normal
1582	dbtask:opened_date:normal:manager	t	t	f	normal
1583	dbtask:opened_date:responsible:manager	t	t	f	normal
1584	dbtask:comment:normal:admin	t	t	t	normal
1585	dbtask:comment:normal:manager	t	t	f	normal
1586	dbtask:comment:normal:creator	t	f	f	normal
1587	dbtask:comment:normal:updater	f	t	f	normal
1914	formalized_data:actualized_evaluation:normal:admin	t	t	t	normal
2015	project:dbentity_id:normal:updater	f	t	f	normal
2036	protection:end_date:normal:manager	t	t	f	normal
2057	protection_type:updater	f	t	f	normal
2078	protection:protection_type_id:normal:creator	t	f	f	normal
2099	valuation:valuation_type:normal:admin	t	t	t	normal
2141	formalized_data:result_type_id:normal:manager	t	t	f	normal
2183	formalized_data:restriction_type_id:normal:creator	t	f	f	normal
1956	support:name:normal:manager	t	t	f	normal
1718	dbtask_watcher:reader	f	f	f	normal
1719	dbtask_watcher:dbuser_id:normal:updater	f	t	f	normal
1720	dbtask_watcher:dbuser_id:normal:reader	f	f	f	normal
1721	dbtask_watcher:dbuser_id:normal:admin	t	t	t	normal
1722	dbtask_watcher:dbuser_id:normal:manager	t	t	f	normal
1723	dbtask_watcher:dbuser_id:normal:creator	t	f	f	normal
1724	dbtask_watcher:dbtask_id:normal:reader	f	f	f	normal
1725	dbtask_watcher:dbtask_id:normal:admin	t	t	t	normal
1726	dbtask_watcher:dbtask_id:normal:manager	t	t	f	normal
1740	dbrole_permission:dbrole_id:normal:creator	t	f	f	normal
1741	dbrole_permission:dbrole_id:normal:updater	f	t	f	normal
1548	dbworkflow:dbworkflow_schema:normal:reader	f	f	f	normal
1194	dbschema_column:type:normal:updater	f	t	f	normal
2724	protection_id:creator	t	f	f	normal
2295	dbtask:dbschema_id:<nil>:creator	t	f	f	<nil>
2316	dbtask:dbuser_id:normal:manager	t	t	f	normal
2787	dbschema_column:link_is_enum:normal:reader	f	f	f	normal
2442	dbrequest:current_index:<nil>:manager	t	t	f	<nil>
2526	dbrequest:dbdest_table_id:normal:updater	f	t	f	normal
2547	dbrequest:dbcreated_by_id:responsible:manager	t	t	f	responsible
2610	dbtask:dbcreated_by_id:responsible:updater	f	t	f	responsible
2652	dbtask:is_close:responsible:updater	f	t	f	responsible
2694	dbtask:description:<nil>:creator	t	f	f	<nil>
1888	formalized_data:capitalization_date:normal:creator	t	f	f	normal
1978	storage_type:reader	f	f	f	normal
1999	project:code:normal:admin	t	t	t	normal
2020	protection:manager	t	t	f	normal
2062	protection_area:updater	f	t	f	normal
2125	formalized_data:project_id:normal:updater	f	t	f	normal
2167	formalized_data:confidentiality_level_id:normal:admin	t	t	t	normal
1439	dbview:index:normal:admin	t	t	t	normal
1609	dbtask:name:normal:admin	t	t	t	normal
1981	result_type:reader	f	f	f	normal
2023	protection:reader	f	f	f	normal
2044	protection:start_date:normal:admin	t	t	t	normal
2086	valuation:admin	t	t	t	normal
2107	valuation:valuation_type:normal:reader	f	f	f	normal
2149	formalized_data:support_id:normal:updater	f	t	f	normal
1621	dbtask:header:normal:creator	t	f	f	normal
1631	dbtask:dbworkflow_id:normal:creator	t	f	f	normal
1632	dbtask:dbworkflow_id:normal:updater	f	t	f	normal
1633	dbtask:dbworkflow_id:normal:reader	f	f	f	normal
1634	dbtask:dbtask_assignee:normal:admin	t	t	t	normal
2003	project:code:normal:creator	t	f	f	normal
1635	dbtask:dbtask_assignee:normal:manager	t	t	f	normal
2024	start_date:admin	t	t	t	normal
2066	protection:protection_area_id:normal:reader	f	f	f	normal
1668	dbworkflow_schema:index:normal:admin	t	t	t	normal
1669	dbtask_assignee:admin	t	t	t	normal
1670	dbtask_assignee:manager	t	t	f	normal
1671	dbtask_assignee:creator	t	f	f	normal
1672	dbtask_assignee:updater	f	t	f	normal
1419	dbview:is_list:normal:manager	t	t	f	normal
1420	dbview:is_list:normal:creator	t	f	f	normal
1421	dbview:is_list:normal:updater	f	t	f	normal
1562	dbtask:dbopened_by_id:responsible:admin	t	t	t	normal
1563	dbtask:dbopened_by_id:normal:manager	t	t	f	normal
1639	dbtask:dbtask_verifyer:normal:admin	t	t	t	normal
1640	dbtask:dbtask_verifyer:normal:manager	t	t	f	normal
1641	dbtask:dbtask_verifyer:normal:creator	t	f	f	normal
1643	dbtask:dbtask_verifyer:normal:reader	f	f	f	normal
2265	dbrequest:name:normal:manager	t	t	f	normal
2725	protection_id:updater	f	t	f	normal
2746	formalized_data_project:project_id:normal:manager	t	t	f	normal
2317	dbtask:dbuser_id:<nil>:manager	t	t	f	<nil>
2767	formalized_data_storage_type:formalized_data_id:normal:manager	t	t	f	normal
2805	formalized_data:confidentiality_level_id:responsible:manager	t	t	f	responsible
2464	dbrequest:created_date:<nil>:updater	f	t	f	<nil>
2485	dbrequest:state:responsible:reader	f	f	f	responsible
2527	dbrequest:dbdest_table_id:responsible:updater	f	t	f	responsible
2548	dbrequest:dbcreated_by_id:normal:creator	t	f	f	normal
2632	dbtask:created_date:<nil>:admin	t	t	t	<nil>
2653	dbtask:is_close:normal:reader	f	f	f	normal
2674	dbtask:priority:<nil>:creator	t	f	f	<nil>
2716	dbtask:dbdest_table_id:responsible:manager	t	t	f	responsible
1644	dbtask:dbtask_watcher:normal:admin	t	t	t	normal
1370	dbentity:dbentity_user:normal:reader	f	f	f	normal
1984	confidentiality_level:creator	t	f	f	normal
1328	dbpermission:write:normal:updater	f	t	f	normal
2005	project:code:normal:updater	f	t	f	normal
2026	start_date:creator	t	f	f	normal
2068	protection:protection_area_id:normal:admin	t	t	t	normal
1329	dbpermission:update:normal:manager	t	t	f	normal
1330	dbpermission:update:normal:creator	t	f	f	normal
1331	dbpermission:update:normal:updater	f	t	f	normal
2726	protection_id:reader	f	f	f	normal
1244	dbschema_column:link_sql_order:normal:manager	t	t	f	normal
1245	dbschema_column:link_sql_order:responsible:manager	t	t	f	normal
1246	dbschema_column:link_sql_order:normal:creator	t	f	f	normal
1247	dbschema_column:link_sql_order:responsible:creator	t	f	f	normal
1187	dbschema_column:readonly:normal:manager	t	t	f	normal
1188	dbschema_column:readonly:normal:creator	t	f	f	normal
1986	confidentiality_level:reader	f	f	f	normal
2007	project:code:normal:reader	f	f	f	normal
2028	start_date:reader	f	f	f	normal
1189	dbschema_column:name:normal:admin	t	t	t	normal
1190	dbschema_column:name:normal:manager	t	t	f	normal
1191	dbschema_column:name:normal:creator	t	f	f	normal
1192	dbschema_column:name:normal:updater	f	t	f	normal
1967	result_family:admin	t	t	t	normal
1977	storage_type:updater	f	t	f	normal
1998	project:creator	t	f	f	normal
2019	protection:admin	t	t	t	normal
2297	dbtask:dbschema_id:<nil>:updater	f	t	f	<nil>
2040	protection:end_date:normal:updater	f	t	f	normal
2209	dbrequest:updater	f	t	f	normal
2318	dbtask:dbuser_id:normal:creator	t	f	f	normal
2806	formalized_data:protections:responsible:manager	t	t	f	responsible
2444	dbrequest:current_index:<nil>:creator	t	f	f	<nil>
2528	dbrequest:dbdest_table_id:normal:reader	f	f	f	normal
2549	dbrequest:dbcreated_by_id:responsible:creator	t	f	f	responsible
2612	dbtask:dbcreated_by_id:responsible:reader	f	f	f	responsible
2654	dbtask:is_close:responsible:reader	f	f	f	responsible
2696	dbtask:description:<nil>:updater	f	t	f	<nil>
2227	dbrequest:current_index:normal:updater	f	t	f	normal
1960	support:name:normal:updater	f	t	f	normal
1446	dbview:sql_order:normal:creator	t	f	f	normal
1447	dbview:sql_order:normal:updater	f	t	f	normal
1448	dbview:sql_order:normal:reader	f	f	f	normal
1449	dbview:sql_view:normal:admin	t	t	t	normal
1650	dbworkflow_schema:reader	f	f	f	normal
1596	dbtask:state:normal:manager	t	t	f	normal
1276	dbuser:email:normal:reader	f	f	f	normal
1677	dbtask_assignee:dbuser_id:normal:updater	f	t	f	normal
1678	dbtask_assignee:dbuser_id:normal:reader	f	f	f	normal
1926	formalized_data:storage_area:normal:creator	t	f	f	normal
2267	dbrequest:name:normal:creator	t	f	f	normal
2727	protection_id:admin	t	t	t	normal
2748	formalized_data_project:project_id:normal:creator	t	f	f	normal
2319	dbtask:dbuser_id:<nil>:creator	t	f	f	<nil>
2769	formalized_data_storage_type:storage_type_id:normal:creator	t	f	f	normal
2807	formalized_data:first_evaluation:responsible:admin	t	t	t	responsible
2466	dbrequest:created_date:<nil>:reader	f	f	f	<nil>
2487	dbrequest:state:responsible:admin	t	t	t	responsible
2529	dbrequest:dbdest_table_id:responsible:reader	f	f	f	responsible
2550	dbrequest:dbcreated_by_id:normal:updater	f	t	f	normal
2634	dbtask:created_date:<nil>:manager	t	t	f	<nil>
2655	dbtask:is_close:normal:admin	t	t	t	normal
2676	dbtask:priority:<nil>:updater	f	t	f	<nil>
2718	dbtask:dbdest_table_id:responsible:creator	t	f	f	responsible
1947	support:updater	f	t	f	normal
1530	dbworkflow:admin	t	t	t	normal
2210	dbrequest:reader	f	f	f	normal
2228	dbrequest:current_index:normal:reader	f	f	f	normal
1239	dbschema_column:link_sql_dir:responsible:creator	t	f	f	normal
1359	dbentity:parent_id:normal:updater	f	t	f	normal
1360	dbentity:parent_id:normal:reader	f	f	f	normal
1361	dbentity:parent_id:normal:admin	t	t	t	normal
1362	dbentity:parent_id:normal:manager	t	t	f	normal
1363	dbentity:parent_id:normal:creator	t	f	f	normal
1332	dbpermission:update:normal:reader	f	f	f	normal
1333	dbpermission:update:normal:admin	t	t	t	normal
1334	dbpermission:delete:normal:creator	t	f	f	normal
1335	dbpermission:delete:normal:updater	f	t	f	normal
1336	dbpermission:delete:normal:reader	f	f	f	normal
1754	formalized_data:name:normal:reader	f	f	f	normal
2217	dbrequest:state:normal:updater	f	t	f	normal
2235	dbrequest:created_date:normal:creator	t	f	f	normal
1554	dbtask:dbschema_id:normal:admin	t	t	t	normal
1555	dbtask:dbschema_id:normal:manager	t	t	f	normal
1494	dbentity_user:end_date:normal:admin	t	t	t	normal
1495	dbentity_user:end_date:normal:manager	t	t	f	normal
1496	dbentity_user:end_date:normal:creator	t	f	f	normal
1497	dbentity_user:end_date:normal:updater	f	t	f	normal
1498	dbentity_user:end_date:normal:reader	f	f	f	normal
1388	dbrole:updater	f	t	f	normal
1709	dbtask_verifyer:state:normal:updater	f	t	f	normal
1710	dbtask_verifyer:state:normal:reader	f	f	f	normal
1240	dbschema_column:link_sql_dir:normal:updater	f	t	f	normal
1241	dbschema_column:link_sql_dir:responsible:updater	f	t	f	normal
1364	dbentity:description:normal:manager	t	t	f	normal
1355	dbentity:name:normal:manager	t	t	f	normal
1406	dbrole:dbrole_attribution:normal:creator	t	f	f	normal
1407	dbrole:dbrole_attribution:normal:updater	f	t	f	normal
1437	dbview:readonly:normal:updater	f	t	f	normal
1248	dbschema_column:link_sql_order:normal:updater	f	t	f	normal
1249	dbschema_column:link_sql_order:responsible:updater	f	t	f	normal
1250	dbschema_column:link_sql_order:normal:reader	f	f	f	normal
1251	dbschema_column:link_sql_order:responsible:reader	f	f	f	normal
1549	dbtask:updater	f	t	f	normal
2131	formalized_data:result_family_id:normal:manager	t	t	f	normal
2173	formalized_data:storage_type_id:normal:manager	t	t	f	normal
1904	formalized_data:first_evaluation_date:normal:admin	t	t	t	normal
1303	dbuser:dbentity_user:normal:creator	t	f	f	normal
1304	dbuser:dbrole_attribution:normal:admin	t	t	t	normal
1305	dbuser:dbrole_attribution:normal:manager	t	t	f	normal
1306	dbuser:dbrole_attribution:normal:creator	t	f	f	normal
1307	dbuser:dbrole_attribution:normal:updater	f	t	f	normal
2728	protection_id:manager	t	t	f	normal
2299	dbtask:dbschema_id:<nil>:reader	f	f	f	<nil>
2320	dbtask:dbuser_id:normal:updater	f	t	f	normal
2446	dbrequest:current_index:<nil>:updater	f	t	f	<nil>
2530	dbrequest:dbdest_table_id:normal:admin	t	t	t	normal
2551	dbrequest:dbcreated_by_id:responsible:updater	f	t	f	responsible
2614	dbtask:dbcreated_by_id:responsible:admin	t	t	t	responsible
2656	dbtask:is_close:responsible:admin	t	t	t	responsible
2698	dbtask:description:<nil>:reader	f	f	f	<nil>
1308	dbuser:dbrole_attribution:normal:reader	f	f	f	normal
1310	dbuser:dbhierarchy:normal:creator	t	f	f	normal
1311	dbuser:dbhierarchy:normal:updater	f	t	f	normal
1312	dbuser:dbhierarchy:normal:reader	f	f	f	normal
1313	dbuser:dbhierarchy:normal:admin	t	t	t	normal
1898	formalized_data:first_evaluation:normal:admin	t	t	t	normal
1542	dbworkflow:description:normal:manager	t	t	f	normal
1543	dbworkflow:description:normal:creator	t	f	f	normal
2211	dbrequest:admin	t	t	t	normal
1615	dbtask:description:normal:manager	t	t	f	normal
1616	dbtask:description:normal:creator	t	f	f	normal
1653	dbworkflow_schema:creator	t	f	f	normal
1654	dbworkflow_schema:dbworkflow_id:normal:reader	f	f	f	normal
1655	dbworkflow_schema:dbworkflow_id:normal:admin	t	t	t	normal
1656	dbworkflow_schema:dbworkflow_id:normal:manager	t	t	f	normal
1657	dbworkflow_schema:dbworkflow_id:normal:creator	t	f	f	normal
1658	dbworkflow_schema:dbworkflow_id:normal:updater	f	t	f	normal
1659	dbworkflow_schema:dbschema_id:normal:reader	f	f	f	normal
1660	dbworkflow_schema:dbschema_id:normal:admin	t	t	t	normal
1661	dbworkflow_schema:dbschema_id:normal:manager	t	t	f	normal
1662	dbworkflow_schema:dbschema_id:normal:creator	t	f	f	normal
1663	dbworkflow_schema:dbschema_id:normal:updater	f	t	f	normal
1664	dbworkflow_schema:index:normal:manager	t	t	f	normal
1665	dbworkflow_schema:index:normal:creator	t	f	f	normal
1666	dbworkflow_schema:index:normal:updater	f	t	f	normal
1667	dbworkflow_schema:index:normal:reader	f	f	f	normal
1698	dbtask_verifyer:reader	f	f	f	normal
1699	dbtask_verifyer:dbuser_id:normal:admin	t	t	t	normal
1242	dbschema_column:link_sql_dir:normal:reader	f	f	f	normal
1243	dbschema_column:link_sql_dir:responsible:reader	f	f	f	normal
1686	dbtask_assignee:dbentity_id:normal:creator	t	f	f	normal
1980	result_type:updater	f	t	f	normal
1687	dbtask_assignee:dbentity_id:normal:updater	f	t	f	normal
2212	dbrequest:manager	t	t	f	normal
2230	dbrequest:dbworkflow_id:normal:updater	f	t	f	normal
2001	project:code:normal:manager	t	t	f	normal
1442	dbview:index:normal:updater	f	t	f	normal
1443	dbview:index:normal:reader	f	f	f	normal
1444	dbview:sql_order:normal:admin	t	t	t	normal
1445	dbview:sql_order:normal:manager	t	t	f	normal
1590	dbtask:created_date:normal:reader	f	f	f	normal
1591	dbtask:created_date:normal:admin	t	t	t	normal
1592	dbtask:created_date:normal:manager	t	t	f	normal
1593	dbtask:created_date:normal:creator	t	f	f	normal
1594	dbtask:state:normal:reader	f	f	f	normal
1539	dbworkflow:description:normal:updater	f	t	f	normal
1540	dbworkflow:description:normal:reader	f	f	f	normal
1173	dbschema_column:admin	t	t	t	normal
2729	formalized_data_id:admin	t	t	t	normal
2750	formalized_data_project:project_id:normal:updater	f	t	f	normal
2321	dbtask:dbuser_id:<nil>:updater	f	t	f	<nil>
2771	formalized_data_storage_type:storage_type_id:normal:updater	f	t	f	normal
2468	dbrequest:created_date:<nil>:admin	t	t	t	<nil>
2489	dbrequest:state:responsible:manager	t	t	f	responsible
2531	dbrequest:dbdest_table_id:responsible:admin	t	t	t	responsible
2552	dbrequest:dbcreated_by_id:normal:reader	f	f	f	normal
2636	dbtask:created_date:<nil>:creator	t	f	f	<nil>
2657	dbtask:is_close:normal:manager	t	t	f	normal
2678	dbtask:priority:<nil>:reader	f	f	f	<nil>
2699	dbtask:dbworkflow_schema_id:normal:admin	t	t	t	normal
1541	dbworkflow:description:normal:admin	t	t	t	normal
1426	dbview:indexable:normal:admin	t	t	t	normal
1278	dbuser:email:normal:manager	t	t	f	normal
1427	dbview:indexable:normal:manager	t	t	f	normal
1428	dbview:indexable:normal:creator	t	f	f	normal
1429	dbview:description:normal:admin	t	t	t	normal
1430	dbview:description:normal:manager	t	t	f	normal
1431	dbview:description:normal:creator	t	f	f	normal
1314	dbpermission:creator	t	f	f	normal
1315	dbpermission:updater	f	t	f	normal
1322	dbpermission:name:normal:updater	f	t	f	normal
1323	dbpermission:name:normal:reader	f	f	f	normal
1252	dbschema_column:link_sql_order:normal:admin	t	t	t	normal
1253	dbschema_column:link_sql_order:responsible:admin	t	t	t	normal
1254	dbschema_column:link_sql_view:normal:admin	t	t	t	normal
1874	formalized_data:ref:normal:admin	t	t	t	normal
1574	dbtask:opened_date:normal:creator	t	f	f	normal
1286	dbuser:password:responsible:updater	f	t	f	normal
1287	dbuser:password:normal:reader	f	f	f	normal
1288	dbuser:password:responsible:reader	f	f	f	normal
1289	dbuser:token:normal:admin	t	t	t	normal
1290	dbuser:token:normal:manager	t	t	f	normal
1291	dbuser:token:normal:creator	t	f	f	normal
1292	dbuser:token:normal:updater	f	t	f	normal
1293	dbuser:token:normal:reader	f	f	f	normal
2213	dbrequest:creator	t	f	f	normal
2231	dbrequest:dbworkflow_id:normal:reader	f	f	f	normal
2085	valuation:reader	f	f	f	normal
2127	formalized_data:project_id:normal:reader	f	f	f	normal
2169	formalized_data:storage_type_id:normal:reader	f	f	f	normal
1707	dbtask_verifyer:dbtask_id:normal:reader	f	f	f	normal
1708	dbtask_verifyer:dbtask_id:normal:admin	t	t	t	normal
2089	valuation_type:admin	t	t	t	normal
1383	dbentity:dbhierarchy:normal:reader	f	f	f	normal
2218	dbrequest:state:normal:reader	f	f	f	normal
1922	formalized_data:actualized_evaluation:normal:reader	f	f	f	normal
1350	dbentity:admin	t	t	t	normal
1460	dbview:through_perms:normal:admin	t	t	t	normal
1461	dbview:through_perms:normal:manager	t	t	f	normal
1462	dbview:through_perms:normal:creator	t	f	f	normal
1463	dbview:through_perms:normal:updater	f	t	f	normal
1464	dbview:dbview_id:normal:reader	f	f	f	normal
1465	dbview:dbview_id:normal:admin	t	t	t	normal
1338	dbpermission:delete:normal:manager	t	t	f	normal
1339	dbpermission:read:normal:creator	t	f	f	normal
1340	dbpermission:read:normal:updater	f	t	f	normal
1341	dbpermission:read:normal:reader	f	f	f	normal
1342	dbpermission:read:normal:admin	t	t	t	normal
1343	dbpermission:read:normal:manager	t	t	f	normal
2280	dbrequest:dbworkflow_id:<nil>:admin	t	t	t	<nil>
2301	dbtask:dbschema_id:<nil>:admin	t	t	t	<nil>
2322	dbtask:dbuser_id:normal:reader	f	f	f	normal
2730	formalized_data_id:manager	t	t	f	normal
2448	dbrequest:current_index:<nil>:reader	f	f	f	<nil>
2532	dbrequest:dbdest_table_id:normal:manager	t	t	f	normal
2553	dbrequest:dbcreated_by_id:responsible:reader	f	f	f	responsible
2616	dbtask:dbcreated_by_id:responsible:manager	t	t	f	responsible
2658	dbtask:is_close:responsible:manager	t	t	f	responsible
2700	dbtask:dbworkflow_schema_id:responsible:admin	t	t	t	responsible
1344	dbpermission:dbrole_permission:normal:admin	t	t	t	normal
1345	dbpermission:dbrole_permission:normal:manager	t	t	f	normal
1346	dbpermission:dbrole_permission:normal:creator	t	f	f	normal
1347	dbpermission:dbrole_permission:normal:updater	f	t	f	normal
1564	dbtask:dbopened_by_id:responsible:manager	t	t	f	normal
1565	dbtask:dbopened_by_id:normal:creator	t	f	f	normal
1566	dbtask:dbopened_by_id:responsible:creator	t	f	f	normal
1567	dbtask:dbopened_by_id:normal:updater	f	t	f	normal
1568	dbtask:dbopened_by_id:responsible:updater	f	t	f	normal
1569	dbtask:dbcreated_by_id:normal:updater	f	t	f	normal
2214	dbrequest:state:normal:admin	t	t	t	normal
1351	dbentity:manager	t	t	f	normal
1353	dbentity:updater	f	t	f	normal
2216	dbrequest:state:normal:creator	t	f	f	normal
2234	dbrequest:created_date:normal:manager	t	t	f	normal
2027	start_date:updater	f	t	f	normal
1484	dbentity_user:dbentity_id:normal:admin	t	t	t	normal
1485	dbentity_user:dbentity_id:normal:manager	t	t	f	normal
2215	dbrequest:state:normal:manager	t	t	f	normal
1486	dbentity_user:dbentity_id:normal:creator	t	f	f	normal
1349	dbentity:reader	f	f	f	normal
1487	dbentity_user:dbentity_id:normal:updater	f	t	f	normal
1488	dbentity_user:dbentity_id:normal:reader	f	f	f	normal
1489	dbentity_user:start_date:normal:manager	t	t	f	normal
1490	dbentity_user:start_date:normal:creator	t	f	f	normal
1491	dbentity_user:start_date:normal:updater	f	t	f	normal
1649	dbworkflow_schema:updater	f	t	f	normal
1924	formalized_data:storage_area:normal:manager	t	t	f	normal
1945	support:manager	t	t	f	normal
1232	dbschema_column:link:normal:reader	f	f	f	normal
2731	formalized_data_id:creator	t	f	f	normal
2752	formalized_data_project:project_id:normal:reader	f	f	f	normal
2323	dbtask:dbuser_id:<nil>:reader	f	f	f	<nil>
2773	formalized_data_storage_type:storage_type_id:normal:reader	f	f	f	normal
2491	dbrequest:state:responsible:creator	t	f	f	responsible
2533	dbrequest:dbdest_table_id:responsible:manager	t	t	f	responsible
2638	dbtask:created_date:<nil>:updater	f	t	f	<nil>
2680	dbtask:name:<nil>:admin	t	t	t	<nil>
2701	dbtask:dbworkflow_schema_id:normal:manager	t	t	f	normal
1177	dbschema_column:dbschema_id:normal:manager	t	t	f	normal
1717	dbtask_watcher:updater	f	t	f	normal
1178	dbschema_column:dbschema_id:normal:creator	t	f	f	normal
1878	formalized_data:ref:normal:creator	t	f	f	normal
1988	confidentiality_level:manager	t	t	f	normal
2009	project:dbentity_id:normal:admin	t	t	t	normal
2030	end_date:creator	t	f	f	normal
1950	name:reader	f	f	f	normal
2072	protection:protection_area_id:normal:creator	t	f	f	normal
2093	valuation_type:reader	f	f	f	normal
2135	formalized_data:result_family_id:normal:updater	f	t	f	normal
2177	formalized_data:storage_type_id:normal:updater	f	t	f	normal
2219	dbrequest:is_close:normal:creator	t	f	f	normal
2237	dbrequest:created_date:normal:reader	f	f	f	normal
1511	dbrole_attribution:dbentity_id:normal:creator	t	f	f	normal
1512	dbrole_attribution:dbentity_id:normal:updater	f	t	f	normal
1179	dbschema_column:required:normal:manager	t	t	f	normal
1180	dbschema_column:required:normal:creator	t	f	f	normal
1181	dbschema_column:required:normal:updater	f	t	f	normal
1182	dbschema_column:required:normal:reader	f	f	f	normal
1183	dbschema_column:required:normal:admin	t	t	t	normal
1534	dbworkflow:name:normal:creator	t	f	f	normal
1535	dbworkflow:name:normal:updater	f	t	f	normal
1756	formalized_data:name:normal:admin	t	t	t	normal
2220	dbrequest:is_close:normal:updater	f	t	f	normal
2238	dbrequest:created_date:normal:admin	t	t	t	normal
1399	dbrole:dbrole_permission:normal:admin	t	t	t	normal
1400	dbrole:dbrole_permission:normal:manager	t	t	f	normal
1401	dbrole:dbrole_permission:normal:creator	t	f	f	normal
2282	dbrequest:dbworkflow_id:<nil>:manager	t	t	f	<nil>
2303	dbtask:dbschema_id:<nil>:manager	t	t	f	<nil>
2324	dbtask:dbentity_id:normal:manager	t	t	f	normal
2732	formalized_data_id:updater	f	t	f	normal
2534	dbrequest:dbschema_id:normal:creator	t	f	f	normal
1402	dbrole:dbrole_permission:normal:updater	f	t	f	normal
1928	formalized_data:storage_area:normal:updater	f	t	f	normal
1987	confidentiality_level:admin	t	t	t	normal
2029	end_date:manager	t	t	f	normal
2050	protection:start_date:normal:updater	f	t	f	normal
2092	valuation_type:updater	f	t	f	normal
2113	valuation:valuation_format:normal:creator	t	f	f	normal
2155	formalized_data:support_id:normal:manager	t	t	f	normal
2197	formalized_data:protection_id:normal:reader	f	f	f	normal
1949	name:updater	f	t	f	normal
1550	dbtask:reader	f	f	f	normal
1551	dbtask:admin	t	t	t	normal
1552	dbtask:manager	t	t	f	normal
1553	dbtask:creator	t	f	f	normal
2618	dbtask:dbcreated_by_id:responsible:creator	t	f	f	responsible
2660	dbtask:urgency:<nil>:updater	f	t	f	<nil>
2702	dbtask:dbworkflow_schema_id:responsible:manager	t	t	f	responsible
1470	dbview:dbschema_id:normal:reader	f	f	f	normal
1471	dbview:dbschema_id:normal:admin	t	t	t	normal
1572	dbtask:dbcreated_by_id:normal:manager	t	t	f	normal
1573	dbtask:dbcreated_by_id:normal:creator	t	f	f	normal
2061	protection_area:creator	t	f	f	normal
1728	dbtask_watcher:dbtask_id:normal:updater	f	t	f	normal
1734	dbrole_permission:admin	t	t	t	normal
2207	formalized_data:valuation_id:normal:reader	f	f	f	normal
1896	formalized_data:first_evaluation:normal:reader	f	f	f	normal
1938	formalized_data:contractual:normal:updater	f	t	f	normal
1371	dbentity:dbentity_user:normal:admin	t	t	t	normal
1269	dbuser:name:normal:admin	t	t	t	normal
1270	dbuser:name:normal:manager	t	t	f	normal
1271	dbuser:name:normal:creator	t	f	f	normal
1579	dbtask:opened_date:responsible:reader	f	f	f	normal
1595	dbtask:state:normal:admin	t	t	t	normal
1679	dbtask_assignee:dbtask_id:normal:manager	t	t	f	normal
1195	dbschema_column:type:normal:reader	f	f	f	normal
1747	dbrole_permission:dbpermission_id:normal:manager	t	t	f	normal
1676	dbtask_assignee:dbuser_id:normal:creator	t	f	f	normal
1544	dbworkflow:dbworkflow_schema:normal:admin	t	t	t	normal
1545	dbworkflow:dbworkflow_schema:normal:manager	t	t	f	normal
2229	dbrequest:dbworkflow_id:normal:creator	t	f	f	normal
1962	support:name:normal:reader	f	f	f	normal
1979	result_type:creator	t	f	f	normal
2021	protection:creator	t	f	f	normal
2042	protection:end_date:normal:reader	f	f	f	normal
2063	protection_area:reader	f	f	f	normal
2084	valuation:updater	f	t	f	normal
2105	valuation:valuation_type:normal:updater	f	t	f	normal
2147	formalized_data:result_type_id:normal:reader	f	f	f	normal
2189	formalized_data:protection_id:normal:admin	t	t	t	normal
1317	dbpermission:admin	t	t	t	normal
1318	dbpermission:manager	t	t	f	normal
1219	dbschema_column:description:normal:admin	t	t	t	normal
1220	dbschema_column:description:normal:manager	t	t	f	normal
1221	dbschema_column:description:normal:creator	t	f	f	normal
1222	dbschema_column:description:normal:updater	f	t	f	normal
1223	dbschema_column:description:normal:reader	f	f	f	normal
1224	dbschema_column:link:normal:admin	t	t	t	normal
1968	result_family:manager	t	t	f	normal
1645	dbtask:dbtask_watcher:normal:manager	t	t	f	normal
1646	dbtask:dbtask_watcher:normal:creator	t	f	f	normal
1647	dbtask:dbtask_watcher:normal:updater	f	t	f	normal
1648	dbtask:dbtask_watcher:normal:reader	f	f	f	normal
1652	dbworkflow_schema:manager	t	t	f	normal
1414	dbview:name:normal:creator	t	f	f	normal
1624	dbtask:dbdest_table_id:normal:reader	f	f	f	normal
2733	formalized_data_id:reader	f	f	f	normal
2304	dbtask:dbrequest_id:normal:admin	t	t	t	normal
2325	dbtask:dbentity_id:<nil>:manager	t	t	f	<nil>
1556	dbtask:dbschema_id:normal:creator	t	f	f	normal
1557	dbtask:dbschema_id:normal:updater	f	t	f	normal
1558	dbtask:dbschema_id:normal:reader	f	f	f	normal
2754	formalized_data_storage_type:admin	t	t	t	normal
2775	formalized_data_storage_type:storage_type_id:normal:admin	t	t	t	normal
1559	dbtask:dbopened_by_id:normal:reader	f	f	f	normal
1560	dbtask:dbopened_by_id:responsible:reader	f	f	f	normal
2430	dbrequest:name:<nil>:admin	t	t	t	<nil>
1167	dbschema:label:normal:manager	t	t	f	normal
1168	dbschema:label:normal:creator	t	f	f	normal
2493	dbrequest:state:responsible:updater	f	t	f	responsible
1169	dbschema_column:manager	t	t	f	normal
1170	dbschema_column:creator	t	f	f	normal
1171	dbschema_column:updater	f	t	f	normal
1748	dbrole_permission:dbpermission_id:normal:creator	t	f	f	normal
1750	formalized_data:manager	t	t	f	normal
1751	formalized_data:creator	t	f	f	normal
1752	formalized_data:updater	f	t	f	normal
1642	dbtask:dbtask_verifyer:normal:updater	f	t	f	normal
1758	formalized_data:name:normal:manager	t	t	f	normal
1373	dbentity:dbentity_user:normal:creator	t	f	f	normal
1374	dbentity:dbrole_attribution:normal:admin	t	t	t	normal
1375	dbentity:dbrole_attribution:normal:manager	t	t	f	normal
1376	dbentity:dbrole_attribution:normal:creator	t	f	f	normal
1377	dbentity:dbrole_attribution:normal:updater	f	t	f	normal
1378	dbentity:dbrole_attribution:normal:reader	f	f	f	normal
2535	dbrequest:dbschema_id:responsible:creator	t	f	f	responsible
1379	dbentity:dbhierarchy:normal:admin	t	t	t	normal
1380	dbentity:dbhierarchy:normal:manager	t	t	f	normal
1381	dbentity:dbhierarchy:normal:creator	t	f	f	normal
1382	dbentity:dbhierarchy:normal:updater	f	t	f	normal
1384	dbrole:reader	f	f	f	normal
2640	dbtask:state:<nil>:manager	t	t	f	<nil>
2682	dbtask:name:<nil>:manager	t	t	f	<nil>
2703	dbtask:dbworkflow_schema_id:normal:creator	t	f	f	normal
1625	dbtask:dbdest_table_id:normal:admin	t	t	t	normal
1536	dbworkflow:name:normal:reader	f	f	f	normal
1537	dbworkflow:name:normal:admin	t	t	t	normal
1538	dbworkflow:name:normal:manager	t	t	f	normal
1997	project:manager	t	t	f	normal
1712	dbtask_verifyer:state:normal:manager	t	t	f	normal
1713	dbtask_verifyer:state:normal:creator	t	f	f	normal
1714	dbtask_watcher:admin	t	t	t	normal
1233	dbschema_column:link:responsible:reader	f	f	f	normal
1234	dbschema_column:link_sql_dir:normal:admin	t	t	t	normal
1235	dbschema_column:link_sql_dir:responsible:admin	t	t	t	normal
1365	dbentity:description:normal:creator	t	f	f	normal
2236	dbrequest:created_date:normal:updater	f	t	f	normal
1908	formalized_data:first_evaluation_date:normal:creator	t	f	f	normal
1753	formalized_data:reader	f	f	f	normal
1236	dbschema_column:link_sql_dir:normal:manager	t	t	f	normal
1237	dbschema_column:link_sql_dir:responsible:manager	t	t	f	normal
1238	dbschema_column:link_sql_dir:normal:creator	t	f	f	normal
1366	dbentity:description:normal:updater	f	t	f	normal
1367	dbentity:description:normal:reader	f	f	f	normal
1368	dbentity:description:normal:admin	t	t	t	normal
1983	result_type:manager	t	t	f	normal
1369	dbentity:dbentity_user:normal:updater	f	t	f	normal
2025	start_date:manager	t	t	f	normal
2046	protection:start_date:normal:manager	t	t	f	normal
2088	valuation:creator	t	f	f	normal
2109	valuation:valuation_format:normal:admin	t	t	t	normal
2151	formalized_data:support_id:normal:reader	f	f	f	normal
2193	formalized_data:protection_id:normal:creator	t	f	f	normal
1571	dbtask:dbcreated_by_id:normal:admin	t	t	t	normal
1608	dbtask:priority:normal:manager	t	t	f	normal
1700	dbtask_verifyer:dbuser_id:normal:manager	t	t	f	normal
1159	dbschema:admin	t	t	t	normal
1160	dbschema:manager	t	t	f	normal
1161	dbschema:creator	t	f	f	normal
1162	dbschema:updater	f	t	f	normal
1163	dbschema:reader	f	f	f	normal
1164	dbschema:label:normal:updater	f	t	f	normal
1165	dbschema:label:normal:reader	f	f	f	normal
1166	dbschema:label:normal:admin	t	t	t	normal
1172	dbschema_column:reader	f	f	f	normal
1174	dbschema_column:dbschema_id:normal:updater	f	t	f	normal
1175	dbschema_column:dbschema_id:normal:reader	f	f	f	normal
1412	dbview:updater	f	t	f	normal
1413	dbview:reader	f	f	f	normal
1176	dbschema_column:dbschema_id:normal:admin	t	t	t	normal
1513	dbrole_attribution:dbentity_id:normal:reader	f	f	f	normal
1514	dbrole_attribution:dbrole_id:normal:admin	t	t	t	normal
2284	dbrequest:dbworkflow_id:<nil>:creator	t	f	f	<nil>
2305	dbtask:dbrequest_id:<nil>:admin	t	t	t	<nil>
2326	dbtask:dbentity_id:normal:creator	t	f	f	normal
2734	formalized_data_project:formalized_data_id:normal:reader	f	f	f	normal
2755	formalized_data_storage_type:manager	t	t	f	normal
2536	dbrequest:dbschema_id:normal:updater	f	t	f	normal
2620	dbtask:comment:<nil>:updater	f	t	f	<nil>
2662	dbtask:urgency:<nil>:reader	f	f	f	<nil>
2704	dbtask:dbworkflow_schema_id:responsible:creator	t	f	f	responsible
1515	dbrole_attribution:dbrole_id:normal:manager	t	t	f	normal
1516	dbrole_attribution:dbrole_id:normal:creator	t	f	f	normal
1517	dbrole_attribution:dbrole_id:normal:updater	f	t	f	normal
1518	dbrole_attribution:dbrole_id:normal:reader	f	f	f	normal
1519	dbrole_attribution:start_date:normal:admin	t	t	t	normal
1520	dbrole_attribution:start_date:normal:manager	t	t	f	normal
1521	dbrole_attribution:start_date:normal:creator	t	f	f	normal
1522	dbrole_attribution:start_date:normal:updater	f	t	f	normal
1523	dbrole_attribution:start_date:normal:reader	f	f	f	normal
1524	dbrole_attribution:end_date:normal:creator	t	f	f	normal
1525	dbrole_attribution:end_date:normal:updater	f	t	f	normal
1526	dbrole_attribution:end_date:normal:reader	f	f	f	normal
1527	dbrole_attribution:end_date:normal:admin	t	t	t	normal
1528	dbrole_attribution:end_date:normal:manager	t	t	f	normal
1531	dbworkflow:manager	t	t	f	normal
1532	dbworkflow:creator	t	f	f	normal
1533	dbworkflow:updater	f	t	f	normal
1348	dbpermission:dbrole_permission:normal:reader	f	f	f	normal
1354	dbentity:name:normal:admin	t	t	t	normal
1409	dbview:admin	t	t	t	normal
1603	dbtask:urgency:normal:reader	f	f	f	normal
1604	dbtask:priority:normal:creator	t	f	f	normal
1606	dbtask:priority:normal:reader	f	f	f	normal
1607	dbtask:priority:normal:admin	t	t	t	normal
1930	formalized_data:storage_area:normal:reader	f	f	f	normal
1989	restriction_type:admin	t	t	t	normal
2031	end_date:updater	f	t	f	normal
2052	protection:start_date:normal:reader	f	f	f	normal
2094	valuation_format:admin	t	t	t	normal
2115	valuation:valuation_format:normal:updater	f	t	f	normal
2157	formalized_data:support_id:normal:creator	t	f	f	normal
2199	formalized_data:valuation_id:normal:admin	t	t	t	normal
1951	name:admin	t	t	t	normal
1506	dbrole_attribution:dbuser_id:normal:creator	t	f	f	normal
1507	dbrole_attribution:dbuser_id:normal:updater	f	t	f	normal
1508	dbrole_attribution:dbuser_id:normal:reader	f	f	f	normal
1509	dbrole_attribution:dbentity_id:normal:admin	t	t	t	normal
1510	dbrole_attribution:dbentity_id:normal:manager	t	t	f	normal
1262	dbschema_column:link_sql_view:normal:reader	f	f	f	normal
1605	dbtask:priority:normal:updater	f	t	f	normal
1263	dbschema_column:link_sql_view:responsible:reader	f	f	f	normal
1265	dbuser:admin	t	t	t	normal
1266	dbuser:manager	t	t	f	normal
1267	dbuser:creator	t	f	f	normal
1268	dbuser:updater	f	t	f	normal
1880	formalized_data:ref:normal:updater	f	t	f	normal
2221	dbrequest:is_close:normal:reader	f	f	f	normal
1932	formalized_data:storage_area:normal:admin	t	t	t	normal
1953	name:creator	t	f	f	normal
1529	dbworkflow:reader	f	f	f	normal
1991	restriction_type:creator	t	f	f	normal
2033	end_date:admin	t	t	t	normal
2054	protection_type:admin	t	t	t	normal
1760	formalized_data:name:normal:creator	t	f	f	normal
2096	valuation_format:creator	t	f	f	normal
2117	valuation:valuation_format:normal:reader	f	f	f	normal
2159	formalized_data:confidentiality_level_id:normal:manager	t	t	f	normal
2201	formalized_data:valuation_id:normal:manager	t	t	f	normal
2222	dbrequest:is_close:normal:admin	t	t	t	normal
1298	dbuser:super_admin:normal:reader	f	f	f	normal
1299	dbuser:dbentity_user:normal:updater	f	t	f	normal
1300	dbuser:dbentity_user:normal:reader	f	f	f	normal
1301	dbuser:dbentity_user:normal:admin	t	t	t	normal
1215	dbschema_column:default_value:normal:admin	t	t	t	normal
1324	dbpermission:write:normal:reader	f	f	f	normal
1481	dbentity_user:dbuser_id:normal:admin	t	t	t	normal
1482	dbentity_user:dbuser_id:normal:manager	t	t	f	normal
1483	dbentity_user:dbuser_id:normal:creator	t	f	f	normal
1459	dbview:through_perms:normal:reader	f	f	f	normal
1184	dbschema_column:readonly:normal:updater	f	t	f	normal
1185	dbschema_column:readonly:normal:reader	f	f	f	normal
1186	dbschema_column:readonly:normal:admin	t	t	t	normal
2306	dbtask:dbrequest_id:normal:manager	t	t	f	normal
2327	dbtask:dbentity_id:<nil>:creator	t	f	f	<nil>
2756	formalized_data_storage_type:creator	t	f	f	normal
2777	formalized_data_storage_type:storage_type_id:normal:manager	t	t	f	normal
2432	dbrequest:name:<nil>:manager	t	t	f	<nil>
2495	dbrequest:is_close:responsible:updater	f	t	f	responsible
2537	dbrequest:dbschema_id:responsible:updater	f	t	f	responsible
2642	dbtask:state:<nil>:creator	t	f	f	<nil>
2684	dbtask:name:<nil>:creator	t	f	f	<nil>
2705	dbtask:dbworkflow_schema_id:normal:updater	f	t	f	normal
1910	formalized_data:first_evaluation_date:normal:updater	f	t	f	normal
1990	restriction_type:manager	t	t	f	normal
2011	project:dbentity_id:normal:manager	t	t	f	normal
2032	end_date:reader	f	f	f	normal
2074	protection:protection_type_id:normal:admin	t	t	t	normal
2095	valuation_format:manager	t	t	f	normal
2137	formalized_data:result_family_id:normal:reader	f	f	f	normal
2179	formalized_data:restriction_type_id:normal:admin	t	t	t	normal
1952	name:manager	t	t	f	normal
1391	dbrole:name:normal:creator	t	f	f	normal
1392	dbrole:name:normal:updater	f	t	f	normal
1393	dbrole:name:normal:reader	f	f	f	normal
1394	dbrole:description:normal:admin	t	t	t	normal
1395	dbrole:description:normal:manager	t	t	f	normal
1396	dbrole:description:normal:creator	t	f	f	normal
1397	dbrole:description:normal:updater	f	t	f	normal
1398	dbrole:description:normal:reader	f	f	f	normal
1264	dbuser:reader	f	f	f	normal
1689	dbtask_assignee:state:normal:admin	t	t	t	normal
1436	dbview:readonly:normal:creator	t	f	f	normal
1711	dbtask_verifyer:state:normal:admin	t	t	t	normal
1690	dbtask_assignee:state:normal:manager	t	t	f	normal
1691	dbtask_assignee:state:normal:creator	t	f	f	normal
1692	dbtask_assignee:state:normal:updater	f	t	f	normal
1890	formalized_data:capitalization_date:normal:updater	f	t	f	normal
1693	dbtask_assignee:state:normal:reader	f	f	f	normal
1694	dbtask_verifyer:admin	t	t	t	normal
1695	dbtask_verifyer:manager	t	t	f	normal
1696	dbtask_verifyer:creator	t	f	f	normal
1701	dbtask_verifyer:dbuser_id:normal:creator	t	f	f	normal
1702	dbtask_verifyer:dbuser_id:normal:updater	f	t	f	normal
1703	dbtask_verifyer:dbuser_id:normal:reader	f	f	f	normal
1704	dbtask_verifyer:dbtask_id:normal:manager	t	t	f	normal
1705	dbtask_verifyer:dbtask_id:normal:creator	t	f	f	normal
1492	dbentity_user:start_date:normal:reader	f	f	f	normal
1493	dbentity_user:start_date:normal:admin	t	t	t	normal
1697	dbtask_verifyer:updater	f	t	f	normal
1440	dbview:index:normal:manager	t	t	f	normal
1441	dbview:index:normal:creator	t	f	f	normal
1900	formalized_data:first_evaluation:normal:manager	t	t	f	normal
1942	formalized_data:contractual:normal:admin	t	t	t	normal
1294	dbuser:super_admin:normal:admin	t	t	t	normal
1474	dbentity_user:admin	t	t	t	normal
1275	dbuser:email:normal:updater	f	t	f	normal
1279	dbuser:password:normal:admin	t	t	t	normal
1295	dbuser:super_admin:normal:manager	t	t	f	normal
1296	dbuser:super_admin:normal:creator	t	f	f	normal
1297	dbuser:super_admin:normal:updater	f	t	f	normal
2286	dbrequest:dbworkflow_id:<nil>:updater	f	t	f	<nil>
2307	dbtask:dbrequest_id:<nil>:manager	t	t	f	<nil>
2328	dbtask:dbentity_id:normal:updater	f	t	f	normal
2736	formalized_data_project:formalized_data_id:normal:admin	t	t	t	normal
2757	formalized_data_storage_type:updater	f	t	f	normal
2538	dbrequest:dbschema_id:normal:reader	f	f	f	normal
2622	dbtask:comment:<nil>:reader	f	f	f	<nil>
2664	dbtask:urgency:<nil>:admin	t	t	t	<nil>
2706	dbtask:dbworkflow_schema_id:responsible:updater	f	t	f	responsible
1403	dbrole:dbrole_permission:normal:reader	f	f	f	normal
1404	dbrole:dbrole_attribution:normal:admin	t	t	t	normal
1405	dbrole:dbrole_attribution:normal:manager	t	t	f	normal
1255	dbschema_column:link_sql_view:responsible:admin	t	t	t	normal
1256	dbschema_column:link_sql_view:normal:manager	t	t	f	normal
1257	dbschema_column:link_sql_view:responsible:manager	t	t	f	normal
1258	dbschema_column:link_sql_view:normal:creator	t	f	f	normal
1259	dbschema_column:link_sql_view:responsible:creator	t	f	f	normal
1260	dbschema_column:link_sql_view:normal:updater	f	t	f	normal
1261	dbschema_column:link_sql_view:responsible:updater	f	t	f	normal
2308	dbtask:dbrequest_id:normal:creator	t	f	f	normal
2329	dbtask:dbentity_id:<nil>:updater	f	t	f	<nil>
2758	formalized_data_storage_type:reader	f	f	f	normal
2779	dbschema_column:link_is_enum:normal:admin	t	t	t	normal
2434	dbrequest:name:<nil>:creator	t	f	f	<nil>
2497	dbrequest:is_close:responsible:reader	f	f	f	responsible
2539	dbrequest:dbschema_id:responsible:reader	f	f	f	responsible
2644	dbtask:state:<nil>:updater	f	t	f	<nil>
2686	dbtask:name:<nil>:updater	f	t	f	<nil>
2707	dbtask:dbworkflow_schema_id:normal:reader	f	f	f	normal
1882	formalized_data:ref:normal:reader	f	f	f	normal
1912	formalized_data:first_evaluation_date:normal:reader	f	f	f	normal
1992	restriction_type:updater	f	t	f	normal
2013	project:dbentity_id:normal:creator	t	f	f	normal
2034	protection:end_date:normal:admin	t	t	t	normal
1954	support:name:normal:admin	t	t	t	normal
2055	protection_type:manager	t	t	f	normal
2076	protection:protection_type_id:normal:manager	t	t	f	normal
2097	valuation_format:updater	f	t	f	normal
2139	formalized_data:result_type_id:normal:admin	t	t	t	normal
2181	formalized_data:restriction_type_id:normal:manager	t	t	f	normal
2223	dbrequest:is_close:normal:manager	t	t	f	normal
2288	dbrequest:dbworkflow_id:<nil>:reader	f	f	f	<nil>
2309	dbtask:dbrequest_id:<nil>:creator	t	f	f	<nil>
2330	dbtask:dbentity_id:normal:reader	f	f	f	normal
2738	formalized_data_project:formalized_data_id:normal:manager	t	t	f	normal
2759	formalized_data_storage_type:formalized_data_id:normal:creator	t	f	f	normal
2540	dbrequest:dbschema_id:normal:admin	t	t	t	normal
2624	dbtask:comment:<nil>:admin	t	t	t	<nil>
1934	formalized_data:contractual:normal:manager	t	t	f	normal
1762	formalized_data:name:normal:updater	f	t	f	normal
1993	restriction_type:reader	f	f	f	normal
2056	protection_type:creator	t	f	f	normal
2098	valuation_format:reader	f	f	f	normal
2119	formalized_data:project_id:normal:admin	t	t	t	normal
2161	formalized_data:confidentiality_level_id:normal:creator	t	f	f	normal
2203	formalized_data:valuation_id:normal:creator	t	f	f	normal
2224	dbrequest:current_index:normal:admin	t	t	t	normal
2666	dbtask:urgency:<nil>:manager	t	t	f	<nil>
2708	dbtask:dbworkflow_schema_id:responsible:reader	f	f	f	responsible
2310	dbtask:dbrequest_id:normal:updater	f	t	f	normal
2331	dbtask:dbentity_id:<nil>:reader	f	f	f	<nil>
2781	dbschema_column:link_is_enum:normal:manager	t	t	f	normal
2436	dbrequest:name:<nil>:updater	f	t	f	<nil>
2499	dbrequest:is_close:responsible:admin	t	t	t	responsible
2541	dbrequest:dbschema_id:responsible:admin	t	t	t	responsible
2646	dbtask:state:<nil>:reader	f	f	f	<nil>
2688	dbtask:name:<nil>:reader	f	f	f	<nil>
2191	formalized_data:protection_id:normal:manager	t	t	f	normal
1588	dbtask:comment:normal:reader	f	f	f	normal
1589	dbtask:created_date:normal:updater	f	t	f	normal
1706	dbtask_verifyer:dbtask_id:normal:updater	f	t	f	normal
1982	result_type:admin	t	t	t	normal
1570	dbtask:dbcreated_by_id:normal:reader	f	f	f	normal
1617	dbtask:description:normal:updater	f	t	f	normal
1618	dbtask:description:normal:reader	f	f	f	normal
1619	dbtask:header:normal:admin	t	t	t	normal
1892	formalized_data:capitalization_date:normal:reader	f	f	f	normal
1620	dbtask:header:normal:manager	t	t	f	normal
1920	formalized_data:actualized_evaluation:normal:updater	f	t	f	normal
1688	dbtask_assignee:dbentity_id:normal:reader	f	f	f	normal
\.


--
-- Data for Name: dbrequest; Type: TABLE DATA; Schema: public; Owner: test
--

COPY public.dbrequest (id, dbworkflow_id, name, current_index, created_date, state, is_close, dbdest_table_id, dbschema_id, dbcreated_by_id) FROM stdin;
51	1	Formalized Datas Entry Request	2	2024-02-20 00:00:00	completed	f	190	309	2
41	1	Formalized Datas Entry Request	0	2024-02-20 14:13:26.880174	pending	f	169	309	2
42	1	Formalized Datas Entry Request	0	2024-02-20 14:13:26.880174	pending	f	171	309	2
43	1	Formalized Datas Entry Request	2	2024-02-20 00:00:00	completed	f	173	309	2
45	1	Formalized Datas Entry Request	2	2024-02-20 00:00:00	completed	f	179	309	2
47	1	Formalized Datas Entry Request	2	2024-02-20 00:00:00	completed	f	183	309	2
49	1	Formalized Datas Entry Request	1	2024-02-20 00:00:00	progressing	f	188	309	2
\.


--
-- Data for Name: dbrole; Type: TABLE DATA; Schema: public; Owner: test
--

COPY public.dbrole (id, name, description) FROM stdin;
1	researcher	no description...
2	CDP attribution	no description...
3	CDG attribution	no description...
4	RH	no description...
5	Juridic	no description...
\.


--
-- Data for Name: dbrole_attribution; Type: TABLE DATA; Schema: public; Owner: test
--

COPY public.dbrole_attribution (id, dbuser_id, dbentity_id, dbrole_id, start_date, end_date) FROM stdin;
1	\N	1	1	2024-02-15 08:53:02.660215	\N
2	\N	2	2	2024-02-15 08:53:02.660215	\N
4	\N	3	3	2024-02-15 08:53:02.660215	\N
5	\N	4	4	2024-02-21 11:12:15.664015	\N
6	\N	5	5	2024-02-28 09:50:15.831208	\N
\.


--
-- Data for Name: dbrole_permission; Type: TABLE DATA; Schema: public; Owner: test
--

COPY public.dbrole_permission (id, dbrole_id, dbpermission_id) FROM stdin;
2	2	1749
4	3	1753
5	1	1995
6	2	1996
7	3	1995
8	1	1966
9	2	1966
10	3	1966
11	1	1981
12	2	1981
13	3	1981
14	1	1944
15	2	1944
16	3	1948
17	1	1986
18	2	1986
19	3	1986
20	1	1993
21	2	1993
22	3	1993
26	1	2085
27	2	2085
28	3	2085
29	1	1978
30	2	1978
31	3	1978
1	1	1751
32	4	1265
33	1	2723
34	2	2723
36	1	2171
37	2	2171
39	1	2754
40	2	2754
43	5	2806
44	5	2802
45	5	2020
46	5	2055
47	5	2060
48	5	1990
49	3	2803
50	3	2087
51	3	2090
52	3	2095
53	2	2805
54	3	2799
55	3	2800
56	3	2801
57	2	2805
58	5	1966
59	5	1981
60	5	1948
42	5	1753
61	5	1995
62	5	2722
63	5	2019
64	3	2094
65	3	2807
66	3	2801
67	3	1914
68	3	2800
69	1	2723
70	2	2723
71	3	2723
72	4	2723
73	5	2723
\.


--
-- Data for Name: dbschema; Type: TABLE DATA; Schema: public; Owner: test
--

COPY public.dbschema (id, name, label) FROM stdin;
5820	dbtask	activity
5028	dbhierarchy	hierarchy
8050	formalized_data_storage_type	storage types related to formalized datas
309	formalized_data	formalized datas
7747	formalized_data_project	formalized datas related to projects
5804	dbrequest	request
779	storage_type	general
780	result_type	general
781	confidentiality_level	general
782	restriction_type	general
783	project	general
784	protection	general
785	start_date	general
786	end_date	general
787	protection_type	general
788	protection_area	general
789	valuation	general
790	valuation_type	general
791	valuation_format	general
742	support	general
743	name	general
133	dbschema	general
134	dbschema_column	general
135	dbuser	general
136	dbpermission	general
137	dbentity	general
138	dbrole	general
139	dbview	general
140	dbentity_user	general
141	dbrole_attribution	general
142	dbworkflow	general
144	dbworkflow_schema	general
145	dbtask_assignee	general
146	dbtask_verifyer	general
147	dbtask_watcher	general
148	dbrole_permission	general
778	result_family	general
\.


--
-- Data for Name: dbschema_column; Type: TABLE DATA; Schema: public; Owner: test
--

COPY public.dbschema_column (id, dbschema_id, required, read_level, readonly, name, type, index, label, placeholder, default_value, description, link, link_sql_dir, link_sql_order, link_sql_view) FROM stdin;
877	783	t	normal	t	code	varchar(255)	1	code	\N	\N	no description...	\N	\N	\N	\N
964	5820	\N	\N	t	dbentity_id	integer	1	entity assignee	\N	\N	no description...	dbentity	\N	\N	\N
971	5820	\N	\N	t	priority	enum:low,medium,high	1	priority	\N	medium	no description...	\N	\N	\N	\N
849	146	\N	\N	f	dbuser_id	integer	1	dbuser id	\N	\N	no description...	dbuser	\N	\N	\N
850	146	\N	\N	f	dbtask_id	integer	1	dbtask id	\N	\N	no description...	dbtask	\N	\N	\N
851	146	\N	\N	f	state	enum:pending,rejected,complete	1	state	\N	pending	no description...	\N	\N	\N	\N
852	147	\N	\N	f	dbuser_id	integer	1	dbuser id	\N	\N	no description...	dbuser	\N	\N	\N
820	141	\N	\N	f	start_date	timestamp	1	start date	\N	2024-02-14 14:30:04.904936+00	no description...	\N	\N	\N	\N
785	136	\N	\N	t	name	varchar(255)	1	name	\N	\N	no description...	\N	\N	\N	\N
901	139	\N	normal	f	is_empty	boolean	1	empty		false	no description...	\N	\N	\N	\N
882	784	f	normal	f	protection_type_id	integer	1	protection type	\N	\N	no description...	protection_type	\N	\N	\N
887	309	f	normal	f	result_type_id	integer	8	result type	\N	\N	no description...	result_type	\N	\N	\N
957	5804	\N	responsible	t	dbdest_table_id	integer	1	reference	\N	\N	no description...	\N	\N	\N	\N
979	7747	f	normal	f	formalized_data_id	integer	1	formalized_data	\N	\N	formalized data ref of the relation between formalized data and project	formalized_data	\N	\N	\N
988	784	t	normal	f	formalized_data_id	integer	1	formalized datas	\N	\N	no description...	formalized_data	\N	\N	\N
965	5820	\N	responsible	t	dbcreated_by_id	integer	1	dbcreated by id	\N	\N	no description...	dbuser	\N	\N	\N
972	5820	\N	\N	t	name	varchar(255)	1	name	\N	\N	no description...	\N	\N	\N	\N
980	7747	t	normal	f	project_id	integer	1	project	\N	\N	project ref of the relation between formalized data and project	project	\N	\N	\N
875	309	f	normal	f	contractual	boolean	16	contractual	\N	\N	no description...	\N	\N	\N	\N
886	309	f	normal	f	result_family_id	integer	8	result family	\N	\N	no description...	result_family	\N	\N	\N
958	5804	\N	responsible	t	dbschema_id	integer	1	form attached	\N	\N	no description...	dbschema	\N	\N	\N
890	309	f	normal	f	storage_types	manytomany	12	storage types	\N	\N	no description...	formalized_data_storage_type	\N	\N	\N
873	309	f	responsible	f	actualized_evaluation	money	6	actualized evaluation	\N	\N	no description...	\N	\N	\N	\N
870	309	f	responsible	f	capitalization_date	timestamp	3	capitalization date	\N	CURRENT_TIME	no description...	\N	\N	\N	\N
823	142	\N	\N	f	description	varchar(255)	1	description	\N	\N	no description...	\N	\N	\N	\N
824	142	\N	\N	f	dbworkflow_schema	manytomany	1	dbworkflow schema	\N	\N	no description...	\N	\N	\N	\N
786	136	\N	\N	f	write	boolean	1	write	\N	\N	no description...	\N	\N	\N	\N
787	136	\N	\N	f	update	boolean	1	update	\N	\N	no description...	\N	\N	\N	\N
788	136	\N	\N	f	delete	boolean	1	delete	\N	\N	no description...	\N	\N	\N	\N
789	136	\N	\N	f	read	enum:normal,responsible,moderator,admin	1	read	\N	normal	no description...	\N	\N	\N	\N
790	136	\N	\N	f	dbrole_permission	manytomany	1	dbrole permission	\N	\N	no description...	\N	\N	\N	\N
853	147	\N	\N	f	dbtask_id	integer	1	dbtask id	\N	\N	no description...	dbtask	\N	\N	\N
843	144	\N	\N	t	dbschema_id	integer	1	dbschema id	\N	\N	no description...	dbschema	\N	\N	\N
805	139	\N	\N	f	readonly	boolean	1	readonly	\N	\N	no description...	\N	\N	\N	\N
881	784	f	normal	f	protection_area_id	integer	1	protection area	\N	\N	no description...	protection_area	\N	\N	\N
989	782	t	normal	f	formalized_data_id		1	formalized datas	\N	\N	no description...	formalized_data	\N	\N	\N
845	145	\N	\N	t	dbuser_id	integer	1	dbuser id	\N	\N	no description...	dbuser	\N	\N	\N
846	145	\N	\N	t	dbtask_id	integer	1	dbtask id	\N	\N	no description...	dbtask	\N	\N	\N
807	139	\N	\N	f	sql_order	varchar(255)	1	sql order	\N	\N	no description...	\N	\N	\N	\N
868	309	t	normal	t	name	varchar(255)	1	name	\N	\N	no description...	\N	\N	\N	\N
812	139	\N	\N	f	dbschema_id	integer	1	dbschema id	\N	\N	no description...	dbschema	\N	\N	\N
814	140	\N	\N	t	dbentity_id	integer	1	dbentity id	\N	\N	no description...	dbentity	\N	\N	\N
797	138	\N	\N	t	name	varchar(255)	1	name	\N	\N	no description...	\N	\N	\N	\N
791	137	\N	\N	t	name	varchar(255)	1	name	\N	\N	no description...	\N	\N	\N	\N
792	137	\N	\N	f	parent_id	varchar(255)	1	parent id	\N	\N	no description...	\N	\N	\N	\N
793	137	\N	\N	f	description	text	1	description	\N	\N	no description...	\N	\N	\N	\N
794	137	\N	\N	f	dbentity_user	manytomany	1	dbentity user	\N	\N	no description...	\N	\N	\N	\N
795	137	\N	\N	f	dbrole_attribution	manytomany	1	dbrole attribution	\N	\N	no description...	\N	\N	\N	\N
796	137	\N	\N	f	dbhierarchy	manytomany	1	dbhierarchy	\N	\N	no description...	\N	\N	\N	\N
771	134	\N	\N	f	default_value	varchar(255)	1	default value	\N	\N	no description...	\N	\N	\N	\N
772	134	\N	\N	f	description	varchar(255)	1	description	\N	no description...	no description...	\N	\N	\N	\N
773	134	\N	responsible	f	link	varchar(255)	1	link	\N	\N	no description...	\N	\N	\N	\N
774	134	\N	responsible	f	link_sql_dir	varchar(255)	1	link sql dir	\N	\N	no description...	\N	\N	\N	\N
775	134	\N	responsible	f	link_sql_order	varchar(255)	1	link sql order	\N	\N	no description...	\N	\N	\N	\N
778	135	\N	\N	t	email	varchar(255)	1	email	\N	\N	no description...	\N	\N	\N	\N
779	135	\N	responsible	f	password	varchar(255)	1	password	\N	\N	no description...	\N	\N	\N	\N
780	135	\N	\N	f	token	varchar(255)	1	token	\N	\N	no description...	\N	\N	\N	\N
798	138	\N	\N	f	description	text	1	description	\N	no description...	no description...	\N	\N	\N	\N
799	138	\N	\N	f	dbrole_permission	manytomany	1	dbrole permission	\N	\N	no description...	\N	\N	\N	\N
800	138	\N	\N	f	dbrole_attribution	manytomany	1	dbrole attribution	\N	\N	no description...	\N	\N	\N	\N
801	139	\N	\N	f	name	varchar(255)	1	name	\N	\N	no description...	\N	\N	\N	\N
802	139	\N	\N	f	is_list	boolean	1	is list	\N	true	no description...	\N	\N	\N	\N
803	139	\N	\N	f	indexable	boolean	1	indexable	\N	true	no description...	\N	\N	\N	\N
966	5820	\N	\N	t	comment	text	1	comment	\N	\N	no description...	\N	\N	\N	\N
813	140	\N	\N	t	dbuser_id	integer	1	dbuser id	\N	\N	no description...	dbuser	\N	\N	\N
808	139	\N	\N	f	sql_view	varchar(255)	1	sql view	\N	\N	no description...	\N	\N	\N	\N
809	139	\N	\N	f	sql_dir	varchar(255)	1	sql dir	\N	\N	no description...	\N	\N	\N	\N
762	133	\N	\N	t	label	varchar(255)	1	label	\N	unknown label	no description...	\N	\N	\N	\N
766	134	\N	\N	t	name	varchar(255)	1	name	\N	\N	no description...	\N	\N	\N	\N
763	134	\N	\N	t	dbschema_id	integer	1	dbschema id	\N	\N	no description...	dbschema	\N	\N	\N
764	134	\N	\N	f	required	boolean	1	required	\N	false	no description...	\N	\N	\N	\N
765	134	\N	\N	f	readonly	boolean	1	readonly	\N	\N	no description...	\N	\N	\N	\N
973	5820	\N	\N	t	description	varchar(255)	1	description	\N	no description...	no description...	\N	\N	\N	\N
981	779	\N	normal	f	name	varchar(255)	1	name	\N	\N	no description...	\N	\N	\N	\N
876	742	t	normal	t	name	varchar(255)	1	name	\N	\N	no description...	\N	\N	\N	\N
952	5804	\N	responsible	t	name	varchar(255)	1	name	\N	\N	no description...	\N	\N	\N	\N
872	309	f	responsible	f	first_evaluation_date	timestamp	5	first evaluation date	\N	CURRENT_TIME	no description...	\N	\N	\N	\N
806	139	\N	\N	f	index	integer	1	index	\N	1	no description...	\N	\N	\N	\N
891	309	f	responsible	f	restriction_types	onetomany	13	restriction types	\N	\N	no description...	restriction_type	\N	\N	\N
990	790	t	normal	f	name	varchar(255)	1	name	\N	\N	no description...	\N	\N	\N	\N
871	309	f	responsible	f	first_evaluation	money	4	first evaluation	\N	\N	no description...	\N	\N	\N	\N
959	5804	\N	responsible	t	dbcreated_by_id	integer	1	dbcreated by id	\N	\N	no description...	dbuser	\N	\N	\N
967	5820	\N	\N	t	created_date	timestamp	1	created date	\N	2024-02-20 14:15:08.176263+00	no description...	\N	\N	\N	\N
960	5804	\N	responsible	t	created_date	timestamp	1	created date	\N	2024-02-20 14:13:26.880174+00	no description...	\N	\N	\N	\N
804	139	\N	\N	f	description	varchar(255)	1	description	\N	no description...	no description...	\N	\N	\N	\N
770	134	\N	\N	f	placeholder	varchar(255)	1	placeholder	\N	\N	no description...	\N	\N	\N	\N
847	145	\N	\N	t	dbentity_id	integer	1	dbentity id	\N	\N	no description...	dbentity	\N	\N	\N
974	5820	\N	normal	t	dbworkflow_schema_id	integer	1	workflow attached	\N	\N	no description...	dbworkflow_schema	\N	\N	\N
953	5804	\N	normal	f	state	enum:pending,progressing,dismiss,completed	1	state	\N	pending	no description...	\N	\N	\N	\N
982	8050	t	normal	f	formalized_data_id	integer	0	formalized data ref	\N	\N	formalized data ref related to storage types	formalized_data	\N	\N	\N
904	142	\N	normal	f	dbschema_id	integer	1	form	\N	\N	no description...	dbschema	\N	\N	\N
893	309	f	responsible	f	valuations	onetomany	15	valuations	\N	\N	no description...	valuation	\N	\N	\N
883	789	f	normal	f	valuation_type_id	integer	1	valuation type	\N	\N	no description...	valuation_type	\N	\N	\N
991	791	t	normal	f	name	varchar(255)	1	name	\N	\N	no description...	\N	\N	\N	\N
878	783	t	normal	t	dbentity_id	integer	1	entity related	\N	\N	no description...	dbentity	\N	\N	\N
888	309	f	normal	f	support_id	integer	9	storage support	\N	\N	no description...	support	\N	\N	\N
848	145	\N	\N	f	state	enum:dismiss,progressing,pending,completed	1	state	\N	pending	no description...	\N	\N	\N	\N
968	5820	\N	\N	f	state	enum:pending,progressing,dismiss,completed	1	state	\N	pending	no description...	\N	\N	\N	\N
975	5820	\N	normal	t	dbdest_table_id	integer	1	reference	\N	\N	no description...	\N	\N	\N	\N
983	8050	t	normal	f	storage_type_id	integer	0	storage type ref	\N	\N	storage type ref related to formalized datas	storage_type	\N	\N	\N
889	309	f	responsible	f	confidentiality_level_id	integer	10	confidentiality level	\N	\N	no description...	confidentiality_level	\N	\N	\N
842	144	\N	\N	t	dbworkflow_id	integer	1	dbworkflow id	\N	\N	no description...	dbworkflow	\N	\N	\N
844	144	\N	\N	f	index	integer	1	index	\N	1	no description...	\N	\N	\N	\N
884	789	f	normal	f	valuation_format_id	integer	1	valuation format	\N	\N	no description...	valuation_format	\N	\N	\N
992	789	t	normal	f	formalized_data_id	integer	1	formalized datas	\N	\N	no description...	formalized_data	\N	\N	\N
811	139	\N	\N	f	dbview_id	integer	1	dbview id	\N	\N	no description...	dbview	\N	\N	\N
879	784	f	normal	f	end_date	timestamp	1	end date	\N	\N	no description...	\N	\N	\N	\N
767	134	\N	\N	f	type	varchar(255)	1	type	\N	\N	no description...	\N	\N	\N	\N
768	134	\N	\N	f	index	integer	1	index	\N	1	no description...	\N	\N	\N	\N
769	134	\N	\N	f	label	varchar(255)	1	label	\N	\N	no description...	\N	\N	\N	\N
905	783	t	normal	f	name	varchar(255)	1	name	\N	\N	no description...	\N	\N	\N	\N
954	5804	\N	responsible	f	is_close	boolean	1	is close	\N	false	no description...	\N	\N	\N	\N
961	5820	\N	\N	t	dbschema_id	integer	1	form attached	\N	\N	no description...	dbschema	\N	\N	\N
969	5820	\N	responsible	f	is_close	boolean	1	is close	\N	false	no description...	\N	\N	\N	\N
976	139	\N	normal	f	sql_restriction	varchar(255)	1	restriction	\N	\N	no description...	\N	\N	\N	\N
815	140	\N	\N	f	start_date	timestamp	1	start date	\N	2024-02-14 14:30:01.232486+00	no description...	\N	\N	\N	\N
816	140	\N	\N	f	end_date	timestamp	1	end date	\N	\N	no description...	\N	\N	\N	\N
821	141	\N	\N	f	end_date	timestamp	1	end date	\N	\N	no description...	\N	\N	\N	\N
880	784	f	normal	f	start_date	timestamp	1	start date	\N	\N	no description...	\N	\N	\N	\N
955	5804	\N	normal	t	current_index	integer	1	current task	\N	0	no description...	\N	\N	\N	\N
817	141	\N	\N	t	dbuser_id	integer	1	dbuser id	\N	\N	no description...	dbuser	\N	\N	\N
818	141	\N	\N	t	dbentity_id	integer	1	dbentity id	\N	\N	no description...	dbentity	\N	\N	\N
819	141	\N	\N	t	dbrole_id	integer	1	dbrole id	\N	\N	no description...	dbrole	\N	\N	\N
822	142	\N	\N	t	name	varchar(255)	1	name	\N	\N	no description...	\N	\N	\N	\N
782	135	\N	\N	f	dbentity_user	manytomany	1	dbentity user	\N	\N	no description...	\N	\N	\N	\N
784	135	\N	\N	f	dbhierarchy	manytomany	1	dbhierarchy	\N	\N	no description...	\N	\N	\N	\N
810	139	\N	\N	f	through_perms	integer	1	through perms	\N	\N	no description...	dbschema	\N	\N	\N
984	134	f	normal	t	link_is_enum	boolean	0	link is enum	\N	\N	define if link is enum	\N	\N	\N	\N
874	309	f	normal	f	storage_area	varchar(255)	11	storage area	\N	\N	no description...	\N	\N	\N	\N
977	309	f	normal	f	projects	manytomany	1	projects	\N	\N	no description...	formalized_data_project	\N	\N	\N
986	782	t	normal	f	name	varchar(255)	1	name	\N	\N	no description...	\N	\N	\N	\N
993	789	t	normal	f	name	varchar(255)	1	name	\N	\N	no description...	\N	\N	\N	\N
962	5820	\N	\N	t	dbrequest_id	integer	1	request attached	\N	\N	no description...	dbrequest	\N	\N	\N
970	5820	\N	\N	t	urgency	enum:low,medium,high	1	urgency	\N	medium	no description...	\N	\N	\N	\N
963	5820	\N	\N	t	dbuser_id	integer	1	user assignee	\N	\N	no description...	dbuser	\N	\N	\N
776	134	\N	responsible	f	link_sql_view	varchar(255)	1	link sql view	\N	\N	no description...	\N	\N	\N	\N
777	135	\N	\N	t	name	varchar(255)	1	name	\N	\N	no description...	\N	\N	\N	\N
978	309	\N	responsible	f	protections	onetomany	1	protections	\N	\N	no description...	protection	\N	\N	\N
781	135	\N	\N	f	super_admin	boolean	1	super admin	\N	\N	no description...	\N	\N	\N	\N
854	147	\N	\N	f	dbentity_id	integer	1	dbentity id	\N	\N	no description...	dbentity	\N	\N	\N
855	148	\N	\N	t	dbrole_id	integer	1	dbrole id	\N	\N	no description...	dbrole	\N	\N	\N
856	148	\N	\N	t	dbpermission_id	integer	1	dbpermission id	\N	\N	no description...	dbpermission	\N	\N	\N
987	784	\N	normal	f	name	varchar(255)	1	name	\N	\N	no description...	\N	\N	\N	\N
869	309	t	normal	t	ref	varchar(255)	2	referencing	\N	\N	no description...	\N	\N	\N	\N
956	5804	\N	\N	f	dbworkflow_id	integer	1	workflow attached	\N	\N	no description...	dbworkflow	\N	\N	\N
\.


--
-- Data for Name: dbtask; Type: TABLE DATA; Schema: public; Owner: test
--

COPY public.dbtask (id, dbschema_id, dbrequest_id, dbuser_id, dbentity_id, dbcreated_by_id, comment, created_date, state, is_close, urgency, priority, name, description, dbworkflow_schema_id, dbdest_table_id) FROM stdin;
116	309	49	3	\N	2	\N	2024-02-20 14:15:08.176263	completed	t	medium	medium	hierarchical verification	hierarchical verification expected by the system, workflow is currently pending.	\N	188
118	309	49	\N	\N	3	\N	2024-02-20 14:15:08.176263	pending	f	medium	medium	Add Juridic protection	Add Juridic protection on formalized datas.	1	188
117	309	51	3	\N	2	\N	2024-02-20 14:15:08.176263	completed	t	medium	medium	hierarchical verification	hierarchical verification expected by the system, workflow is currently pending.	\N	190
108	309	43	\N	\N	3	\N	2024-02-20 14:15:08.176263	completed	t	medium	medium	Add Juridic protection	Add Juridic protection on formalized datas.	1	173
119	309	51	\N	\N	3	\N	2024-02-20 14:15:08.176263	completed	t	medium	medium	Add Juridic protection	Add Juridic protection on formalized datas.	1	190
120	309	51	\N	\N	6	\N	2024-02-20 14:15:08.176263	pending	f	medium	medium	Add valuation	Add valuation on formalized datas.	2	190
109	309	43	\N	\N	6	\N	2024-02-20 14:15:08.176263	completed	t	medium	medium	Add valuation	Add valuation on formalized datas.	2	173
110	309	45	3	\N	2	\N	2024-02-20 14:15:08.176263	completed	t	medium	medium	hierarchical verification	hierarchical verification expected by the system, workflow is currently pending.	\N	179
111	309	45	\N	\N	3	\N	2024-02-20 14:15:08.176263	completed	t	medium	medium	Add Juridic protection	Add Juridic protection on formalized datas.	1	179
112	309	45	\N	\N	6	\N	2024-02-20 14:15:08.176263	completed	t	medium	medium	Add valuation	Add valuation on formalized datas.	2	179
113	309	47	3	\N	2	\N	2024-02-20 14:15:08.176263	completed	t	medium	medium	hierarchical verification	hierarchical verification expected by the system, workflow is currently pending.	\N	183
114	309	47	\N	\N	3	\N	2024-02-20 14:15:08.176263	completed	t	medium	medium	Add Juridic protection	Add Juridic protection on formalized datas.	1	183
105	309	43	3	\N	2	\N	2024-02-20 14:15:08.176263	completed	t	medium	medium	hierarchical verification	hierarchical verification expected by the system, workflow is currently pending.	\N	173
115	309	47	\N	\N	6	\N	2024-02-20 14:15:08.176263	completed	t	medium	medium	Add valuation	Add valuation on formalized datas.	2	183
\.


--
-- Data for Name: dbuser; Type: TABLE DATA; Schema: public; Owner: test
--

COPY public.dbuser (id, name, email, password, token, super_admin) FROM stdin;
1	root	admin@super.com	$argon2id$v=19$m=65536,t=3,p=4$JooiEtVXatRxSz16N9uo2g$Y2dAHdLAK06013FhDHQ/xhd+UL2yInwDAvRS1+KKD3c	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MDgxMDIxMDQsInN1cGVyX2FkbWluIjpmYWxzZSwidXNlcl9pZCI6InlheWFAZ21haWwuY29tIn0.pEZVXQ69PUTgLMOcF0X-9mfc4QW22nNJDE9MwSGji-w	t
2	yaya	yaya@gmail.com	$argon2id$v=19$m=65536,t=3,p=4$JooiEtVXatRxSz16N9uo2g$Y2dAHdLAK06013FhDHQ/xhd+UL2yInwDAvRS1+KKD3c	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MDgxMDIxMDQsInN1cGVyX2FkbWluIjpmYWxzZSwidXNlcl9pZCI6InlheWFAZ21haWwuY29tIn0.pEZVXQ69PUTgLMOcF0X-9mfc4QW22nNJDE9MwSGji-w	f
3	CDP	cdp@gmail.com	$argon2id$v=19$m=65536,t=3,p=4$JooiEtVXatRxSz16N9uo2g$Y2dAHdLAK06013FhDHQ/xhd+UL2yInwDAvRS1+KKD3c	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MDgxMDIxMDQsInN1cGVyX2FkbWluIjpmYWxzZSwidXNlcl9pZCI6InlheWFAZ21haWwuY29tIn0.pEZVXQ69PUTgLMOcF0X-9mfc4QW22nNJDE9MwSGji-w	f
4	CDG	cdg@gmail.com	$argon2id$v=19$m=65536,t=3,p=4$JooiEtVXatRxSz16N9uo2g$Y2dAHdLAK06013FhDHQ/xhd+UL2yInwDAvRS1+KKD3c	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MDgxMDIxMDQsInN1cGVyX2FkbWluIjpmYWxzZSwidXNlcl9pZCI6InlheWFAZ21haWwuY29tIn0.pEZVXQ69PUTgLMOcF0X-9mfc4QW22nNJDE9MwSGji-w	f
5	Cédric	cedric@gmail.com	$argon2id$v=19$m=65536,t=3,p=4$JooiEtVXatRxSz16N9uo2g$Y2dAHdLAK06013FhDHQ/xhd+UL2yInwDAvRS1+KKD3c	\N	f
6	juridic	juridic@gmail.com	$argon2id$v=19$m=65536,t=3,p=4$JooiEtVXatRxSz16N9uo2g$Y2dAHdLAK06013FhDHQ/xhd+UL2yInwDAvRS1+KKD3c		f
\.


--
-- Data for Name: dbview; Type: TABLE DATA; Schema: public; Owner: test
--

COPY public.dbview (id, name, is_list, indexable, description, readonly, index, sql_order, sql_view, sql_dir, through_perms, dbview_id, dbschema_id, is_empty, sql_restriction) FROM stdin;
2	Assigned activity	t	t	no description...	f	1	\N	name,state,description,urgency,priority, dbcreated_by_id	\N	\N	\N	5820	\N	is_close=false
7	Archived activity	t	t	no description...	t	1	\N	name,state,description,urgency,priority, dbcreated_by_id	\N	\N	\N	5820	\N	is_close=true
8	RH view of user	t	t	no description...	f	1	\N	\N	\N	\N	\N	135	\N	\N
9	Requests List View	t	t	no description...	t	1	\N	\N	\N	\N	\N	5804	\N	\N
1	Formalized Datas	t	t	no description...	t	1	\N	\N	\N	783	\N	309	\N	\N
5	New Request	f	t	no description...	f	0	\N	name,dbworkflow_id	\N	\N	\N	5804	t	\N
\.


--
-- Data for Name: dbworkflow; Type: TABLE DATA; Schema: public; Owner: test
--

COPY public.dbworkflow (id, name, description, dbschema_id) FROM stdin;
1	Formalized Datas Entry	workflow for Formalized Datas Entry	309
2	user creator	\N	5820
\.


--
-- Data for Name: dbworkflow_schema; Type: TABLE DATA; Schema: public; Owner: test
--

COPY public.dbworkflow_schema (id, dbworkflow_id, dbschema_id, index, description, name, dbuser_id, dbentity_id, urgency, priority) FROM stdin;
1	1	309	1	Add Juridic protection on formalized datas.	Add Juridic protection	\N	5	medium	medium
2	1	309	2	Add valuation on formalized datas.	Add valuation	\N	3	medium	medium
\.


--
-- Data for Name: formalized_data; Type: TABLE DATA; Schema: public; Owner: test
--

COPY public.formalized_data (id, name, ref, capitalization_date, first_evaluation, first_evaluation_date, actualized_evaluation, storage_area, contractual, result_family_id, result_type_id, support_id, confidentiality_level_id, restriction_type_id, valuation_id) FROM stdin;
190	qdqsd	T1000	2024-03-06 13:37:35.031363	\N	2024-03-06 13:37:35.039499	\N	TOULOUSE	f	1	3	2	2	\N	\N
169	Test 1000	T1000	\N	\N	\N	\N	TOULOUSE	f	1	1	1	\N	\N	\N
171	TEST 2000	T2000	\N	\N	\N	\N	TOULOUSE	f	1	1	1	2	\N	\N
173	TEST 3000	T3000	2024-02-29 00:00:00	1002	2024-02-29 00:00:00	1000	TOULOUSE	f	1	1	1	3	\N	\N
188	Test	T20000	\N	\N	\N	\N	TOULOUSE	f	1	1	1	1	\N	\N
179	New Request	NR1000	2024-02-29 00:00:00	1002	2024-02-29 00:00:00	1200	TOULOUSE	f	1	1	1	6	\N	\N
183	MyNewTest	T2000	2024-03-01 00:00:00	13213153	2024-03-01 00:00:00	5163513153	TOULOUSE	f	1	1	1	3	\N	\N
\.


--
-- Data for Name: formalized_data_project; Type: TABLE DATA; Schema: public; Owner: test
--

COPY public.formalized_data_project (id, formalized_data_id, project_id) FROM stdin;
67	188	1
223	190	1
16	169	1
17	171	1
18	173	1
20	179	1
23	183	1
\.


--
-- Data for Name: formalized_data_storage_type; Type: TABLE DATA; Schema: public; Owner: test
--

COPY public.formalized_data_storage_type (id, formalized_data_id, storage_type_id) FROM stdin;
12	173	3
13	173	1
17	179	1
18	179	2
19	179	3
20	179	4
441	190	2
442	190	4
443	190	5
444	190	6
28	183	1
29	183	2
30	183	3
31	183	4
32	183	5
185	188	1
186	188	5
\.


--
-- Data for Name: project; Type: TABLE DATA; Schema: public; Owner: test
--

COPY public.project (id, name, code, dbentity_id) FROM stdin;
1	eden	EDEN	1
\.


--
-- Data for Name: protection; Type: TABLE DATA; Schema: public; Owner: test
--

COPY public.protection (id, name, end_date, start_date, protection_area_id, protection_type_id, formalized_data_id) FROM stdin;
63	fhghy2	2024-03-06 00:00:00	2024-03-06 00:00:00	1	1	190
62	rrrrttttqsdd	2024-03-06 00:00:00	2024-03-06 00:00:00	1	1	190
56	qsdqsdqsd24	2024-03-06 00:00:00	2024-03-06 00:00:00	1	1	190
\.


--
-- Data for Name: protection_area; Type: TABLE DATA; Schema: public; Owner: test
--

COPY public.protection_area (id, name) FROM stdin;
1	france
2	UE
3	out of the UE
\.


--
-- Data for Name: protection_type; Type: TABLE DATA; Schema: public; Owner: test
--

COPY public.protection_type (id, name) FROM stdin;
1	SOLEAU envelope
2	timestamp
3	patent application filed
4	patent granted
5	APP registration certificate
6	no protection
\.


--
-- Data for Name: restriction_type; Type: TABLE DATA; Schema: public; Owner: test
--

COPY public.restriction_type (id, name, formalized_data_id) FROM stdin;
10	ppppp	108
11	yay	108
12	zeeee	173
13	SETUP	179
14	RESTRICTION HARD	179
15	my protection	183
16	my super restriction	183
28	aaaaa	190
29	bbbb	190
\.


--
-- Data for Name: result_family; Type: TABLE DATA; Schema: public; Owner: test
--

COPY public.result_family (id, name) FROM stdin;
1	report
2	internal technical note
3	laboratory workbook
4	career guide
5	manuscript article
6	thesis manuscript
7	innovation filing sheet
\.


--
-- Data for Name: result_type; Type: TABLE DATA; Schema: public; Owner: test
--

COPY public.result_type (id, name) FROM stdin;
1	process
2	method
3	algorithm
4	software
5	specification
6	conception
7	test results
8	sample/test vehicle
9	demonstrator/proof of concept
10	database
11	standard
12	innovation idea
13	irt test bench 
\.


--
-- Data for Name: storage_type; Type: TABLE DATA; Schema: public; Owner: test
--

COPY public.storage_type (id, name) FROM stdin;
1	IRT server
2	external data center
3	hard disk
4	USB
5	paper storager
6	physical conservation
\.


--
-- Data for Name: support; Type: TABLE DATA; Schema: public; Owner: test
--

COPY public.support (id, name) FROM stdin;
1	paper
2	digital
\.


--
-- Data for Name: valuation; Type: TABLE DATA; Schema: public; Owner: test
--

COPY public.valuation (id, name, valuation_type_id, valuation_format_id, formalized_data_id) FROM stdin;
1	My valuation test	1	1	173
2	valuation 2	1	1	173
3	My Valuation Particular	1	1	179
4	My valuation	1	1	183
5	my valuation to eden	1	1	190
\.


--
-- Data for Name: valuation_format; Type: TABLE DATA; Schema: public; Owner: test
--

COPY public.valuation_format (id, name) FROM stdin;
1	scientific journal article
2	conference presentation
3	published thesis dissertation
4	training
5	BIP contribution to a new project
6	license (patent/software)
7	service provision
\.


--
-- Data for Name: valuation_type; Type: TABLE DATA; Schema: public; Owner: test
--

COPY public.valuation_type (id, name) FROM stdin;
1	scientific
2	economic
\.


--
-- Name: confidentiality_level_id_seq; Type: SEQUENCE SET; Schema: public; Owner: test
--

SELECT pg_catalog.setval('public.confidentiality_level_id_seq', 7, true);


--
-- Name: dbentity_id_seq; Type: SEQUENCE SET; Schema: public; Owner: test
--

SELECT pg_catalog.setval('public.dbentity_id_seq', 5, true);


--
-- Name: dbentity_user_id_seq; Type: SEQUENCE SET; Schema: public; Owner: test
--

SELECT pg_catalog.setval('public.dbentity_user_id_seq', 5, true);


--
-- Name: dbhierarchy_id_seq; Type: SEQUENCE SET; Schema: public; Owner: test
--

SELECT pg_catalog.setval('public.dbhierarchy_id_seq', 1, true);


--
-- Name: dbpermission_id_seq; Type: SEQUENCE SET; Schema: public; Owner: test
--

SELECT pg_catalog.setval('public.dbpermission_id_seq', 2807, true);


--
-- Name: dbrequest_id_seq; Type: SEQUENCE SET; Schema: public; Owner: test
--

SELECT pg_catalog.setval('public.dbrequest_id_seq', 51, true);


--
-- Name: dbrole_attribution_id_seq; Type: SEQUENCE SET; Schema: public; Owner: test
--

SELECT pg_catalog.setval('public.dbrole_attribution_id_seq', 6, true);


--
-- Name: dbrole_id_seq; Type: SEQUENCE SET; Schema: public; Owner: test
--

SELECT pg_catalog.setval('public.dbrole_id_seq', 5, true);


--
-- Name: dbrole_permission_id_seq; Type: SEQUENCE SET; Schema: public; Owner: test
--

SELECT pg_catalog.setval('public.dbrole_permission_id_seq', 73, true);


--
-- Name: dbschema_column_id_seq; Type: SEQUENCE SET; Schema: public; Owner: test
--

SELECT pg_catalog.setval('public.dbschema_column_id_seq', 993, true);


--
-- Name: dbschema_id_seq; Type: SEQUENCE SET; Schema: public; Owner: test
--

SELECT pg_catalog.setval('public.dbschema_id_seq', 10495, true);


--
-- Name: dbtask_id_seq; Type: SEQUENCE SET; Schema: public; Owner: test
--

SELECT pg_catalog.setval('public.dbtask_id_seq', 120, true);


--
-- Name: dbuser_id_seq; Type: SEQUENCE SET; Schema: public; Owner: test
--

SELECT pg_catalog.setval('public.dbuser_id_seq', 6, true);


--
-- Name: dbview_id_seq; Type: SEQUENCE SET; Schema: public; Owner: test
--

SELECT pg_catalog.setval('public.dbview_id_seq', 9, true);


--
-- Name: dbworkflow_id_seq; Type: SEQUENCE SET; Schema: public; Owner: test
--

SELECT pg_catalog.setval('public.dbworkflow_id_seq', 2, true);


--
-- Name: dbworkflow_schema_id_seq; Type: SEQUENCE SET; Schema: public; Owner: test
--

SELECT pg_catalog.setval('public.dbworkflow_schema_id_seq', 2, true);


--
-- Name: formalized_data_id_seq; Type: SEQUENCE SET; Schema: public; Owner: test
--

SELECT pg_catalog.setval('public.formalized_data_id_seq', 194, true);


--
-- Name: formalized_data_project_id_seq; Type: SEQUENCE SET; Schema: public; Owner: test
--

SELECT pg_catalog.setval('public.formalized_data_project_id_seq', 223, true);


--
-- Name: formalized_data_storage_type_id_seq; Type: SEQUENCE SET; Schema: public; Owner: test
--

SELECT pg_catalog.setval('public.formalized_data_storage_type_id_seq', 444, true);


--
-- Name: project_id_seq; Type: SEQUENCE SET; Schema: public; Owner: test
--

SELECT pg_catalog.setval('public.project_id_seq', 1, true);


--
-- Name: protection_area_id_seq; Type: SEQUENCE SET; Schema: public; Owner: test
--

SELECT pg_catalog.setval('public.protection_area_id_seq', 3, true);


--
-- Name: protection_id_seq; Type: SEQUENCE SET; Schema: public; Owner: test
--

SELECT pg_catalog.setval('public.protection_id_seq', 63, true);


--
-- Name: protection_type_id_seq; Type: SEQUENCE SET; Schema: public; Owner: test
--

SELECT pg_catalog.setval('public.protection_type_id_seq', 6, true);


--
-- Name: restriction_type_id_seq; Type: SEQUENCE SET; Schema: public; Owner: test
--

SELECT pg_catalog.setval('public.restriction_type_id_seq', 29, true);


--
-- Name: result_family_id_seq; Type: SEQUENCE SET; Schema: public; Owner: test
--

SELECT pg_catalog.setval('public.result_family_id_seq', 7, true);


--
-- Name: result_type_id_seq; Type: SEQUENCE SET; Schema: public; Owner: test
--

SELECT pg_catalog.setval('public.result_type_id_seq', 13, true);


--
-- Name: sq_dbentity; Type: SEQUENCE SET; Schema: public; Owner: test
--

SELECT pg_catalog.setval('public.sq_dbentity', 1, true);


--
-- Name: sq_dbform; Type: SEQUENCE SET; Schema: public; Owner: test
--

SELECT pg_catalog.setval('public.sq_dbform', 4, true);


--
-- Name: sq_dbformfields; Type: SEQUENCE SET; Schema: public; Owner: test
--

SELECT pg_catalog.setval('public.sq_dbformfields', 2, true);


--
-- Name: sq_dbrole; Type: SEQUENCE SET; Schema: public; Owner: test
--

SELECT pg_catalog.setval('public.sq_dbrole', 2, true);


--
-- Name: sq_dbtableaccess; Type: SEQUENCE SET; Schema: public; Owner: test
--

SELECT pg_catalog.setval('public.sq_dbtableaccess', 1, false);


--
-- Name: sq_dbtableview; Type: SEQUENCE SET; Schema: public; Owner: test
--

SELECT pg_catalog.setval('public.sq_dbtableview', 1, false);


--
-- Name: sq_dbuser; Type: SEQUENCE SET; Schema: public; Owner: test
--

SELECT pg_catalog.setval('public.sq_dbuser', 214, true);


--
-- Name: sq_dbuserrole; Type: SEQUENCE SET; Schema: public; Owner: test
--

SELECT pg_catalog.setval('public.sq_dbuserrole', 2, true);


--
-- Name: storage_type_id_seq; Type: SEQUENCE SET; Schema: public; Owner: test
--

SELECT pg_catalog.setval('public.storage_type_id_seq', 6, true);


--
-- Name: support_id_seq; Type: SEQUENCE SET; Schema: public; Owner: test
--

SELECT pg_catalog.setval('public.support_id_seq', 2, true);


--
-- Name: valuation_format_id_seq; Type: SEQUENCE SET; Schema: public; Owner: test
--

SELECT pg_catalog.setval('public.valuation_format_id_seq', 7, true);


--
-- Name: valuation_id_seq; Type: SEQUENCE SET; Schema: public; Owner: test
--

SELECT pg_catalog.setval('public.valuation_id_seq', 5, true);


--
-- Name: valuation_type_id_seq; Type: SEQUENCE SET; Schema: public; Owner: test
--

SELECT pg_catalog.setval('public.valuation_type_id_seq', 2, true);


--
-- Name: confidentiality_level confidentiality_level_name_key; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.confidentiality_level
    ADD CONSTRAINT confidentiality_level_name_key UNIQUE (name);


--
-- Name: confidentiality_level confidentiality_level_pkey; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.confidentiality_level
    ADD CONSTRAINT confidentiality_level_pkey PRIMARY KEY (id);


--
-- Name: dbentity dbentity_pkey; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbentity
    ADD CONSTRAINT dbentity_pkey PRIMARY KEY (id);


--
-- Name: dbentity_user dbentity_user_pkey; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbentity_user
    ADD CONSTRAINT dbentity_user_pkey PRIMARY KEY (id);


--
-- Name: dbhierarchy dbhierarchy_pkey; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbhierarchy
    ADD CONSTRAINT dbhierarchy_pkey PRIMARY KEY (id);


--
-- Name: dbpermission dbpermission_name_unique; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbpermission
    ADD CONSTRAINT dbpermission_name_unique UNIQUE (name);


--
-- Name: dbpermission dbpermission_pkey; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbpermission
    ADD CONSTRAINT dbpermission_pkey PRIMARY KEY (id);


--
-- Name: dbrequest dbrequest_pkey; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbrequest
    ADD CONSTRAINT dbrequest_pkey PRIMARY KEY (id);


--
-- Name: dbrole_attribution dbrole_attribution_pkey; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbrole_attribution
    ADD CONSTRAINT dbrole_attribution_pkey PRIMARY KEY (id);


--
-- Name: dbrole dbrole_name_unique; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbrole
    ADD CONSTRAINT dbrole_name_unique UNIQUE (name);


--
-- Name: dbrole_permission dbrole_permission_pkey; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbrole_permission
    ADD CONSTRAINT dbrole_permission_pkey PRIMARY KEY (id);


--
-- Name: dbrole dbrole_pkey; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbrole
    ADD CONSTRAINT dbrole_pkey PRIMARY KEY (id);


--
-- Name: dbschema_column dbschema_column_pkey; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbschema_column
    ADD CONSTRAINT dbschema_column_pkey PRIMARY KEY (id);


--
-- Name: dbschema dbschema_name_unique; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbschema
    ADD CONSTRAINT dbschema_name_unique UNIQUE (name);


--
-- Name: dbschema dbschema_pkey; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbschema
    ADD CONSTRAINT dbschema_pkey PRIMARY KEY (id);


--
-- Name: dbtask dbtask_pkey; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbtask
    ADD CONSTRAINT dbtask_pkey PRIMARY KEY (id);


--
-- Name: dbuser dbuser_email_unique; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbuser
    ADD CONSTRAINT dbuser_email_unique UNIQUE (email);


--
-- Name: dbuser dbuser_name_unique; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbuser
    ADD CONSTRAINT dbuser_name_unique UNIQUE (name);


--
-- Name: dbuser dbuser_pkey; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbuser
    ADD CONSTRAINT dbuser_pkey PRIMARY KEY (id);


--
-- Name: dbview dbview_pkey; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbview
    ADD CONSTRAINT dbview_pkey PRIMARY KEY (id);


--
-- Name: dbworkflow dbworkflow_name_unique; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbworkflow
    ADD CONSTRAINT dbworkflow_name_unique UNIQUE (name);


--
-- Name: dbworkflow dbworkflow_pkey; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbworkflow
    ADD CONSTRAINT dbworkflow_pkey PRIMARY KEY (id);


--
-- Name: dbworkflow_schema dbworkflow_schema_name_unique; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbworkflow_schema
    ADD CONSTRAINT dbworkflow_schema_name_unique UNIQUE (name);


--
-- Name: dbworkflow_schema dbworkflow_schema_pkey; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbworkflow_schema
    ADD CONSTRAINT dbworkflow_schema_pkey PRIMARY KEY (id);


--
-- Name: formalized_data formalized_data_name_key; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.formalized_data
    ADD CONSTRAINT formalized_data_name_key UNIQUE (name);


--
-- Name: formalized_data formalized_data_pkey; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.formalized_data
    ADD CONSTRAINT formalized_data_pkey PRIMARY KEY (id);


--
-- Name: formalized_data_project formalized_data_project_pkey; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.formalized_data_project
    ADD CONSTRAINT formalized_data_project_pkey PRIMARY KEY (id);


--
-- Name: formalized_data_storage_type formalized_data_storage_type_pkey; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.formalized_data_storage_type
    ADD CONSTRAINT formalized_data_storage_type_pkey PRIMARY KEY (id);


--
-- Name: project project_name_key; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.project
    ADD CONSTRAINT project_name_key UNIQUE (name);


--
-- Name: project project_pkey; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.project
    ADD CONSTRAINT project_pkey PRIMARY KEY (id);


--
-- Name: protection_area protection_area_name_key; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.protection_area
    ADD CONSTRAINT protection_area_name_key UNIQUE (name);


--
-- Name: protection_area protection_area_pkey; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.protection_area
    ADD CONSTRAINT protection_area_pkey PRIMARY KEY (id);


--
-- Name: protection protection_name_key; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.protection
    ADD CONSTRAINT protection_name_key UNIQUE (name);


--
-- Name: protection protection_pkey; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.protection
    ADD CONSTRAINT protection_pkey PRIMARY KEY (id);


--
-- Name: protection_type protection_type_name_key; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.protection_type
    ADD CONSTRAINT protection_type_name_key UNIQUE (name);


--
-- Name: protection_type protection_type_pkey; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.protection_type
    ADD CONSTRAINT protection_type_pkey PRIMARY KEY (id);


--
-- Name: restriction_type restriction_type_name_key; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.restriction_type
    ADD CONSTRAINT restriction_type_name_key UNIQUE (name);


--
-- Name: restriction_type restriction_type_pkey; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.restriction_type
    ADD CONSTRAINT restriction_type_pkey PRIMARY KEY (id);


--
-- Name: result_family result_family_name_key; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.result_family
    ADD CONSTRAINT result_family_name_key UNIQUE (name);


--
-- Name: result_family result_family_pkey; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.result_family
    ADD CONSTRAINT result_family_pkey PRIMARY KEY (id);


--
-- Name: result_type result_type_name_key; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.result_type
    ADD CONSTRAINT result_type_name_key UNIQUE (name);


--
-- Name: result_type result_type_pkey; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.result_type
    ADD CONSTRAINT result_type_pkey PRIMARY KEY (id);


--
-- Name: storage_type storage_type_name_key; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.storage_type
    ADD CONSTRAINT storage_type_name_key UNIQUE (name);


--
-- Name: storage_type storage_type_pkey; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.storage_type
    ADD CONSTRAINT storage_type_pkey PRIMARY KEY (id);


--
-- Name: support support_pkey; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.support
    ADD CONSTRAINT support_pkey PRIMARY KEY (id);


--
-- Name: valuation_format valuation_format_name_key; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.valuation_format
    ADD CONSTRAINT valuation_format_name_key UNIQUE (name);


--
-- Name: valuation_format valuation_format_pkey; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.valuation_format
    ADD CONSTRAINT valuation_format_pkey PRIMARY KEY (id);


--
-- Name: valuation valuation_name_key; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.valuation
    ADD CONSTRAINT valuation_name_key UNIQUE (name);


--
-- Name: valuation valuation_pkey; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.valuation
    ADD CONSTRAINT valuation_pkey PRIMARY KEY (id);


--
-- Name: valuation_type valuation_type_name_key; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.valuation_type
    ADD CONSTRAINT valuation_type_name_key UNIQUE (name);


--
-- Name: valuation_type valuation_type_pkey; Type: CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.valuation_type
    ADD CONSTRAINT valuation_type_pkey PRIMARY KEY (id);


--
-- Name: dbrequest fk_dbcreated_by_id; Type: FK CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbrequest
    ADD CONSTRAINT fk_dbcreated_by_id FOREIGN KEY (dbcreated_by_id) REFERENCES public.dbuser(id);


--
-- Name: dbtask fk_dbcreated_by_id; Type: FK CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbtask
    ADD CONSTRAINT fk_dbcreated_by_id FOREIGN KEY (dbcreated_by_id) REFERENCES public.dbuser(id);


--
-- Name: dbworkflow_schema fk_dbentity_assignee_id; Type: FK CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbworkflow_schema
    ADD CONSTRAINT fk_dbentity_assignee_id FOREIGN KEY (dbentity_id) REFERENCES public.dbentity(id);


--
-- Name: dbentity_user fk_dbentity_id; Type: FK CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbentity_user
    ADD CONSTRAINT fk_dbentity_id FOREIGN KEY (dbentity_id) REFERENCES public.dbentity(id);


--
-- Name: dbrole_attribution fk_dbentity_id; Type: FK CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbrole_attribution
    ADD CONSTRAINT fk_dbentity_id FOREIGN KEY (dbentity_id) REFERENCES public.dbentity(id);


--
-- Name: dbtask fk_dbentity_id; Type: FK CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbtask
    ADD CONSTRAINT fk_dbentity_id FOREIGN KEY (dbentity_id) REFERENCES public.dbentity(id);


--
-- Name: dbrole_permission fk_dbpermission_id; Type: FK CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbrole_permission
    ADD CONSTRAINT fk_dbpermission_id FOREIGN KEY (dbpermission_id) REFERENCES public.dbpermission(id);


--
-- Name: dbtask fk_dbrequest_id; Type: FK CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbtask
    ADD CONSTRAINT fk_dbrequest_id FOREIGN KEY (dbrequest_id) REFERENCES public.dbrequest(id);


--
-- Name: dbrole_attribution fk_dbrole_id; Type: FK CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbrole_attribution
    ADD CONSTRAINT fk_dbrole_id FOREIGN KEY (dbrole_id) REFERENCES public.dbrole(id);


--
-- Name: dbrole_permission fk_dbrole_id; Type: FK CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbrole_permission
    ADD CONSTRAINT fk_dbrole_id FOREIGN KEY (dbrole_id) REFERENCES public.dbrole(id);


--
-- Name: dbworkflow_schema fk_dbschema_id; Type: FK CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbworkflow_schema
    ADD CONSTRAINT fk_dbschema_id FOREIGN KEY (dbschema_id) REFERENCES public.dbschema(id);


--
-- Name: dbview fk_dbschema_id; Type: FK CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbview
    ADD CONSTRAINT fk_dbschema_id FOREIGN KEY (dbschema_id) REFERENCES public.dbschema(id);


--
-- Name: dbworkflow fk_dbschema_id; Type: FK CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbworkflow
    ADD CONSTRAINT fk_dbschema_id FOREIGN KEY (dbschema_id) REFERENCES public.dbschema(id);


--
-- Name: dbrequest fk_dbschema_id; Type: FK CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbrequest
    ADD CONSTRAINT fk_dbschema_id FOREIGN KEY (dbschema_id) REFERENCES public.dbschema(id);


--
-- Name: dbtask fk_dbschema_id; Type: FK CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbtask
    ADD CONSTRAINT fk_dbschema_id FOREIGN KEY (dbschema_id) REFERENCES public.dbschema(id);


--
-- Name: dbschema_column fk_dbschema_id; Type: FK CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbschema_column
    ADD CONSTRAINT fk_dbschema_id FOREIGN KEY (dbschema_id) REFERENCES public.dbschema(id);


--
-- Name: dbworkflow_schema fk_dbuser_assignee_id; Type: FK CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbworkflow_schema
    ADD CONSTRAINT fk_dbuser_assignee_id FOREIGN KEY (dbuser_id) REFERENCES public.dbuser(id);


--
-- Name: dbentity_user fk_dbuser_id; Type: FK CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbentity_user
    ADD CONSTRAINT fk_dbuser_id FOREIGN KEY (dbuser_id) REFERENCES public.dbuser(id);


--
-- Name: dbrole_attribution fk_dbuser_id; Type: FK CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbrole_attribution
    ADD CONSTRAINT fk_dbuser_id FOREIGN KEY (dbuser_id) REFERENCES public.dbuser(id);


--
-- Name: dbtask fk_dbuser_id; Type: FK CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbtask
    ADD CONSTRAINT fk_dbuser_id FOREIGN KEY (dbuser_id) REFERENCES public.dbuser(id);


--
-- Name: dbview fk_dbview_id; Type: FK CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbview
    ADD CONSTRAINT fk_dbview_id FOREIGN KEY (dbview_id) REFERENCES public.dbview(id);


--
-- Name: dbworkflow_schema fk_dbworkflow_id; Type: FK CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbworkflow_schema
    ADD CONSTRAINT fk_dbworkflow_id FOREIGN KEY (dbworkflow_id) REFERENCES public.dbworkflow(id);


--
-- Name: dbrequest fk_dbworkflow_id; Type: FK CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbrequest
    ADD CONSTRAINT fk_dbworkflow_id FOREIGN KEY (dbworkflow_id) REFERENCES public.dbworkflow(id);


--
-- Name: dbtask fk_dbworkflow_schema_id; Type: FK CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbtask
    ADD CONSTRAINT fk_dbworkflow_schema_id FOREIGN KEY (dbworkflow_schema_id) REFERENCES public.dbworkflow_schema(id);


--
-- Name: formalized_data_project fk_formalized_data_id; Type: FK CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.formalized_data_project
    ADD CONSTRAINT fk_formalized_data_id FOREIGN KEY (formalized_data_id) REFERENCES public.formalized_data(id);


--
-- Name: formalized_data_project fk_project_id; Type: FK CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.formalized_data_project
    ADD CONSTRAINT fk_project_id FOREIGN KEY (project_id) REFERENCES public.project(id);


--
-- Name: formalized_data_storage_type fk_storage_type_id; Type: FK CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.formalized_data_storage_type
    ADD CONSTRAINT fk_storage_type_id FOREIGN KEY (storage_type_id) REFERENCES public.storage_type(id);


--
-- Name: dbview fk_through_perms; Type: FK CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.dbview
    ADD CONSTRAINT fk_through_perms FOREIGN KEY (through_perms) REFERENCES public.dbschema(id);


--
-- Name: formalized_data_storage_type formalized_data_storage_type_formalized_data_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: test
--

ALTER TABLE ONLY public.formalized_data_storage_type
    ADD CONSTRAINT formalized_data_storage_type_formalized_data_id_fkey FOREIGN KEY (formalized_data_id) REFERENCES public.formalized_data(id);


--
-- PostgreSQL database dump complete
--

