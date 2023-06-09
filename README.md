# Configurable Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

The Configurable package provides a simple and flexible way to handle configuration data in your Go projects. It allows you to define and manage various types of configuration variables, such as integers, strings, booleans, durations, and more. This package supports configuration parsing from JSON, YAML, and INI files, as well as environment variables.

## Installation

To use the Configurable package in your project, you need to have Go installed and set up. Then, you can install the package by running the following command in your terminal:

```shell
go get -u github.com/andreimerlescu/configurable
```

## Usage

To use the Configurable package in your Go code, you need to import it:

```go
import "github.com/andreimerlescu/configurable"
```

### Creating a Configurable Instance

To get started, you need to create an instance of the Configurable struct by calling the `NewConfigurable()` function:

```go
config := configurable.NewConfigurable()
```

### Defining Configuration Variables

The Configurable package provides several methods to define different types of configuration variables. Each method takes a name, default value, and usage description as parameters and returns a pointer to the respective variable:

```go
port := config.NewInt("port", 8080, "The port number to listen on")
timeout := config.NewDuration("timeout", time.Second * 5, "The timeout duration for requests")
debug := config.NewBool("debug", false, "Enable debug mode")
```

### Loading Configuration from Files

You can load configuration data from JSON, YAML, and INI files using the `LoadFile()` method:

```go
err := config.LoadFile("config.json")
if err != nil {
    // Handle error
}
```

The package automatically parses the file based on its extension. Make sure to place the file in the correct format in the specified location.

### Parsing Command-Line Arguments

The Configurable package also allows you to parse command-line arguments. Call the `Parse()` method to parse the arguments after defining your configuration variables:

```go
err := config.Parse("")
if err != nil {
    // Handle error
}
```

Passing an empty string to `Parse()` means it will only parse the command-line arguments and not load any file.

### Accessing Configuration Values

You can access the values of your configuration variables using the respective getter methods:

```go
fmt.Println("Port:", *port)
fmt.Println("Timeout:", *timeout)
fmt.Println("Debug mode:", *debug)
```

### Environment Variables

The Configurable package supports setting configuration values through environment variables. If an environment variable with the same name as a configuration variable exists, the package will automatically assign its value to the respective variable. Ensure that the environment variables are in uppercase and match the configuration variable names.

### Displaying Usage Information

To generate a usage string with information about your configuration variables, use the `Usage()` method:

```go
usage := config.Usage()
fmt.Println(usage)
```

The generated usage string includes information about each configuration variable, including its name, default value, description, and the source from which it was set (flag, environment, JSON, YAML, or INI).

## License

This package is distributed under the MIT License. See the [LICENSE](LICENSE) file for more information.

## Contributing

Contributions to this package are welcome. If you find any issues or have suggestions for improvements, please open an issue or submit a pull request.

Enjoy using the Configurable package in your projects!