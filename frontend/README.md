# Hub Frontend

Modern React frontend for the Hub git hosting service built with Next.js 15, TypeScript, and Tailwind CSS.

## Features

- **Authentication** - Login/register forms with JWT token management
- **Dashboard** - Overview of repositories, activity, and statistics
- **Repository Management** - Browse, create, and manage repositories
- **Responsive Design** - Mobile-first design that works on all devices
- **Dark Mode Support** - Automatic theme switching based on system preferences
- **Component Library** - Custom UI components built with Headless UI
- **State Management** - Zustand for lightweight state management
- **Type Safety** - Full TypeScript support with strict typing

## Tech Stack

- **Framework**: Next.js 15 with App Router
- **Language**: TypeScript
- **Styling**: Tailwind CSS v4 with custom design tokens
- **UI Components**: Custom components with Headless UI
- **State Management**: Zustand
- **Forms**: React Hook Form
- **HTTP Client**: Axios with SWR for data fetching
- **Testing**: Jest + React Testing Library + Playwright E2E
- **Code Quality**: ESLint, Prettier, Husky

## Getting Started

1. **Install dependencies**:
   ```bash
   npm install
   ```

2. **Set up environment variables**:
   ```bash
   cp .env.example .env.local
   ```
   Edit `.env.local` with your API endpoint:
   ```
   NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
   ```

3. **Run the development server**:
   ```bash
   npm run dev
   ```

4. **Open your browser**:
   Visit [http://localhost:3000](http://localhost:3000)

## Available Scripts

- `npm run dev` - Start development server with Turbopack
- `npm run build` - Build for production
- `npm run start` - Start production server
- `npm run lint` - Run ESLint
- `npm run lint:fix` - Fix ESLint issues
- `npm run format` - Format code with Prettier
- `npm run format:check` - Check code formatting
- `npm run test` - Run unit tests
- `npm run test:watch` - Run tests in watch mode
- `npm run test:ci` - Run tests with coverage
- `npm run test:e2e` - Run end-to-end tests
- `npm run test:e2e:ui` - Run E2E tests with UI mode
- `npm run test:e2e:headed` - Run E2E tests in headed mode
- `npm run test:e2e:debug` - Debug E2E tests step by step
- `npm run test:e2e:report` - View E2E test reports
- `npm run type-check` - Check TypeScript types

## Project Structure

```
src/
├── app/                    # Next.js App Router pages
│   ├── dashboard/         # Dashboard page
│   ├── login/            # Login page
│   ├── register/         # Register page
│   └── repositories/     # Repository pages
├── components/           # React components
│   ├── ui/              # Base UI components
│   ├── layout/          # Layout components
│   └── forms/           # Form components
├── lib/                 # Utility libraries
│   ├── api.ts           # API client
│   └── utils.ts         # Helper functions
├── store/               # Zustand stores
│   ├── auth.ts          # Authentication store
│   └── app.ts           # Application store
├── types/               # TypeScript type definitions
└── styles/              # Global styles
```

## Authentication

The frontend includes a complete authentication system:

- **Login/Register forms** with validation
- **JWT token management** with automatic refresh
- **Protected routes** with authentication checks
- **User profile management**

## Components

### Base UI Components

- `Button` - Customizable button with variants and sizes
- `Input` - Form input with validation and error states
- `Card` - Content container with header, body, and footer
- `Modal` - Accessible modal dialogs
- `Dropdown` - Dropdown menus with keyboard navigation
- `Avatar` - User avatar with fallback initials
- `Badge` - Status and category badges

### Layout Components

- `AppLayout` - Main application layout with sidebar and header
- `Header` - Top navigation with search and user menu
- `Sidebar` - Side navigation with collapsible menu

## Testing

The project includes comprehensive testing with both unit and end-to-end tests:

### Unit Tests (Jest + React Testing Library)
Tests individual components and functions:

```bash
# Run unit tests
npm run test

# Run tests in watch mode
npm run test:watch

# Run tests with coverage
npm run test:ci
```

### End-to-End Tests (Playwright)
Tests complete user workflows across multiple browsers:

```bash
# Run all E2E tests
npm run test:e2e

# Run E2E tests with interactive UI
npm run test:e2e:ui

# Run E2E tests in visible browser mode
npm run test:e2e:headed

# Debug E2E tests step by step
npm run test:e2e:debug

# View E2E test reports
npm run test:e2e:report
```

E2E test coverage includes:
- **Authentication flows** - Login, register, logout
- **Dashboard functionality** - Repository lists, activity feeds, statistics
- **Navigation** - Header, sidebar, mobile menu interactions
- **Responsive design** - Mobile and desktop layouts
- **Cross-browser compatibility** - Chrome, Firefox, Safari

See `e2e/README.md` for detailed E2E testing documentation.

## Responsive Design

The application is built with a mobile-first approach and includes:

- **Responsive grid layouts** using CSS Grid and Flexbox
- **Mobile navigation** with collapsible sidebar
- **Touch-friendly interactions** for mobile devices
- **Optimized typography** that scales across screen sizes

## Contributing

1. **Code Style**: Follow the ESLint and Prettier configurations
2. **Components**: Create reusable components in the `ui/` directory
3. **Testing**: Add unit tests for new components
4. **Types**: Use TypeScript for all new code

## Deployment

The application can be deployed to any platform that supports Next.js:

### Vercel (Recommended)
```bash
npm run build
```
Deploy to Vercel using the Vercel CLI or GitHub integration.

### Docker
A Dockerfile is included for containerized deployments.

### Static Export
For static hosting, enable static export in `next.config.js`.

## Environment Variables

- `NEXT_PUBLIC_API_URL` - Backend API endpoint URL

## Browser Support

- Chrome (latest)
- Firefox (latest)
- Safari (latest)
- Edge (latest)