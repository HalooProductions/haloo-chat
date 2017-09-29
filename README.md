# haloo-chat
Open source chat implementation

## Installing database 
Head to https://www.cockroachlabs.com and follow the instructions there.
### After installing
Run the following commands in your terminal:  
`cockroach start --insecure --host=localhost`  
`cockroach sql --insecure` and input the following SQL:  
```sql
CREATE DATABASE haloochat;
GRANT ALL ON DATABASE haloochat TO maxroach;
```