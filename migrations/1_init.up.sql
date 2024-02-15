create table public.prompts
(
    id           bigserial
        constraint prompts_pk primary key,
    prompt_text  text,
    prompt_stage bigint
);

create unique index if not exists prompts_prompt_text_prompt_stage_uindex
    on public.prompts (prompt_text, prompt_stage);