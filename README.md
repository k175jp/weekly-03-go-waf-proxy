# weekly-03-go-waf-proxy

A lightweight, defensive Reverse Proxy and Minimal WAF (Web Application Firewall) environment built with Go and Docker Compose.

This project was created as part of my weekly development series.

## Features
 - Reverse Proxy Core
 - Docker-native Healthchecking
 - Vulnerable Sandbox Environment
 - Request Validator (In Development)

## Tech Stack
 - Go
 - Docker
 - Docker Compose

## Getting Started
### Run with Docker Compose
```bash
docker compose up --build
```

### Endpoints
- **WAF Protected Proxy Route**: http://localhost:9000/search
- **Direct Backend Route (Bypass)**: http://localhost:8080/search

### How to Test (Vulnerability Lab)
1. **Access without WAF**: Open `http://localhost:8080/search?q=<script>alert(1)</script>` -> The payload executes successfully (Vulnerability confirmed).
2. **Access through WAF**: Open `http://localhost:9000/search?q=<script>alert(1)</script>` -> *(In Development)* The WAF interceptor detects the malicious signature and drops the connection with a `403` status.

## Purpose
The goal of this project was to practice:
 - Network programming and proxy implementations using Go's `net/http` stack.
 - Advanced container orchestration using Docker Compose healthchecks.
 - Web application security fundamentals from both the offensive (vulnerability creation) and defensive (WAF/filtering) perspectives.
 - Understanding the boundary of responsibilities between infrastructure-level inspection (WAF) and context-aware validation.

## Notes
This is a small experimental project focused on learning and implementation rather than production readiness.
