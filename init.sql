-- Initial database setup for local development
CREATE USER hyttekos WITH ENCRYPTED PASSWORD 'hyttekos';
CREATE DATABASE hyttekos OWNER hyttekos;
GRANT ALL PRIVILEGES ON DATABASE hyttekos TO hyttekos;