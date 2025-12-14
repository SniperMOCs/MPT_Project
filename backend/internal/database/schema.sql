--
-- PostgreSQL database dump
--

\restrict kO4rCcDd5DG1NwPkKGeYtxdEgxiuZmKSpZOtBRnhOzRsHWFeTEMqsixdwSRcf4x

-- Dumped from database version 18.1
-- Dumped by pg_dump version 18.1

-- Started on 2025-12-15 02:34:58

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
-- TOC entry 860 (class 1247 OID 16432)
-- Name: content_type_enum; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.content_type_enum AS ENUM (
    'POST',
    'STORY',
    'CLIP'
);


--
-- TOC entry 863 (class 1247 OID 16440)
-- Name: notification_type_enum; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.notification_type_enum AS ENUM (
    'LIKE',
    'COMMENT',
    'MENTION',
    'FOLLOW'
);


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- TOC entry 224 (class 1259 OID 16545)
-- Name: chat_message; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.chat_message (
    id uuid DEFAULT gen_random_uuid() CONSTRAINT chatmessage_id_not_null NOT NULL,
    chat_id uuid CONSTRAINT chatmessage_chat_id_not_null NOT NULL,
    sender_id uuid CONSTRAINT chatmessage_sender_id_not_null NOT NULL,
    content text CONSTRAINT chatmessage_content_not_null NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT message_content_not_empty_check CHECK ((length(TRIM(BOTH FROM content)) > 0))
);


--
-- TOC entry 223 (class 1259 OID 16523)
-- Name: chat_user; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.chat_user (
    id uuid DEFAULT gen_random_uuid() CONSTRAINT chatuser_id_not_null NOT NULL,
    chat_id uuid CONSTRAINT chatuser_chat_id_not_null NOT NULL,
    user_id uuid CONSTRAINT chatuser_user_id_not_null NOT NULL,
    joined_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- TOC entry 222 (class 1259 OID 16514)
-- Name: chats; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.chats (
    id uuid DEFAULT gen_random_uuid() CONSTRAINT chat_id_not_null NOT NULL,
    title character varying(100),
    is_group boolean DEFAULT false,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- TOC entry 221 (class 1259 OID 16485)
-- Name: comments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.comments (
    id uuid DEFAULT gen_random_uuid() CONSTRAINT comment_id_not_null NOT NULL,
    author_id uuid CONSTRAINT comment_author_id_not_null NOT NULL,
    parent_id uuid,
    content_id uuid CONSTRAINT comment_content_id_not_null NOT NULL,
    text text CONSTRAINT comment_text_not_null NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT comment_text_not_empty_check CHECK ((length(TRIM(BOTH FROM text)) > 0))
);


--
-- TOC entry 220 (class 1259 OID 16466)
-- Name: contents; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.contents (
    id uuid DEFAULT gen_random_uuid() CONSTRAINT content_id_not_null NOT NULL,
    author_id uuid CONSTRAINT content_author_id_not_null NOT NULL,
    content_type public.content_type_enum CONSTRAINT content_content_type_not_null NOT NULL,
    description text,
    media_urls text[],
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    is_visible boolean DEFAULT true,
    CONSTRAINT media_urls_not_empty_check CHECK ((array_length(media_urls, 1) > 0))
);


--
-- TOC entry 226 (class 1259 OID 16590)
-- Name: follows; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.follows (
    id uuid DEFAULT gen_random_uuid() CONSTRAINT follow_id_not_null NOT NULL,
    follower_id uuid CONSTRAINT follow_follower_id_not_null NOT NULL,
    following_id uuid CONSTRAINT follow_following_id_not_null NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT no_self_follow_check CHECK ((follower_id <> following_id))
);


--
-- TOC entry 227 (class 1259 OID 16613)
-- Name: likes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.likes (
    id uuid DEFAULT gen_random_uuid() CONSTRAINT "Like_id_not_null" NOT NULL,
    user_id uuid CONSTRAINT "Like_user_id_not_null" NOT NULL,
    content_id uuid CONSTRAINT "Like_content_id_not_null" NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- TOC entry 225 (class 1259 OID 16569)
-- Name: notifications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications (
    id uuid DEFAULT gen_random_uuid() CONSTRAINT notification_id_not_null NOT NULL,
    user_id uuid CONSTRAINT notification_user_id_not_null NOT NULL,
    content_id uuid,
    notification_type public.notification_type_enum CONSTRAINT notification_notification_type_not_null NOT NULL,
    content_type public.content_type_enum,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- TOC entry 219 (class 1259 OID 16449)
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id uuid DEFAULT gen_random_uuid() CONSTRAINT "User_id_not_null" NOT NULL,
    username character varying(50) CONSTRAINT "User_username_not_null" NOT NULL,
    full_name character varying(100),
    password character varying(255) CONSTRAINT "User_password_not_null" NOT NULL,
    avatar_url text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    description text,
    CONSTRAINT full_name_length_check CHECK ((length((full_name)::text) >= 2)),
    CONSTRAINT username_length_check CHECK ((length((username)::text) >= 3))
);


--
-- TOC entry 4942 (class 2606 OID 16622)
-- Name: likes Like_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.likes
    ADD CONSTRAINT "Like_pkey" PRIMARY KEY (id);


--
-- TOC entry 4920 (class 2606 OID 16463)
-- Name: users User_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT "User_pkey" PRIMARY KEY (id);


--
-- TOC entry 4922 (class 2606 OID 16465)
-- Name: users User_username_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT "User_username_key" UNIQUE (username);


--
-- TOC entry 4928 (class 2606 OID 16522)
-- Name: chats chat_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chats
    ADD CONSTRAINT chat_pkey PRIMARY KEY (id);


--
-- TOC entry 4934 (class 2606 OID 16558)
-- Name: chat_message chatmessage_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chat_message
    ADD CONSTRAINT chatmessage_pkey PRIMARY KEY (id);


--
-- TOC entry 4930 (class 2606 OID 16532)
-- Name: chat_user chatuser_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chat_user
    ADD CONSTRAINT chatuser_pkey PRIMARY KEY (id);


--
-- TOC entry 4926 (class 2606 OID 16498)
-- Name: comments comment_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comment_pkey PRIMARY KEY (id);


--
-- TOC entry 4924 (class 2606 OID 16479)
-- Name: contents content_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.contents
    ADD CONSTRAINT content_pkey PRIMARY KEY (id);


--
-- TOC entry 4938 (class 2606 OID 16600)
-- Name: follows follow_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.follows
    ADD CONSTRAINT follow_pkey PRIMARY KEY (id);


--
-- TOC entry 4936 (class 2606 OID 16579)
-- Name: notifications notification_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notification_pkey PRIMARY KEY (id);


--
-- TOC entry 4932 (class 2606 OID 16534)
-- Name: chat_user unique_chat_user; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chat_user
    ADD CONSTRAINT unique_chat_user UNIQUE (chat_id, user_id);


--
-- TOC entry 4940 (class 2606 OID 16602)
-- Name: follows unique_follow_pair; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.follows
    ADD CONSTRAINT unique_follow_pair UNIQUE (follower_id, following_id);


--
-- TOC entry 4944 (class 2606 OID 16624)
-- Name: likes unique_like; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.likes
    ADD CONSTRAINT unique_like UNIQUE (user_id, content_id);


--
-- TOC entry 4951 (class 2606 OID 16559)
-- Name: chat_message fk_chatmessage_chat; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chat_message
    ADD CONSTRAINT fk_chatmessage_chat FOREIGN KEY (chat_id) REFERENCES public.chats(id) ON DELETE CASCADE;


--
-- TOC entry 4952 (class 2606 OID 16564)
-- Name: chat_message fk_chatmessage_sender; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chat_message
    ADD CONSTRAINT fk_chatmessage_sender FOREIGN KEY (sender_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- TOC entry 4949 (class 2606 OID 16535)
-- Name: chat_user fk_chatuser_chat; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chat_user
    ADD CONSTRAINT fk_chatuser_chat FOREIGN KEY (chat_id) REFERENCES public.chats(id) ON DELETE CASCADE;


--
-- TOC entry 4950 (class 2606 OID 16540)
-- Name: chat_user fk_chatuser_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chat_user
    ADD CONSTRAINT fk_chatuser_user FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- TOC entry 4946 (class 2606 OID 16499)
-- Name: comments fk_comment_author; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT fk_comment_author FOREIGN KEY (author_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- TOC entry 4947 (class 2606 OID 16509)
-- Name: comments fk_comment_content; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT fk_comment_content FOREIGN KEY (content_id) REFERENCES public.contents(id) ON DELETE CASCADE;


--
-- TOC entry 4948 (class 2606 OID 16504)
-- Name: comments fk_comment_parent; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT fk_comment_parent FOREIGN KEY (parent_id) REFERENCES public.comments(id) ON DELETE CASCADE;


--
-- TOC entry 4945 (class 2606 OID 16480)
-- Name: contents fk_content_author; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.contents
    ADD CONSTRAINT fk_content_author FOREIGN KEY (author_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- TOC entry 4955 (class 2606 OID 16603)
-- Name: follows fk_follow_follower; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.follows
    ADD CONSTRAINT fk_follow_follower FOREIGN KEY (follower_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- TOC entry 4956 (class 2606 OID 16608)
-- Name: follows fk_follow_following; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.follows
    ADD CONSTRAINT fk_follow_following FOREIGN KEY (following_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- TOC entry 4957 (class 2606 OID 16630)
-- Name: likes fk_like_content; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.likes
    ADD CONSTRAINT fk_like_content FOREIGN KEY (content_id) REFERENCES public.contents(id) ON DELETE CASCADE;


--
-- TOC entry 4958 (class 2606 OID 16625)
-- Name: likes fk_like_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.likes
    ADD CONSTRAINT fk_like_user FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- TOC entry 4953 (class 2606 OID 16585)
-- Name: notifications fk_notification_content; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT fk_notification_content FOREIGN KEY (content_id) REFERENCES public.contents(id) ON DELETE SET NULL;


--
-- TOC entry 4954 (class 2606 OID 16580)
-- Name: notifications fk_notification_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT fk_notification_user FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


-- Completed on 2025-12-15 02:34:59

--
-- PostgreSQL database dump complete
--

\unrestrict kO4rCcDd5DG1NwPkKGeYtxdEgxiuZmKSpZOtBRnhOzRsHWFeTEMqsixdwSRcf4x

