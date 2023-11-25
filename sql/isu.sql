create index livestream_tags_livestream_id_idx on livestream_tags (livestream_id);

create index livestream_user_id_idx on livestreams (user_id);

create index icons_user_id_idx on icons (user_id);

create index themes_user_id_idx on themes (user_id);

create index ng_words_user_id_livestream_id_idx on ng_words (user_id, livestream_id);

create index resevation_slots_start_at_end_at_idx on reservation_slots (start_at, end_at);
