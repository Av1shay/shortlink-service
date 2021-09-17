## Shortlink-service assignment
## Run local

`.env` file
````
PORT=8080
SHORTLINK_BASE_URL="http://localhost:8080"
MONGO_URI="mongodb://username:password@localhost:26000"
MONGO_DB=db
MONGO_COLLECTION=shortlinks
````
`docker-compose up -d`

`go run main.go`

POST http://localhost:8080/s/generate \
JSON Body:
```json
{
  "keyType": "standard",
  "redirects": [
    {
      "from": 0,
      "to": 12,
      "url": "http://google.com"
    },
    {
      "from": 12,
      "to": 24,
      "url": "https://youtube.com"
    }
  ]
}
```
Response: `http://localhost:8080/e`

Body for UUID key:
```json
{
  "keyType": "uuid",
  "redirects": [
    {
      "from": 0,
      "to": 12,
      "url": "http://google.com"
    }
  ]
}
```
Response: `http://localhost:8080/u/8b821463-3c68-4832-47e2-39d905c6d84a`
## Test
`make test`