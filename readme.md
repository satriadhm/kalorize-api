# Running Echo Framework Application in Golang on Port 8080

This README provides instructions on how to run an Echo framework application in Golang on port 8080.

## Prerequisites

Before running the application, make sure you have the following installed:

- Golang: [Download and install Golang](https://golang.org/dl/)
- Echo framework: Install Echo framework by running the following command:
  ```shell
  go get -u github.com/labstack/echo/v4
  ```

## Running the Application

Follow these steps to run the Echo framework application on port 8080:

1. Clone the repository or navigate to the project directory.

2. Build the application by running the following command:
    ```shell
    go build
    ```

3. Start the application by running the following command:
    ```shell
    ./<application_name>
    ```

    Replace `<application_name>` with the name of your application binary.

4. Open your web browser and navigate to `http://localhost:8080` to access the application.

## Configuration

By default, the application runs on port 8080. If you want to change the port, you can modify the code in your application's main file. Look for the following line:
