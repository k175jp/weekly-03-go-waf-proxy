# weekly-03-go-waf-proxy

A lightweight, defensive Reverse Proxy and Minimal WAF (Web Application Firewall) environment built with Go and Docker Compose.

This project was created as part of my weekly development series.

## Features
- **Reverse Proxy Core**: Forwards incoming client traffic to internal upstream services using Go's standard `httputil.ReverseProxy`.
- **Docker-native Healthchecking**: Ensures zero downtime/errors during startup by orchestrating container dependency using `depends_on` with a `service_healthy` condition.
- **Vulnerable Sandbox Environment**: Packaged with an intentional Reflected XSS vulnerable application for immediate, safe payload testing.
- **Request Validator (In Development)**: Inspects query parameters and request bodies against custom security signatures to return `403 Forbidden` for malicious requests without altering the original payload data.

## Tech Stack
- Go (Golang)
- Docker
- Docker Compose

## Getting Started
### Run with Docker Compose
Ensure you have Docker and Docker Compose installed.

docker compose up --build

### Endpoints
- **WAF Protected Proxy Route**: http://localhost:9000/search
- **Direct Backend Route (Bypass)**: http://localhost:8080/search

### How to Test (Vulnerability Lab)
1. **Access without WAF**: Open `http://localhost:8080/search?q=<script>alert(1)</script>` -> The payload executes successfully (Vulnerability confirmed).
2. **Access through WAF**: Open `http://localhost:9000/search?q=<script>alert(1)</script>` -> *(In Development)* The WAF interceptor detects the malicious signature and drops the connection with a `403` status.

## Purpose
The goal of this project was to practice:
- Network programming and proxy implementations using Go's `net/http` stack.
- Advanced container orchestration using Docker Compose healthchecks (`wget` spidering on Alpine).
- Web application security fundamentals from both the offensive (vulnerability creation) and defensive (WAF/filtering) perspectives.
- Understanding the boundary of responsibilities between infrastructure-level inspection (WAF) and context-aware validation.

## Notes
This is a small experimental project focused on learning and implementation rather than production readiness.

The project serves as an interactive security sandbox environment for engineers to understand how web application attacks are analyzed and dropped at the infrastructure edge.