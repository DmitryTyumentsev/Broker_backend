alter table users
    alter column id type uuid
using id::uuid;
