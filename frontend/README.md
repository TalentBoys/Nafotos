# AwesomeSharing Frontend

A modern file sharing platform frontend built with React, TypeScript, and Vite.

[中文文档](./README-ZH_CN.md)

## Features

- File and folder management with upload/download capabilities
- Album and timeline views for media files
- User authentication and authorization
- Multi-language support (i18n)
- Permission group management
- Domain configuration for multi-tenant support
- Responsive design with Tailwind CSS

## Tech Stack

- **React 18** - UI framework
- **TypeScript** - Type-safe development
- **Vite** - Fast build tool and dev server
- **React Router** - Client-side routing
- **Axios** - HTTP client
- **i18next** - Internationalization
- **Tailwind CSS** - Utility-first CSS framework
- **Lucide React** - Icon library

## Prerequisites

- Node.js >= 16.0.0
- npm or yarn package manager

## Getting Started

### Installation

```bash
# Install dependencies
npm install
```

### Development

```bash
# Start development server
npm run dev
```

The development server includes:
- Hot Module Replacement (HMR)
- API proxy to backend
- Configurable ports via `.env.local` (see Configuration section)

### Build

```bash
# Build for production
npm run build
```

The production build will be output to the `dist` directory.

### Preview Production Build

```bash
# Preview production build locally
npm run preview
```

### Linting

```bash
# Run ESLint
npm run lint
```

## Project Structure

```
frontend/
├── public/              # Static assets
├── src/
│   ├── components/      # Reusable React components
│   │   ├── FileGrid.tsx
│   │   ├── Layout.tsx
│   │   ├── Modal.tsx
│   │   └── ProtectedRoute.tsx
│   ├── contexts/        # React contexts (Auth, etc.)
│   ├── hooks/           # Custom React hooks
│   ├── locales/         # i18n translation files
│   ├── pages/           # Page components
│   │   ├── Albums.tsx
│   │   ├── DomainConfig.tsx
│   │   ├── FileDetail.tsx
│   │   ├── FolderManagement.tsx
│   │   ├── Folders.tsx
│   │   ├── Login.tsx
│   │   ├── PermissionGroupManagement.tsx
│   │   ├── Settings.tsx
│   │   ├── Timeline.tsx
│   │   └── UserManagement.tsx
│   ├── services/        # API service modules
│   ├── types/           # TypeScript type definitions
│   ├── utils/           # Utility functions
│   ├── App.tsx          # Main app component
│   ├── main.tsx         # Application entry point
│   └── i18n.ts          # i18n configuration
├── index.html           # HTML template
├── vite.config.ts       # Vite configuration
├── tailwind.config.js   # Tailwind CSS configuration
├── tsconfig.json        # TypeScript configuration
└── package.json         # Project dependencies

```

## Configuration

### Configuring Development Ports

For local development, create a `.env.local` file in the **project root** (not in the frontend directory):

```bash
# .env.local (in project root)
BACKEND_PORT=8080
FRONTEND_PORT=3000
```

The Vite configuration will automatically read these values and:
- Start the frontend dev server on `FRONTEND_PORT` (default: 3000)
- Configure the API proxy to target `http://localhost:BACKEND_PORT` (default: 8080)

This allows different developers to use different ports without modifying configuration files.

### Environment Variables

The application uses Vite's environment variable system. For custom environment variables, create a `.env` file:

```env
VITE_API_URL=http://localhost:8080
```

### API Proxy

The development server is configured to proxy API requests to the backend:

- Frontend: `http://localhost:FRONTEND_PORT` (default: 3000)
- Backend API: `http://localhost:BACKEND_PORT` (default: 8080)
- Proxy: `/api/*` → `http://localhost:BACKEND_PORT/api/*`

Port values are read from `.env.local` in the project root.

## Available Scripts

- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm run preview` - Preview production build
- `npm run lint` - Run ESLint

## License

MIT
