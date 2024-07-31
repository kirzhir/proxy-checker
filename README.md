## Proxy-Checker

Proxy-Checker is a powerful and versatile tool for checking the functionality of public proxy addresses in the format HOST:PORT
. It offers both command-line and web interface options, making it easy to use in various environments.


## Installation
To install Proxy-Checker, clone this repository and build the project:
  ```sh
git clone git@github.com:kirzhir/proxy-checker.git
cd proxy-checker
make build
  ```

## Usage
### Command-Line Interface (CLI)

You can use Proxy-Checker from the command line to check proxies. Here are some examples:

Read from standard input and output to a file:
* Read from standard input and output to a file:
  ```sh
  cat /path/to/proxies.txt | ./bin/cli -output=~/path/to/ok-proxies.txt
  ```
* Check proxies from a file:
  ```sh
  ./bin/cli -input=~/path/to/proxies.csv
  ```
* Default settings (input from stdin and output to stdout):
  ```sh
  ./bin/cli
  ```

### HTTP Server

#### API

* Proxy-Checker can be run as an HTTP server to support API requests. Here’s how to start the server:
  ```sh
  CONFIG_PATH=config/local.yaml ./bin/server
  ```
* Once the server is running, you can check proxies using a POST request:
  ```sh
  curl -X POST 'http://0.0.0.0:8082/api/v1/check' -d '["127.0.0.1:1234", "192.168.0.0:321"]'
  ```

#### Web Interface

Proxy-Checker also provides a web interface for checking proxies. Navigate to the running server's address in your web browser,  to see the form


<!-- LICENSE -->
## License

Distributed under the MIT License. See `LICENSE` for more information.

<p align="right">(<a href="#readme-top">back to top</a>)</p>