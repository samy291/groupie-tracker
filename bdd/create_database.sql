-- Script de création de la base de données Groupie-Tracker

-- Table des utilisateurs
CREATE TABLE IF NOT EXISTS USER (
    id INTEGER PRIMARY KEY,
    pseudo TEXT NOT NULL,
    email TEXT NOT NULL,
    password TEXT NOT NULL
);

-- Table des salles de jeu
CREATE TABLE IF NOT EXISTS ROOMS (
    id INTEGER PRIMARY KEY,
    created_by INTEGER NOT NULL,
    max_player INTEGER NOT NULL,
    name TEXT NOT NULL,
    mode TEXT NOT NULL,
    id_game INTEGER,
    FOREIGN KEY (created_by) REFERENCES USER(id),
    FOREIGN KEY (id_game) REFERENCES GAMES(id)
);

-- Table des utilisateurs dans les salles de jeu avec leurs scores
CREATE TABLE IF NOT EXISTS ROOM_USERS (
    id_room INTEGER,
    id_user INTEGER,
    score INTEGER,
    FOREIGN KEY (id_room) REFERENCES ROOMS(id),
    FOREIGN KEY (id_user) REFERENCES USER(id),
    PRIMARY KEY (id_room, id_user)
);

-- Table des jeux
CREATE TABLE IF NOT EXISTS GAMES (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);
