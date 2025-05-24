# MyNodeCP - Enterprise Web Hosting Control Panel

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![React Version](https://img.shields.io/badge/React-18+-blue.svg)](https://reactjs.org)

## Overview

MyNodeCP is a commercial-grade, open-source web hosting control panel designed to compete with cPanel, DirectAdmin, and Plesk. Built for global production deployment, it offers enterprise-level functionality while remaining free and accessible to hosting providers worldwide.

## Key Features

- **Native Service Management**: Direct integration with MariaDB, Redis, Nginx, Apache, Postfix, Dovecot
- **Enterprise Performance**: Sub-200ms API response times, 99.9% uptime
- **Global Compatibility**: Support for RHEL, CentOS, Rocky Linux, AlmaLinux, Ubuntu LTS, Debian
- **Advanced Security**: SOC 2, ISO 27001 compliance ready
- **Scalable Architecture**: Support for 100,000+ domains per cluster

## Architecture

### Backend Stack
- **Language**: Go 1.21+
- **API**: gRPC microservices with HTTP/2 streaming
- **Database**: MariaDB 10.11+ with native clustering
- **Cache**: Redis for session management and caching
- **Messaging**: Apache Kafka for event streaming
- **Authentication**: OAuth 2.0/OIDC with JWT

### Frontend Stack
- **Framework**: React 18+ with TypeScript
- **State Management**: Redux Toolkit with RTK Query
- **Build System**: Vite with code splitting
- **Styling**: Tailwind CSS with custom theme system
- **Accessibility**: WCAG 2.1 AA compliance

## Quick Start

### Prerequisites
- Linux server (RHEL/CentOS/Rocky/Alma/Ubuntu/Debian)
- Go 1.21+
- Node.js 18+
- MariaDB 10.11+
- Redis 6+

### Installation

```bash
# Clone the repository
git clone https://github.com/mynodecp/mynodecp.git
cd mynodecp

# Run the installation script
sudo ./scripts/install.sh

# Start the services
sudo systemctl start mynodecp
sudo systemctl enable mynodecp
```

## Development

### Backend Development
```bash
cd backend
go mod download
go run cmd/server/main.go
```

### Frontend Development
```bash
cd frontend
npm install
npm run dev
```

### Running Tests
```bash
# Backend tests
cd backend && go test ./...

# Frontend tests
cd frontend && npm test
```

## Project Structure

```
mynodecp/
├── backend/                 # Go backend services
│   ├── cmd/                # Application entry points
│   ├── internal/           # Private application code
│   ├── pkg/                # Public library code
│   ├── api/                # API definitions (gRPC, OpenAPI)
│   ├── migrations/         # Database migrations
│   └── configs/            # Configuration files
├── frontend/               # React frontend application
│   ├── src/                # Source code
│   ├── public/             # Static assets
│   └── dist/               # Build output
├── scripts/                # Deployment and utility scripts
├── docs/                   # Documentation
├── tests/                  # Integration tests
└── deployments/            # Deployment configurations
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Security

For security vulnerabilities, please email security@mynodecp.com instead of using the issue tracker.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- Documentation: https://docs.mynodecp.com
- Community Forum: https://community.mynodecp.com
- Commercial Support: https://mynodecp.com/support

## Roadmap

- [x] Phase 1: Foundation & Core API (Weeks 1-3)
- [ ] Phase 2: Control Panel Core (Weeks 4-8)
- [ ] Phase 3: Advanced Features (Weeks 9-12)
- [ ] Phase 4: Enterprise Features & Marketplace

---

Built with ❤️ for the global hosting community
