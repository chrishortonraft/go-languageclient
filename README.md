# Go LSP Proxy

A Go-based HTTP server that acts as a proxy for interacting with the Pyright language server. This proxy server supports Language Server Protocol (LSP) methods such as hover and completion, allowing clients to interact with Pyright over HTTP.

## Features

* Interacts with the Pyright language server through the **Language Server Protocol** (LSP).
* Provides HTTP endpoints to process LSP requests like `hover`, `completion`, and others remaining to be implemented. The `client` package provides an easy interface to interact with other language servers as you desire.
* Built with Go and Docker, ensuring easy deployment and portability.

## Requirements

* **Go** 1.23 or higher
* **Docker** (for containerized usage)
* **Pyright** (Node.js package, automatically installed in the container)

## Installation

### 1. Clone the Repository

```bash
git clone https://github.com/chrishortonraft/go-languageclient.git
cd go-languageclient
```

### 2. Build the Docker Image

You can build the Docker image using the following command:

```bash
make docker
```

or for Mac M-series processors

```bash
make docker-arm
```

This will:

* Set up the Go environment
* Build your proxy server
* Expose port `8080`
* Install **Pyright** using **npm** inside the Docker container.

### 3. Run the Docker Container

To run the container, execute:

```bash
make run
```

or for Mac M-series processors

```bash
make run-arm
```

This will start the HTTP server, which will be accessible on port `8080` of your local machine.

## Usage

The HTTP server provides a simple API for sending LSP requests to Pyright.

### 1. Sending a `hover` Request

For a proof of concept - `hover` and `completion` methods are available to query.

These example requests will run on the following code, and request `hover` and `completion` from the Language Server.

```python
def main():
    print('Hello World')
```

#### Example Request

```json
{
  "method": "hover",
  "params": {
    "textDocument": {
      "uri": "file:///app/workspace/main.py"
    },
    "position": {
      "line": 0,
      "character": 17
    }
  }
}
```

You can use `curl` to make this request. Just copy and paste this command:

```bash
curl -X POST http://localhost:8080/api \
     -H "Content-Type: application/json" \
     -d '{
           "method": "hover",
           "params": {
             "textDocument": {
               "uri": "file:///app/workspace/main.py"
             },
             "position": {
               "line": 0,
               "character": 17
             }
           }
         }'
```

### 2. Sending a `completion` Request

Similarly, to request code completion, send a `POST` request to the `/api` endpoint with the following JSON body:

#### Example Request

```bash
curl -X POST http://localhost:8080/api \
     -H "Content-Type: application/json" \
     -d '{
           "method": "completion",
           "params": {
             "textDocument": {
               "uri": "file:///app/workspace/main.py"
             },
             "position": {
               "line": 0,
               "character": 17
             }
           }
         }'
```

### 3. Handling Responses

The server will respond with the result of the request. If successful, you will receive a JSON object containing the response from the Pyright language server.

#### Example Response for `hover`

```json
{
  "result": {
    "id": "3418dfe0-c46f-41ca-a615-f77dcac6adbc",
    "jsonrpc": "2.0",
    "result": {
      "contents": {
        "kind": "plaintext",
        "value": "(function) def print(\n    *values: object,\n    sep: str | None = \" \",\n    end: str | None = \"\\n\",\n    file: SupportsWrite[str] | None = None,\n    flush: Literal[False] = False\n) -> None"
      },
      "range": {
        "end": {
          "character": 6,
          "line": 1
        },
        "start": {
          "character": 1,
          "line": 1
        }
      }
    }
  }
}
```

* **Building from source**: You can also build the project locally without Docker, but using Docker is recommended for simplicity and portability.

## Contributing

Feel free to fork the repository and submit pull requests. You can open issues for any bugs, improvements, or features you would like to see.

## License

This project is licensed under the Apache License - see the [LICENSE](https://github.com/chrishortonraft/go-languageclient/blob/main/LICENSE) file for details.
