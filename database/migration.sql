/* Migration 25.9.2017 */
CREATE DATABASE IF NOT EXISTS haloochat;

CREATE TABLE IF NOT EXISTS chat_users 
    (id SERIAL PRIMARY KEY, 
    name VARCHAR(255), 
    email VARCHAR(255), 
    password VARCHAR(255), 
    last_seen TIMESTAMPTZ, 
    profile_picture VARCHAR(255));
    
CREATE TABLE IF NOT EXISTS chatlog 
    (id SERIAL PRIMARY KEY, 
    sender SERIAL, 
    receiver SERIAL, 
    message TEXT, 
    timestamp TIMESTAMPTZ);

/* Migration 25.10.2017 */

CREATE TABLE IF NOT EXISTS rooms
    (id SERIAL PRIMARY KEY,
    name VARCHAR(255),
    picture VARCHAR(255));

CREATE TABLE IF NOT EXISTS room_has_users
    (room_id SERIAL NOT NULL REFERENCES rooms (id),
    user_id SERIAL NOT NULL REFERENCES chat_users (id),
    is_admin BOOLEAN DEFAULT false,
    INDEX (room_id, user_id));

/* Migration 1.11.2017 */

CREATE TABLE IF NOT EXISTS user_conversations
    (user_id SERIAL NOT NULL REFERENCES chat_users (id),
    receiver_user_id SERIAL NOT NULL REFERENCES chat_users (id),
    INDEX (user_id, receiver_user_id));

INSERT INTO rooms (name, picture) VALUES ('Welcome!', 'placeholder.jpg');
