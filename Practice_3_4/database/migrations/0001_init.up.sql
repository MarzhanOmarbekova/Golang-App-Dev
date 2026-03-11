CREATE TABLE IF NOT EXISTS users (
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(255)        NOT NULL,
    email      VARCHAR(255) UNIQUE NOT NULL,
    age        INT                 NOT NULL,
    gender     VARCHAR(10)         NOT NULL DEFAULT 'other',
    birth_date DATE                NOT NULL DEFAULT '2000-01-01',
    created_at TIMESTAMP           DEFAULT NOW()
    );

CREATE TABLE IF NOT EXISTS user_friends (
                                            user_id   INTEGER REFERENCES users(id) ON DELETE CASCADE,
    friend_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, friend_id),
    CONSTRAINT no_self_friendship CHECK (user_id <> friend_id)
    );


INSERT INTO users (name, email, age, gender, birth_date) VALUES
('Alice Johnson',   'alice@example.com',   28, 'female', '1996-03-15'),
('Bob Smith',       'bob@example.com',     32, 'male',   '1992-07-22'),
('Carol White',     'carol@example.com',   25, 'female', '1999-11-05'),
('David Brown',     'david@example.com',   35, 'male',   '1989-01-30'),
('Eva Martinez',    'eva@example.com',     27, 'female', '1997-06-18'),
('Frank Lee',       'frank@example.com',   30, 'male',   '1994-09-12'),
('Grace Kim',       'grace@example.com',   22, 'female', '2002-04-03'),
('Henry Wilson',    'henry@example.com',   40, 'male',   '1984-12-25'),
('Ivy Chen',        'ivy@example.com',     29, 'female', '1995-08-17'),
('Jack Taylor',     'jack@example.com',    33, 'male',   '1991-02-28'),
('Karen Davis',     'karen@example.com',   26, 'female', '1998-05-10'),
('Leo Garcia',      'leo@example.com',     31, 'male',   '1993-10-07'),
('Mia Rodriguez',   'mia@example.com',     24, 'female', '2000-07-14'),
('Nathan Harris',   'nathan@example.com',  38, 'male',   '1986-03-21'),
('Olivia Clark',    'olivia@example.com',  23, 'female', '2001-09-09'),
('Paul Lewis',      'paul@example.com',    36, 'male',   '1988-11-16'),
('Quinn Walker',    'quinn@example.com',   21, 'other',  '2003-01-27'),
('Rachel Hall',     'rachel@example.com',  34, 'female', '1990-06-04'),
('Sam Young',       'sam@example.com',     37, 'male',   '1987-04-19'),
('Tina Allen',      'tina@example.com',    20, 'female', '2004-08-31')
    ON CONFLICT DO NOTHING;

INSERT INTO user_friends (user_id, friend_id) VALUES
                                                  (1,2),(2,1),
                                                  (1,3),(3,1),
                                                  (1,4),(4,1),
                                                  (1,5),(5,1),
                                                  (1,6),(6,1),
                                                  (1,7),(7,1);

INSERT INTO user_friends (user_id, friend_id) VALUES
                                                  (2,3),(3,2),
                                                  (2,4),(4,2),
                                                  (2,5),(5,2),
                                                  (2,6),(6,2),
                                                  (2,8),(8,2);

INSERT INTO user_friends (user_id, friend_id) VALUES
                                                  (3,9),(9,3),
                                                  (4,10),(10,4),
                                                  (5,11),(11,5),
                                                  (6,12),(12,6),
                                                  (7,13),(13,7),
                                                  (8,14),(14,8),
                                                  (9,15),(15,9),
                                                  (10,16),(16,10),
                                                  (11,17),(17,11),
                                                  (12,18),(18,12),
                                                  (13,19),(19,13),
                                                  (14,20),(20,14)
    ON CONFLICT DO NOTHING;