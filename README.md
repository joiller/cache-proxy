This is a submission for [the Caching Proxy in roadmap.sh](https://roadmap.sh/projects/caching-server)

## How to Run
Clone the repository and run the following command:
```bash
git clone https://github.com/joiller/cache-proxy.git
cd cache-proxy
```
Run the following command to start the server:
```bash
go build main.go
go run main.go --origin http://dummyjson.com --port 8080
```
In another terminal to run the following command to test the server:
```bash
# Test the server
curl http://localhost:8080/docs

# Clear the cache
go run main.go --clear-cache
```
