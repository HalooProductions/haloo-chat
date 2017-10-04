CREATE DATABASE IF NOT EXISTS haloochat;

CREATE TABLE IF NOT EXISTS users 
    (id SERIAL PRIMARY KEY, 
    name VARCHAR(255), 
    email VARCHAR(255), 
    password VARCHAR(255), 
    last_seen TIMESTAMPTZ, 
    profile_picture VARCHAR(255));
    
CREATE TABLE IF NOT EXISTS chatlog 
    (id SERIAL PRIMARY KEY, 
    sender integer, 
    message TEXT, 
    timestamp TIMESTAMPTZ);
