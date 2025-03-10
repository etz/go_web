# Mist (github.com/etz/go_web)

This is a Go web application that performs simple user authentication using Steam OpenID. This is a work in progress for a personal project (Mist) to track historical CS2 skin price data from online marketplaces. 

## Features

- Web server (basic routing, html templates)
- Simple user authentication using Steam OpenID


## Todo
- implement database connection
- build relevant api

## Getting Started

### Prerequisites

- Go 1.16 or higher
- Steam API Key (for authentication)

### Installation

1. Clone the repository:

```
bash
git clone https://github.com/yourusername/mist.git
cd mist
```

2. Create a `.env` file in the root directory and add your Steam API key:

```
STEAM_API_KEY=your_api_key_here
PORT=8080
```

3. Build and run the application:

```go run main.go```

4. Open your browser and navigate to Open your browser and navigate to `URL_ADDRESS:8080`.

## Project Structure
```go_web/
├── auth/             # Authentication package
├── components/       # HTML templates
├── static/           # Static assets
│   ├── css/          # Stylesheets
│   ├── images/       # Images and icons
│   └── js/           # JavaScript files
└── main.go           # Main application file, web router and handlers
```


