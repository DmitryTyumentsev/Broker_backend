alter table users
    alter column id type uuid
        using id::uuid,
    add constraint users_email_unique unique (email);