# haloo-chat
Open source chat implementation

## Installing database 
Head to https://www.cockroachlabs.com and follow the instructions there.
### After installing
Drop the cockroach executable to /bin folder (the folder probably doesn't exist)
### Migrations
Add all migrations to the database/migration.sql file. The SQL should be valid SQL for CockroachDB. The chat application does not check for correctness of the SQL file.