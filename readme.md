# TypeSense Playground

1. Clone the project

`git clone https://github.com/26kito/typesense-dump.git`

2. Setup

`go mod download`

3. Setup and start typesense in Docker

`docker-compose up -d`

4. Run service

`go run main.go`

5. Create a new collection in local

Http Method POST | Endpoint: http://localhost:3000/create-collection

6. Insert dummy data from JSON file

Http Method POST | Endpoint: http://localhost:3000/add-document
 
7. Find document from collection

Http Method GET | http://localhost:3000/search?q={keywords}