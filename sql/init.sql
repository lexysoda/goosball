DROP TABLE IF EXISTS users;
CREATE TABLE users (
  ID STRING PRIMARY KEY NOT NULL,
  DisplayName STRING NOT NULL,
  RealName STRING DEFAULT '',
  Avatar STRING DEFAULT '',
  Mu FLOAT NOT NULL,
  SigmaSq FLOAT NOT NULL
);

DROP TABLE IF EXISTS sets;
CREATE TABLE sets (
  ID INTEGER PRIMARY KEY AUTOINCREMENT,
  P1 STRING NOT NULL,
  P2 STRING NOT NULL,
  P3 STRING NOT NULL,
  P4 STRING NOT NULL,
  FOREIGN KEY(P1) REFERENCES users(ID),
  FOREIGN KEY(P2) REFERENCES users(ID),
  FOREIGN KEY(P3) REFERENCES users(ID),
  FOREIGN KEY(P4) REFERENCES users(ID)
);

DROP TABLE IF EXISTS games;
CREATE TABLE games (
  ID INTEGER PRIMARY KEY AUTOINCREMENT,
  SetID INTEGER NOT NULL,
  GoalsA INTEGER NOT NULL,
  GoalsB INTEGER NOT NULL,
  StartTime TIMESTAMP NOT NULL,
  EndTime TIMESTAMP NOT NULL,
  FOREIGN KEY(SetID) REFERENCES sets(ID)
);