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
  cat /path/to/proxies.txt | ./proxy-checker cli -o=~/path/to/ok-proxies.txt
  ```
* Check proxies from a file:
  ```sh
  ./proxy-checker cli -i=~/path/to/proxies.csv
  ```
* Default settings (input from stdin and output to stdout):
  ```sh
  ./proxy-checker cli
  ```

### HTTP Server

#### API

* Proxy-Checker can be run as an HTTP server to support API requests. Here’s how to start the server:
  ```sh
  CONFIG_PATH=config/local.yaml ./proxy-checker serve
  ```
* Once the server is running, you can check proxies using a POST request:
  ```sh
  curl -X POST 'http://0.0.0.0:8082/api/v1/check' -d '["127.0.0.1:1234", "192.168.0.0:321"]'
  ```

#### Web Interface

Proxy-Checker also provides a web interface for checking proxies. Navigate to the running server's address in your web browser,  to see the form 

<img width="200" alt="Снимок экрана 2024-08-01 в 14 21 48" src="https://github.com/user-attachments/assets/4bc778f8-8f76-40d0-a1d9-587e795f07d7">


## Configuration

#### HTTP Server

The HTTP server reads its configuration from a file specified by the CONFIG_PATH environment variable. You need to set this variable to the path of your configuration file before running the server. The configuration file should be in YAML format.
* Example
  ```sh
  CONFIG_PATH=config/local.yaml ./proxy-checker serve
  ```
* Structure
  ```yaml
  env: "local"
  http_server:
    address: "0.0.0.0:8082"
    timeout: 4s
    idle_timeout: 30s
  proxy_checker:
    api: "http://checkip.amazonaws.com"
    timeout: 1800ms
    concurrency: 100
  ```

#### CLI

The CLI utility reads all its configuration from environment variables. Each configuration parameter should be set as an environment variable before running the CLI utility.
* Example
  ```sh
  API=https://api64.ipify.org TIMEOUT=1s CONCURRENCY=200 ./proxy-checker cli -v
  ```

<!-- LICENSE -->
## License

Distributed under the MIT License. See `LICENSE` for more information.

<p align="right">(<a href="#readme-top">back to top</a>)</p>
