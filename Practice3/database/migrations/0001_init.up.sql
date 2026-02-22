create table if not exists users (
    id serial primary key,
    name varchar(255) not null,
    email varchar(255) unique not null,
    age int not null,
    created_at  timestamp default now()
);

insert into users(name, email, age) values ('John Doe', 'john@example.com', 30);